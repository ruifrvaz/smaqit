package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// The Infrastructure spec's `## Environments` table declares each
// existing-k3s environment's target: where it runs, how CI reaches it, and who
// owns its Ingress. It is the single source the k3s fillers and the workflow
// renderer read, through `smaqit spec environments --json`.

const environmentsHeading = "## Environments"

var (
	reachabilityClasses = []string{"public", "private-mesh"}
	ingressModes        = []string{"own-tls", "platform-default-tls", "platform-owned"}
	probeClients        = []string{"curl", "wget"}

	envSlugPattern       = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
	hostPattern          = regexp.MustCompile(`^[a-z0-9]([a-z0-9.-]*[a-z0-9])?$`)
	secretNamePattern    = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)
	actionRefPattern     = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_./-]+@[A-Za-z0-9_./-]+$`)
	actionInputPattern   = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*$`)
	healthPathPattern    = regexp.MustCompile(`^/[A-Za-z0-9._~/-]*$`)
	environmentNameRegex = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)
)

// environmentColumns maps each required table header to its JSON field.
var environmentColumns = []string{
	"Environment", "Machine Slug", "Host", "Reachability", "Ingress Mode",
	"Ingress Class", "ClusterIssuer", "Health Path", "Mesh Setup-Key Secret", "Mesh Wait-For Host",
}

type infraEnvironment struct {
	Environment        string `json:"environment"`
	MachineSlug        string `json:"machineSlug"`
	Host               string `json:"host"`
	Reachability       string `json:"reachability"`
	IngressMode        string `json:"ingressMode"`
	IngressClass       string `json:"ingressClass"`
	ClusterIssuer      string `json:"clusterIssuer"`
	HealthPath         string `json:"healthPath"`
	MeshSetupKeySecret string `json:"meshSetupKeySecret"`
	MeshWaitForHost    string `json:"meshWaitForHost"`
}

// verifiesInCluster reports whether CI cannot reach the app from the public
// internet, so verification must run inside the cluster.
func (e infraEnvironment) verifiesInCluster() bool {
	return e.Reachability == "private-mesh" || e.IngressMode == "platform-owned"
}

type meshAccess struct {
	Action        string `json:"action"`
	SetupKeyInput string `json:"setupKeyInput"`
	WaitForInput  string `json:"waitForInput"`
}

type infraEnvironments struct {
	SchemaVersion int                `json:"schemaVersion"`
	Spec          string             `json:"spec"`
	Environments  []infraEnvironment `json:"environments"`
	Mesh          *meshAccess        `json:"mesh"`
	ProbeClient   string             `json:"probeClient"`
}

func (set *infraEnvironments) find(name string) *infraEnvironment {
	for i := range set.Environments {
		if set.Environments[i].Environment == name {
			return &set.Environments[i]
		}
	}
	return nil
}

// markdownSection returns the body of a `## ` section, or false when absent.
func markdownSection(text, heading string) (string, bool) {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != heading {
			continue
		}
		var body []string
		for _, next := range lines[i+1:] {
			if strings.HasPrefix(next, "## ") {
				break
			}
			body = append(body, next)
		}
		return strings.Join(body, "\n"), true
	}
	return "", false
}

// markdownTable returns the rows of the first pipe table in a section,
// header first, separator row dropped.
func markdownTable(section string) [][]string {
	var rows [][]string
	for _, line := range strings.Split(section, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			if len(rows) > 0 {
				break
			}
			continue
		}
		line = strings.TrimSuffix(strings.TrimPrefix(line, "|"), "|")
		cells := strings.Split(line, "|")
		for i := range cells {
			cells[i] = strings.TrimSpace(cells[i])
		}
		if isSeparatorRow(cells) {
			continue
		}
		rows = append(rows, cells)
	}
	return rows
}

func isSeparatorRow(cells []string) bool {
	for _, cell := range cells {
		if strings.Trim(cell, ":- ") != "" {
			return false
		}
	}
	return true
}

// cellValue normalizes a table cell: code spans are unwrapped, and empty
// markers or unfilled template placeholders count as no value.
func cellValue(cell string) string {
	value := strings.TrimSpace(cell)
	if len(value) >= 2 && strings.HasPrefix(value, "`") && strings.HasSuffix(value, "`") {
		value = strings.TrimSpace(value[1 : len(value)-1])
	}
	switch strings.ToLower(value) {
	case "", "-", "—", "–", "n/a", "none":
		return ""
	}
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		return ""
	}
	return value
}

// constraintValues reads the `## Constraints` table as constraint → value.
func constraintValues(text string) map[string]string {
	values := map[string]string{}
	section, ok := markdownSection(text, "## Constraints")
	if !ok {
		return values
	}
	rows := markdownTable(section)
	for _, row := range rows[min(1, len(rows)):] {
		if len(row) >= 2 {
			values[row[0]] = cellValue(row[1])
		}
	}
	return values
}

// loadInfraEnvironments finds the one Infrastructure spec that declares an
// Environments table and parses it. It returns nil, nil when no spec declares one.
func loadInfraEnvironments(root string) (*infraEnvironments, []error) {
	dir := filepath.Join(root, "specs", "infrastructure")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, []error{fmt.Errorf("SLOTS-SPEC-MISSING: reading %s: %v", dir, err)}
	}
	var found []string
	var text string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, []error{fmt.Errorf("SLOTS-SPEC-MISSING: reading %s: %v", path, err)}
		}
		if _, ok := markdownSection(string(content), environmentsHeading); ok {
			found = append(found, filepath.ToSlash(filepath.Join("specs", "infrastructure", entry.Name())))
			text = string(content)
		}
	}
	switch len(found) {
	case 0:
		return nil, nil
	case 1:
	default:
		sort.Strings(found)
		return nil, []error{fmt.Errorf("SLOTS-SPEC-AMBIGUOUS: more than one Infrastructure spec declares %q (%s); declare every environment in one spec", environmentsHeading, strings.Join(found, ", "))}
	}
	set, errs := parseInfraEnvironments(text)
	if set != nil {
		set.Spec = found[0]
	}
	return set, errs
}

// parseInfraEnvironments parses the table and the constraint rows it depends
// on. Structural problems are returned; field rules are checked separately.
func parseInfraEnvironments(text string) (*infraEnvironments, []error) {
	section, _ := markdownSection(text, environmentsHeading)
	rows := markdownTable(section)
	if len(rows) == 0 {
		return nil, []error{fmt.Errorf("SLOTS-SPEC-MISSING: %q has no table", environmentsHeading)}
	}
	index := map[string]int{}
	for i, header := range rows[0] {
		index[header] = i
	}
	var errs []error
	for _, column := range environmentColumns {
		if _, ok := index[column]; !ok {
			errs = append(errs, fmt.Errorf("SLOTS-SPEC-FIELD-MISSING: %q table has no %q column", environmentsHeading, column))
		}
	}
	if len(errs) > 0 {
		return nil, errs
	}
	set := &infraEnvironments{SchemaVersion: 1, Environments: []infraEnvironment{}}
	seen := map[string]bool{}
	for _, row := range rows[1:] {
		get := func(column string) string {
			if i := index[column]; i < len(row) {
				return cellValue(row[i])
			}
			return ""
		}
		env := infraEnvironment{
			Environment:        get("Environment"),
			MachineSlug:        get("Machine Slug"),
			Host:               get("Host"),
			Reachability:       get("Reachability"),
			IngressMode:        get("Ingress Mode"),
			IngressClass:       get("Ingress Class"),
			ClusterIssuer:      get("ClusterIssuer"),
			HealthPath:         get("Health Path"),
			MeshSetupKeySecret: get("Mesh Setup-Key Secret"),
			MeshWaitForHost:    get("Mesh Wait-For Host"),
		}
		if env.Environment == "" {
			errs = append(errs, fmt.Errorf("SLOTS-SPEC-FIELD-MISSING: an %q row has no Environment", environmentsHeading))
			continue
		}
		if seen[env.Environment] {
			errs = append(errs, fmt.Errorf("SLOTS-SPEC-FIELD-INVALID: environment %q is declared more than once", env.Environment))
			continue
		}
		seen[env.Environment] = true
		set.Environments = append(set.Environments, env)
	}
	constraints := constraintValues(text)
	if action := constraints["Mesh Join Action"]; action != "" || constraints["Mesh Setup-Key Input"] != "" || constraints["Mesh Wait-For Input"] != "" {
		set.Mesh = &meshAccess{Action: action, SetupKeyInput: constraints["Mesh Setup-Key Input"], WaitForInput: constraints["Mesh Wait-For Input"]}
	}
	set.ProbeClient = constraints["Probe Client"]
	return set, errs
}

// validateEnvironment checks one row's fields against its class and mode.
func validateEnvironment(env infraEnvironment) []error {
	var errs []error
	at := func(field string) string { return fmt.Sprintf("environment %q %s", env.Environment, field) }
	missing := func(field, why string) {
		errs = append(errs, fmt.Errorf("SLOTS-SPEC-FIELD-MISSING: %s is required%s", at(field), why))
	}
	invalid := func(field, why string) {
		errs = append(errs, fmt.Errorf("SLOTS-SPEC-FIELD-INVALID: %s %s", at(field), why))
	}
	if !environmentNameRegex.MatchString(env.Environment) {
		invalid("name", "must be letters, digits, '-' or '_'")
	}
	switch {
	case env.MachineSlug == "":
		missing("Machine Slug", "")
	case !envSlugPattern.MatchString(env.MachineSlug):
		invalid("Machine Slug", fmt.Sprintf("%q must be lowercase letters, digits, and inner hyphens", env.MachineSlug))
	}
	switch {
	case env.Reachability == "":
		missing("Reachability", " (one of "+strings.Join(reachabilityClasses, ", ")+")")
	case !containsString(reachabilityClasses, env.Reachability):
		invalid("Reachability", fmt.Sprintf("%q must be one of %s", env.Reachability, strings.Join(reachabilityClasses, ", ")))
	}
	switch {
	case env.IngressMode == "":
		missing("Ingress Mode", " (one of "+strings.Join(ingressModes, ", ")+")")
	case !containsString(ingressModes, env.IngressMode):
		invalid("Ingress Mode", fmt.Sprintf("%q must be one of %s", env.IngressMode, strings.Join(ingressModes, ", ")))
	}
	ownsIngress := env.IngressMode != "" && env.IngressMode != "platform-owned"
	switch {
	case env.Host == "" && ownsIngress:
		missing("Host", " unless Ingress Mode is platform-owned")
	case env.Host != "" && !hostPattern.MatchString(env.Host):
		invalid("Host", fmt.Sprintf("%q must be a lowercase hostname", env.Host))
	}
	switch {
	case env.IngressClass == "" && ownsIngress:
		missing("Ingress Class", " unless Ingress Mode is platform-owned")
	case env.IngressClass != "" && env.IngressMode == "platform-owned":
		invalid("Ingress Class", "must be empty when Ingress Mode is platform-owned (no Ingress is applied)")
	}
	switch {
	case env.ClusterIssuer == "" && env.IngressMode == "own-tls":
		missing("ClusterIssuer", " when Ingress Mode is own-tls")
	case env.ClusterIssuer != "" && env.IngressMode != "" && env.IngressMode != "own-tls":
		invalid("ClusterIssuer", "must be empty unless Ingress Mode is own-tls (the platform serves TLS)")
	}
	switch {
	case env.HealthPath == "":
		missing("Health Path", "")
	case !healthPathPattern.MatchString(env.HealthPath):
		invalid("Health Path", fmt.Sprintf("%q must be an absolute path such as /health.json", env.HealthPath))
	}
	private := env.Reachability == "private-mesh"
	switch {
	case env.MeshSetupKeySecret == "" && private:
		missing("Mesh Setup-Key Secret", " when Reachability is private-mesh")
	case env.MeshSetupKeySecret != "" && !private:
		invalid("Mesh Setup-Key Secret", "must be empty unless Reachability is private-mesh")
	case env.MeshSetupKeySecret != "" && (!secretNamePattern.MatchString(env.MeshSetupKeySecret) || strings.HasPrefix(env.MeshSetupKeySecret, "GITHUB_")):
		invalid("Mesh Setup-Key Secret", fmt.Sprintf("%q must be an uppercase GitHub secret name (A-Z, 0-9, _), not starting with GITHUB_", env.MeshSetupKeySecret))
	}
	switch {
	case env.MeshWaitForHost == "" && private:
		missing("Mesh Wait-For Host", " when Reachability is private-mesh")
	case env.MeshWaitForHost != "" && !private:
		invalid("Mesh Wait-For Host", "must be empty unless Reachability is private-mesh")
	case env.MeshWaitForHost != "" && !hostPattern.MatchString(env.MeshWaitForHost):
		invalid("Mesh Wait-For Host", fmt.Sprintf("%q must be a lowercase hostname", env.MeshWaitForHost))
	}
	return errs
}

// validateSharedConstraints checks the Constraints rows the given rows depend on.
func validateSharedConstraints(set *infraEnvironments, rows []infraEnvironment) []error {
	var errs []error
	needMesh, needProbe := false, false
	for _, env := range rows {
		needMesh = needMesh || env.Reachability == "private-mesh"
		needProbe = needProbe || env.verifiesInCluster()
	}
	mesh := set.Mesh
	if mesh == nil {
		mesh = &meshAccess{}
	}
	if needMesh {
		for _, field := range []struct{ name, value string }{
			{"Mesh Join Action", mesh.Action},
			{"Mesh Setup-Key Input", mesh.SetupKeyInput},
			{"Mesh Wait-For Input", mesh.WaitForInput},
		} {
			if field.value == "" {
				errs = append(errs, fmt.Errorf("SLOTS-SPEC-FIELD-MISSING: Constraints row %q is required when any environment is private-mesh", field.name))
			}
		}
	}
	if mesh.Action != "" && !actionRefPattern.MatchString(mesh.Action) {
		errs = append(errs, fmt.Errorf("SLOTS-SPEC-FIELD-INVALID: Constraints row \"Mesh Join Action\" %q must be a pinned uses: reference (owner/repo[/path]@ref)", mesh.Action))
	}
	for _, field := range []struct{ name, value string }{
		{"Mesh Setup-Key Input", mesh.SetupKeyInput},
		{"Mesh Wait-For Input", mesh.WaitForInput},
	} {
		if field.value != "" && !actionInputPattern.MatchString(field.value) {
			errs = append(errs, fmt.Errorf("SLOTS-SPEC-FIELD-INVALID: Constraints row %q %q must be an action input name", field.name, field.value))
		}
	}
	switch {
	case set.ProbeClient == "" && needProbe:
		errs = append(errs, fmt.Errorf("SLOTS-SPEC-FIELD-MISSING: Constraints row \"Probe Client\" (%s) is required when any environment verifies in-cluster (private-mesh or platform-owned)", strings.Join(probeClients, " or ")))
	case set.ProbeClient != "" && !containsString(probeClients, set.ProbeClient):
		errs = append(errs, fmt.Errorf("SLOTS-SPEC-FIELD-INVALID: Constraints row \"Probe Client\" %q must be one of %s", set.ProbeClient, strings.Join(probeClients, ", ")))
	}
	return errs
}

// validateInfraEnvironments checks every row, or only the named environment
// when one is given. A missing table or row is an error.
func validateInfraEnvironments(set *infraEnvironments, environment string) []error {
	if set == nil {
		return []error{fmt.Errorf("SLOTS-SPEC-MISSING: no Infrastructure spec under specs/infrastructure/ declares an %q table; existing-k3s needs one row per environment", environmentsHeading)}
	}
	rows := set.Environments
	if environment != "" {
		env := set.find(environment)
		if env == nil {
			return []error{fmt.Errorf("SLOTS-SPEC-MISSING: %s %q table has no row for environment %q", set.Spec, environmentsHeading, environment)}
		}
		rows = []infraEnvironment{*env}
	} else if len(rows) == 0 {
		return []error{fmt.Errorf("SLOTS-SPEC-MISSING: %s %q table has no rows", set.Spec, environmentsHeading)}
	}
	var errs []error
	for _, env := range rows {
		errs = append(errs, validateEnvironment(env)...)
	}
	return append(errs, validateSharedConstraints(set, rows)...)
}

// existingK3sTarget loads and validates the spec row an existing-k3s plan targets.
func existingK3sTarget(root, environment string) (*infraEnvironment, []error) {
	set, errs := loadInfraEnvironments(root)
	if len(errs) > 0 {
		return nil, errs
	}
	if errs := validateInfraEnvironments(set, environment); len(errs) > 0 {
		return nil, errs
	}
	return set.find(environment), nil
}

func cmdSpec(args []string) {
	if len(args) == 0 || args[0] != "environments" {
		exitSlotErrors([]error{errors.New("SLOTS-ARGUMENT-INVALID: Usage: smaqit spec environments [--env=<env>] [--json]")})
	}
	root, err := findProjectRoot(".")
	if err != nil {
		exitSlotErrors([]error{fmt.Errorf("SLOTS-ARGUMENT-INVALID: %v; run 'smaqit init' first", err)})
	}
	environment, asJSON := "", false
	for _, arg := range args[1:] {
		switch {
		case arg == "--json":
			asJSON = true
		case strings.HasPrefix(arg, "--env="):
			environment = strings.TrimPrefix(arg, "--env=")
		default:
			exitSlotErrors([]error{fmt.Errorf("SLOTS-ARGUMENT-INVALID: unknown flag %q", arg)})
		}
	}
	set, errs := loadInfraEnvironments(root)
	if len(errs) == 0 {
		errs = validateInfraEnvironments(set, environment)
	}
	if len(errs) > 0 {
		exitSlotErrors(errs)
	}
	if environment != "" {
		set.Environments = []infraEnvironment{*set.find(environment)}
	}
	if asJSON {
		out, _ := json.MarshalIndent(set, "", "  ")
		fmt.Println(string(out))
		return
	}
	fmt.Printf("Environments from %s:\n\n", set.Spec)
	for _, env := range set.Environments {
		fmt.Printf("  %-12s machine=%s reachability=%s ingress=%s host=%s health=%s\n",
			env.Environment, env.MachineSlug, env.Reachability, env.IngressMode, orDash(env.Host), env.HealthPath)
	}
	if set.Mesh != nil && set.Mesh.Action != "" {
		fmt.Printf("\n  mesh join: %s (inputs %s, %s)\n", set.Mesh.Action, set.Mesh.SetupKeyInput, set.Mesh.WaitForInput)
	}
	if set.ProbeClient != "" {
		fmt.Printf("  probe client: %s\n", set.ProbeClient)
	}
}

func orDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}
