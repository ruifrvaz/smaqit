package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed templates/slots.yml.template
var slotsTemplate embed.FS

// slotsConfigPath is the project-owned binding file. It is created once by
// init and never overwritten by smaqit afterwards.
const slotsConfigPath = ".smaqit/slots.yml"

// Default filler kinds in the slot catalogue.
const (
	fillerNone    = "none"    // slot is skipped
	fillerSkill   = "skill"   // one named skill
	fillerFamily  = "family"  // a skill family selected by the stack (name prefix)
	fillerBuiltin = "builtin" // performed by the deployment agent itself
)

var provisioningModes = []string{"provision", "existing-owned", "existing-shared", "existing-unmanaged", "existing-k3s"}

var deploymentPaths = []string{"standard", "existing-cicd-pr", "ci-push"}

type slotFiller struct {
	Kind      string
	Name      string
	Operation string
}

type slotDef struct {
	Name     string
	DayTwo   bool
	Purpose  string
	Defaults map[string]slotFiller // provisioning_mode → default filler
}

func noFiller() slotFiller { return slotFiller{Kind: fillerNone} }
func skillFiller(name, operation string) slotFiller {
	return slotFiller{Kind: fillerSkill, Name: name, Operation: operation}
}
func familyFiller(prefix, operation string) slotFiller {
	return slotFiller{Kind: fillerFamily, Name: prefix, Operation: operation}
}
func builtinFiller(name string) slotFiller { return slotFiller{Kind: fillerBuiltin, Name: name} }

// byMode builds a per-mode default table. Modes absent from overrides take fallback.
func byMode(fallback slotFiller, overrides map[string]slotFiller) map[string]slotFiller {
	defaults := map[string]slotFiller{}
	for _, mode := range provisioningModes {
		if filler, ok := overrides[mode]; ok {
			defaults[mode] = filler
		} else {
			defaults[mode] = fallback
		}
	}
	return defaults
}

func vmModes(filler slotFiller) map[string]slotFiller {
	return map[string]slotFiller{
		"provision": filler, "existing-owned": filler, "existing-shared": filler, "existing-unmanaged": filler,
	}
}

const (
	skillVaultLoader  = "smaqit.infrastructure-vault-loader"
	skillOnboarding   = "smaqit.infrastructure-request-k3s-onboarding"
	skillCICDGenerate = "smaqit.infrastructure-cicd-generate"
	skillProvision    = "smaqit.infrastructure-provision-cyso"
	skillVMBootstrap  = "smaqit.infrastructure-vm-bootstrap"
	skillRepoConfig   = "smaqit.infrastructure-repo-config"
	skillDeployVerify = "smaqit.infrastructure-deploy-verify"
	skillDomainTLS    = "smaqit.infrastructure-domain-tls"
	familyImageBuild  = "smaqit.infrastructure-image-build-push"
	familyDeployRsync = "smaqit.infrastructure-deploy-rsync"
	familyDeployK3s   = "smaqit.infrastructure-deploy-k3s-app"
)

// deploySlots is the deploy-phase slot catalogue in execution order. Day-two
// slots never run as part of a full sequence; they are invoked one at a time.
var deploySlots = []slotDef{
	{
		Name:     "platform-prereqs",
		Purpose:  "Prepare platform-level prerequisites the target needs before credentials or onboarding",
		Defaults: byMode(noFiller(), nil),
	},
	{
		Name:     "access-credentials",
		Purpose:  "Load the credentials needed to reach the provider, platform, and repository",
		Defaults: byMode(skillFiller(skillVaultLoader, "access"), nil),
	},
	{
		Name:     "platform-onboarding",
		Purpose:  "Request tenancy on a platform-owned target and gate on its approval",
		Defaults: byMode(noFiller(), map[string]slotFiller{"existing-k3s": skillFiller(skillOnboarding, "request")}),
	},
	{
		Name:    "target-credentials",
		Purpose: "Load credentials the target issues or authorizes for this app",
		Defaults: byMode(skillFiller(skillVaultLoader, "target"), map[string]slotFiller{
			"provision": noFiller(), "existing-owned": noFiller(),
		}),
	},
	{
		Name:     "artifact-build",
		Purpose:  "Author and build deployable artifacts (container image, manifests)",
		Defaults: byMode(noFiller(), map[string]slotFiller{"existing-k3s": familyFiller(familyImageBuild, "build")}),
	},
	{
		Name:    "iac-generate",
		Purpose: "Generate Infrastructure as Code from the consolidated specifications",
		Defaults: byMode(noFiller(), map[string]slotFiller{
			"provision": builtinFiller("generate-iac"), "existing-owned": builtinFiller("generate-iac"),
		}),
	},
	{
		Name:     "ci-generate",
		Purpose:  "Generate or update the CI/CD workflows that deploy the target",
		Defaults: byMode(skillFiller(skillCICDGenerate, "generate"), nil),
	},
	{
		Name:    "provision",
		Purpose: "Provision or reconcile the target infrastructure through the guarded IaC tool",
		Defaults: byMode(noFiller(), map[string]slotFiller{
			"provision": skillFiller(skillProvision, "apply"), "existing-owned": skillFiller(skillProvision, "apply"),
		}),
	},
	{
		Name:     "host-bootstrap",
		Purpose:  "Bootstrap the host operating system and runtime the app deploys onto",
		Defaults: byMode(noFiller(), vmModes(skillFiller(skillVMBootstrap, "bootstrap"))),
	},
	{
		Name:     "environment-config",
		Purpose:  "Configure repository and environment secrets and variables for CI/CD",
		Defaults: byMode(skillFiller(skillRepoConfig, "sync"), nil),
	},
	{
		Name:    "app-deploy",
		Purpose: "Deploy the application onto the target",
		Defaults: byMode(familyFiller(familyDeployRsync, "deploy"), map[string]slotFiller{
			"existing-k3s": familyFiller(familyDeployK3s, "deploy"),
		}),
	},
	{
		Name:     "verify",
		Purpose:  "Verify the deployed revision is healthy and serving",
		Defaults: byMode(skillFiller(skillDeployVerify, "verify"), nil),
	},
	{
		Name:     "tls",
		Purpose:  "Issue and install TLS for a declared custom domain",
		Defaults: byMode(noFiller(), vmModes(skillFiller(skillDomainTLS, "issue"))),
	},
	{
		Name:     "replace",
		DayTwo:   true,
		Purpose:  "Replace a target in place (day-two operation)",
		Defaults: byMode(noFiller(), nil),
	},
	{
		Name:     "teardown",
		DayTwo:   true,
		Purpose:  "Destroy the target through the guarded IaC tool (day-two operation)",
		Defaults: byMode(noFiller(), map[string]slotFiller{"provision": builtinFiller("iac-destroy")}),
	},
}

// pathRule adjusts the mode defaults for a deployment path.
type pathRule struct {
	Refused     map[string]bool // binding refused, slot always skipped
	DefaultNone map[string]bool // full-sequence default is none; a binding or a single-slot run still fills it
	CIExecuted  map[string]bool // performed by the CI pipeline; the agent monitors
}

func slotSet(names ...string) map[string]bool {
	set := map[string]bool{}
	for _, name := range names {
		set[name] = true
	}
	return set
}

var deploymentPathRules = map[string]pathRule{
	"standard": {
		DefaultNone: slotSet("environment-config", "tls"),
	},
	"existing-cicd-pr": {
		Refused:     slotSet("iac-generate", "ci-generate", "provision", "host-bootstrap"),
		DefaultNone: slotSet("platform-onboarding", "tls"),
		CIExecuted:  slotSet("artifact-build", "app-deploy"),
	},
	// ci-push mirrors greenfield production: workflows were generated and
	// committed in the dev sweep, the pipeline provisions and deploys, and the
	// production host is configured by provisioning (cloud-init), not bootstrap.
	"ci-push": {
		DefaultNone: slotSet("iac-generate", "ci-generate", "host-bootstrap", "tls"),
		CIExecuted:  slotSet("artifact-build", "provision", "app-deploy"),
	},
}

func findSlotDef(name string) (slotDef, bool) {
	for _, def := range deploySlots {
		if def.Name == name {
			return def, true
		}
	}
	return slotDef{}, false
}

// slotBinding is one project binding from .smaqit/slots.yml.
type slotBinding struct {
	Value        string // "default", "none", or "" for a skill binding
	Skill        string
	Operation    string
	Params       map[string]string
	Environments []string
}

type slotsConfig struct {
	Path     string
	Bindings map[string]slotBinding
}

var slotsSkillNamePattern = regexp.MustCompile(`^[a-z0-9]+([.-][a-z0-9]+)*$`)
var slotsOperationPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// loadSlotsConfig reads and strictly validates .smaqit/slots.yml under root.
// A missing file is not an error: every slot then resolves to its profile default.
func loadSlotsConfig(root string) (*slotsConfig, []error) {
	path := filepath.Join(root, slotsConfigPath)
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, []error{fmt.Errorf("SLOTS-CONFIG-INVALID: reading %s: %w", slotsConfigPath, err)}
	}
	return parseSlotsConfig(path, b)
}

func parseSlotsConfig(path string, b []byte) (*slotsConfig, []error) {
	content := strings.ReplaceAll(string(b), "\r\n", "\n")
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(content), &doc); err != nil {
		return nil, []error{fmt.Errorf("SLOTS-CONFIG-INVALID: %s is not valid YAML: %w", slotsConfigPath, err)}
	}
	if len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
		return nil, []error{fmt.Errorf("SLOTS-CONFIG-INVALID: %s must be a mapping", slotsConfigPath)}
	}
	cfg := &slotsConfig{Path: path, Bindings: map[string]slotBinding{}}
	var errs []error
	top := doc.Content[0]
	found := map[string]bool{}
	for i := 0; i < len(top.Content); i += 2 {
		key, value := top.Content[i].Value, top.Content[i+1]
		found[key] = true
		switch key {
		case "schemaVersion":
			if value.Kind != yaml.ScalarNode || value.Value != "1" {
				errs = append(errs, fmt.Errorf("SLOTS-CONFIG-INVALID: schemaVersion must be 1, got %q", value.Value))
			}
		case "phase":
			if value.Kind != yaml.ScalarNode || value.Value != "deploy" {
				errs = append(errs, fmt.Errorf("SLOTS-CONFIG-INVALID: phase must be \"deploy\", got %q", value.Value))
			}
		case "slots":
			errs = append(errs, parseSlotBindings(value, cfg.Bindings)...)
		default:
			errs = append(errs, fmt.Errorf("SLOTS-CONFIG-INVALID: unsupported top-level field %q", key))
		}
	}
	for _, key := range []string{"schemaVersion", "phase"} {
		if !found[key] {
			errs = append(errs, fmt.Errorf("SLOTS-CONFIG-INVALID: missing top-level field %q", key))
		}
	}
	return cfg, errs
}

func parseSlotBindings(node *yaml.Node, bindings map[string]slotBinding) []error {
	if node.Kind == yaml.ScalarNode && node.Tag == "!!null" {
		return nil
	}
	if node.Kind != yaml.MappingNode {
		return []error{errors.New("SLOTS-CONFIG-INVALID: slots must be a mapping of slot name to binding")}
	}
	var errs []error
	for i := 0; i < len(node.Content); i += 2 {
		name, value := node.Content[i].Value, node.Content[i+1]
		if _, ok := findSlotDef(name); !ok {
			errs = append(errs, fmt.Errorf("SLOTS-UNKNOWN-SLOT: %q is not a deploy-phase slot (known: %s)", name, strings.Join(slotNames(true), ", ")))
			continue
		}
		binding, err := parseSlotBinding(name, value)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		bindings[name] = binding
	}
	return errs
}

func parseSlotBinding(slot string, node *yaml.Node) (slotBinding, error) {
	if node.Kind == yaml.ScalarNode {
		if node.Value == "default" || node.Value == "none" {
			return slotBinding{Value: node.Value}, nil
		}
		return slotBinding{}, fmt.Errorf("SLOTS-CONFIG-INVALID: slot %q must be \"default\", \"none\", or a mapping with a skill", slot)
	}
	if node.Kind != yaml.MappingNode {
		return slotBinding{}, fmt.Errorf("SLOTS-CONFIG-INVALID: slot %q must be \"default\", \"none\", or a mapping with a skill", slot)
	}
	binding := slotBinding{}
	for i := 0; i < len(node.Content); i += 2 {
		key, value := node.Content[i].Value, node.Content[i+1]
		switch key {
		case "skill":
			if value.Kind != yaml.ScalarNode || !slotsSkillNamePattern.MatchString(value.Value) {
				return slotBinding{}, fmt.Errorf("SLOTS-CONFIG-INVALID: slot %q skill must be a skill name, got %q", slot, value.Value)
			}
			binding.Skill = value.Value
		case "operation":
			if value.Kind != yaml.ScalarNode || !slotsOperationPattern.MatchString(value.Value) {
				return slotBinding{}, fmt.Errorf("SLOTS-CONFIG-INVALID: slot %q operation must be a lowercase hyphenated name, got %q", slot, value.Value)
			}
			binding.Operation = value.Value
		case "params":
			if value.Kind != yaml.MappingNode {
				return slotBinding{}, fmt.Errorf("SLOTS-CONFIG-INVALID: slot %q params must be a mapping", slot)
			}
			binding.Params = map[string]string{}
			for j := 0; j < len(value.Content); j += 2 {
				if value.Content[j+1].Kind != yaml.ScalarNode {
					return slotBinding{}, fmt.Errorf("SLOTS-CONFIG-INVALID: slot %q param %q must be a scalar", slot, value.Content[j].Value)
				}
				binding.Params[value.Content[j].Value] = value.Content[j+1].Value
			}
		case "environments":
			if value.Kind != yaml.SequenceNode || len(value.Content) == 0 {
				return slotBinding{}, fmt.Errorf("SLOTS-CONFIG-INVALID: slot %q environments must be a non-empty list", slot)
			}
			for _, env := range value.Content {
				if env.Kind != yaml.ScalarNode || env.Value == "" {
					return slotBinding{}, fmt.Errorf("SLOTS-CONFIG-INVALID: slot %q environments must contain names", slot)
				}
				binding.Environments = append(binding.Environments, env.Value)
			}
		default:
			return slotBinding{}, fmt.Errorf("SLOTS-CONFIG-INVALID: slot %q has unsupported field %q", slot, key)
		}
	}
	if binding.Skill == "" {
		return slotBinding{}, fmt.Errorf("SLOTS-CONFIG-INVALID: slot %q binding requires a skill", slot)
	}
	return binding, nil
}

func (b slotBinding) appliesTo(env string) bool {
	if len(b.Environments) == 0 {
		return true
	}
	for _, candidate := range b.Environments {
		if candidate == env {
			return true
		}
	}
	return false
}

func slotNames(includeDayTwo bool) []string {
	var names []string
	for _, def := range deploySlots {
		if def.DayTwo && !includeDayTwo {
			continue
		}
		names = append(names, def.Name)
	}
	return names
}

// installedSkill is a SKILL.md found on disk with its declared slots.
type installedSkill struct {
	Name  string
	Path  string
	Slots map[string][]string // slot → declared operations
}

// skillSearchDirs lists skill roots in lookup order: project-local first, then
// the global shared tree (Copilot, Codex), then the Claude Code tree.
func skillSearchDirs(root string) []string {
	dirs := []string{
		filepath.Join(root, ".agents", "skills"),
		filepath.Join(root, ".claude", "skills"),
		filepath.Join(root, ".github", "skills"),
	}
	for _, kind := range []string{"shared-skills", "claude-skills"} {
		if dir, err := resolveGlobalDir(kind); err == nil {
			dirs = append(dirs, dir)
		}
	}
	return dirs
}

type skillIndex struct {
	dirs  []string
	cache map[string]*installedSkill
	errs  map[string]error
}

func newSkillIndex(root string) *skillIndex {
	return &skillIndex{dirs: skillSearchDirs(root), cache: map[string]*installedSkill{}, errs: map[string]error{}}
}

// find returns the first installed copy of a skill, or nil when none exists.
func (idx *skillIndex) find(name string) (*installedSkill, error) {
	if skill, ok := idx.cache[name]; ok {
		return skill, idx.errs[name]
	}
	var found *installedSkill
	var findErr error
	for _, dir := range idx.dirs {
		path := filepath.Join(dir, name, "SKILL.md")
		if _, err := os.Stat(path); err != nil {
			continue
		}
		slots, err := readSkillSlots(path)
		found = &installedSkill{Name: name, Path: path, Slots: slots}
		findErr = err
		break
	}
	idx.cache[name] = found
	idx.errs[name] = findErr
	return found, findErr
}

// family returns installed skills named prefix or prefix-*, sorted by name.
func (idx *skillIndex) family(prefix string) []*installedSkill {
	seen := map[string]bool{}
	var names []string
	for _, dir := range idx.dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			name := entry.Name()
			if !entry.IsDir() || seen[name] || (name != prefix && !strings.HasPrefix(name, prefix+"-")) {
				continue
			}
			seen[name] = true
			names = append(names, name)
		}
	}
	sort.Strings(names)
	var skills []*installedSkill
	for _, name := range names {
		if skill, err := idx.find(name); skill != nil && err == nil {
			skills = append(skills, skill)
		}
	}
	return skills
}

// readSkillSlots parses metadata.smaqit-slots from a SKILL.md frontmatter.
func readSkillSlots(path string) (map[string][]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseSkillSlots(path, b)
}

func parseSkillSlots(path string, b []byte) (map[string][]string, error) {
	content := strings.ReplaceAll(string(b), "\r\n", "\n")
	if !strings.HasPrefix(content, "---\n") {
		return nil, fmt.Errorf("SLOTS-SKILL-INVALID: %s has no YAML frontmatter", path)
	}
	closing := strings.Index(content[4:], "\n---")
	if closing < 0 {
		return nil, fmt.Errorf("SLOTS-SKILL-INVALID: %s frontmatter has no closing fence", path)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(content[4:4+closing]), &doc); err != nil {
		return nil, fmt.Errorf("SLOTS-SKILL-INVALID: %s frontmatter is not valid YAML: %w", path, err)
	}
	slots := map[string][]string{}
	if len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
		return slots, nil
	}
	metadata := mappingValue(doc.Content[0], "metadata")
	if metadata == nil {
		return slots, nil
	}
	if metadata.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("SLOTS-SKILL-INVALID: %s metadata must be a mapping", path)
	}
	declared := mappingValue(metadata, "smaqit-slots")
	if declared == nil {
		return slots, nil
	}
	if declared.Kind != yaml.ScalarNode {
		return nil, fmt.Errorf("SLOTS-SKILL-INVALID: %s metadata.smaqit-slots must be a string of slot=operation pairs", path)
	}
	for _, pair := range strings.Split(declared.Value, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		slot, operation, ok := strings.Cut(pair, "=")
		slot, operation = strings.TrimSpace(slot), strings.TrimSpace(operation)
		if !ok || !slotsOperationPattern.MatchString(operation) {
			return nil, fmt.Errorf("SLOTS-SKILL-INVALID: %s metadata.smaqit-slots entry %q must be slot=operation", path, pair)
		}
		if _, known := findSlotDef(slot); !known {
			return nil, fmt.Errorf("SLOTS-SKILL-INVALID: %s declares unknown slot %q", path, slot)
		}
		slots[slot] = append(slots[slot], operation)
	}
	return slots, nil
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

// slotPlanEntry is one resolved slot.
type slotPlanEntry struct {
	Order      int               `json:"order"`
	Slot       string            `json:"slot"`
	Source     string            `json:"source"`   // bound, profile, none, refused
	Executor   string            `json:"executor"` // local, ci, or "" when skipped
	Kind       string            `json:"kind"`     // skill, family, builtin, none
	Skill      string            `json:"skill,omitempty"`
	Operation  string            `json:"operation,omitempty"`
	SkillPath  string            `json:"skillPath,omitempty"`
	Family     string            `json:"family,omitempty"`
	Candidates []familyCandidate `json:"candidates,omitempty"`
	Builtin    string            `json:"builtin,omitempty"`
	Params     map[string]string `json:"params,omitempty"`
	Purpose    string            `json:"purpose"`
}

type familyCandidate struct {
	Skill     string `json:"skill"`
	SkillPath string `json:"skillPath"`
}

type slotPlan struct {
	SchemaVersion int             `json:"schemaVersion"`
	Phase         string          `json:"phase"`
	Mode          string          `json:"mode"`
	Environment   string          `json:"environment"`
	Path          string          `json:"deploymentPath"`
	Config        string          `json:"config"`
	Slots         []slotPlanEntry `json:"slots"`
	Warnings      []string        `json:"warnings,omitempty"`
}

type slotPlanOptions struct {
	Mode        string
	Environment string
	Path        string
	Slot        string // single-slot (day-two) run when set
}

func validateSlotPlanOptions(opts slotPlanOptions) []error {
	var errs []error
	if !containsString(provisioningModes, opts.Mode) {
		errs = append(errs, fmt.Errorf("SLOTS-ARGUMENT-INVALID: --mode must be one of %s", strings.Join(provisioningModes, ", ")))
	}
	if !containsString(deploymentPaths, opts.Path) {
		errs = append(errs, fmt.Errorf("SLOTS-ARGUMENT-INVALID: --path must be one of %s", strings.Join(deploymentPaths, ", ")))
	}
	if opts.Environment == "" {
		errs = append(errs, errors.New("SLOTS-ARGUMENT-INVALID: --env is required"))
	}
	if opts.Slot != "" {
		if _, ok := findSlotDef(opts.Slot); !ok {
			errs = append(errs, fmt.Errorf("SLOTS-ARGUMENT-INVALID: --slot %q is not a deploy-phase slot", opts.Slot))
		}
	}
	return errs
}

// resolveSlotPlan resolves the ordered slot plan for one environment run, or a
// single slot when opts.Slot is set. Every problem is reported, not only the first.
func resolveSlotPlan(root string, opts slotPlanOptions) (*slotPlan, []error) {
	if errs := validateSlotPlanOptions(opts); len(errs) > 0 {
		return nil, errs
	}
	cfg, errs := loadSlotsConfig(root)
	if len(errs) > 0 {
		return nil, errs
	}
	plan := &slotPlan{SchemaVersion: 1, Phase: "deploy", Mode: opts.Mode, Environment: opts.Environment, Path: opts.Path}
	if cfg != nil {
		plan.Config = slotsConfigPath
	}
	rule := deploymentPathRules[opts.Path]
	idx := newSkillIndex(root)
	order := 0
	for _, def := range deploySlots {
		if opts.Slot != "" && def.Name != opts.Slot {
			continue
		}
		if opts.Slot == "" && def.DayTwo {
			continue
		}
		order++
		entry := slotPlanEntry{Order: order, Slot: def.Name, Purpose: def.Purpose}
		var binding *slotBinding
		if cfg != nil {
			if b, ok := cfg.Bindings[def.Name]; ok && b.appliesTo(opts.Environment) {
				binding = &b
			}
		}
		switch {
		case rule.Refused[def.Name]:
			if binding != nil && binding.Value != "default" && binding.Value != "none" {
				errs = append(errs, fmt.Errorf("SLOTS-BINDING-REFUSED: slot %q cannot be bound under deployment path %q", def.Name, opts.Path))
			}
			entry.Source, entry.Kind = "refused", fillerNone
		case binding != nil && binding.Value == "none":
			entry.Source, entry.Kind = "none", fillerNone
		case binding != nil && binding.Value == "":
			entry.Source = "bound"
			if err := resolveBoundEntry(idx, def.Name, *binding, &entry); err != nil {
				errs = append(errs, err)
			}
		default:
			filler := def.Defaults[opts.Mode]
			if rule.DefaultNone[def.Name] && opts.Slot == "" {
				filler = noFiller()
			}
			entry.Source = "profile"
			if filler.Kind == fillerNone {
				entry.Source = "none"
			}
			warning, err := resolveProfileEntry(idx, def.Name, filler, &entry)
			if err != nil {
				errs = append(errs, err)
			}
			if warning != "" {
				plan.Warnings = append(plan.Warnings, warning)
			}
		}
		if entry.Kind != fillerNone {
			entry.Executor = "local"
			if rule.CIExecuted[def.Name] {
				entry.Executor = "ci"
			}
		}
		plan.Slots = append(plan.Slots, entry)
	}
	if len(errs) > 0 {
		return nil, errs
	}
	return plan, nil
}

func resolveBoundEntry(idx *skillIndex, slot string, binding slotBinding, entry *slotPlanEntry) error {
	entry.Kind, entry.Skill, entry.Params = fillerSkill, binding.Skill, binding.Params
	skill, err := idx.find(binding.Skill)
	if err != nil {
		return err
	}
	if skill == nil {
		return fmt.Errorf("SLOTS-SKILL-MISSING: slot %q is bound to %q, which is not installed in project-local or global skill directories", slot, binding.Skill)
	}
	entry.SkillPath = skill.Path
	operation, err := selectOperation(slot, skill, binding.Operation)
	if err != nil {
		return err
	}
	entry.Operation = operation
	return nil
}

func resolveProfileEntry(idx *skillIndex, slot string, filler slotFiller, entry *slotPlanEntry) (string, error) {
	entry.Kind = filler.Kind
	switch filler.Kind {
	case fillerBuiltin:
		entry.Builtin = filler.Name
	case fillerSkill:
		entry.Skill = filler.Name
		skill, err := idx.find(filler.Name)
		if err != nil {
			return "", err
		}
		if skill == nil {
			return "", fmt.Errorf("SLOTS-SKILL-MISSING: default filler %q for slot %q is not installed; reinstall smaqit's global payload", filler.Name, slot)
		}
		entry.SkillPath = skill.Path
		operation, err := selectOperation(slot, skill, filler.Operation)
		if err != nil {
			return "", err
		}
		entry.Operation = operation
	case fillerFamily:
		entry.Family, entry.Operation = filler.Name, filler.Operation
		for _, skill := range idx.family(filler.Name) {
			if containsString(skill.Slots[slot], filler.Operation) {
				entry.Candidates = append(entry.Candidates, familyCandidate{Skill: skill.Name, SkillPath: skill.Path})
			}
		}
		if len(entry.Candidates) == 0 {
			return fmt.Sprintf("slot %q: no installed %s* skill declares %s=%s; the orchestrator must match or synthesize one", slot, filler.Name, slot, filler.Operation), nil
		}
	}
	return "", nil
}

// selectOperation checks the skill declares the slot and picks the operation.
func selectOperation(slot string, skill *installedSkill, requested string) (string, error) {
	declared := skill.Slots[slot]
	if len(declared) == 0 {
		return "", fmt.Errorf("SLOTS-SLOT-UNDECLARED: skill %q does not declare slot %q in metadata.smaqit-slots (%s)", skill.Name, slot, skill.Path)
	}
	if requested != "" {
		if !containsString(declared, requested) {
			return "", fmt.Errorf("SLOTS-OPERATION-INVALID: skill %q declares slot %q with operation(s) %s, not %q", skill.Name, slot, strings.Join(declared, ", "), requested)
		}
		return requested, nil
	}
	if len(declared) > 1 {
		return "", fmt.Errorf("SLOTS-OPERATION-INVALID: skill %q declares slot %q with several operations (%s); set operation in the binding", skill.Name, slot, strings.Join(declared, ", "))
	}
	return declared[0], nil
}

// validateSlotBindings checks the config and every skill binding independently
// of mode, environment, and path. Used by `smaqit slots validate` and `smaqit validate`.
func validateSlotBindings(root string) []error {
	cfg, errs := loadSlotsConfig(root)
	if len(errs) > 0 || cfg == nil {
		return errs
	}
	idx := newSkillIndex(root)
	names := make([]string, 0, len(cfg.Bindings))
	for name := range cfg.Bindings {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		binding := cfg.Bindings[name]
		if binding.Value != "" {
			continue
		}
		var entry slotPlanEntry
		if err := resolveBoundEntry(idx, name, binding, &entry); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// installSlotsSkeleton writes the commented .smaqit/slots.yml once. An existing
// file is project-owned and is never overwritten.
func installSlotsSkeleton(root string) (string, error) {
	path := filepath.Join(root, slotsConfigPath)
	if _, err := os.Stat(path); err == nil {
		return "preserved", nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	content, err := slotsTemplate.ReadFile("templates/slots.yml.template")
	if err != nil {
		return "", fmt.Errorf("reading slots template: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		return "", err
	}
	return "created", nil
}

func containsString(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func cmdSlots(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: smaqit slots <plan|validate> [flags]")
		os.Exit(1)
	}
	root, err := findProjectRoot(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "SLOTS-ARGUMENT-INVALID: %v; run 'smaqit init' first\n", err)
		os.Exit(1)
	}
	switch args[0] {
	case "plan":
		opts := slotPlanOptions{Path: "standard"}
		asJSON := false
		var argErrs []error
		for _, arg := range args[1:] {
			switch {
			case arg == "--json":
				asJSON = true
			case strings.HasPrefix(arg, "--phase="):
				if phase := strings.TrimPrefix(arg, "--phase="); phase != "deploy" {
					argErrs = append(argErrs, fmt.Errorf("SLOTS-ARGUMENT-INVALID: only --phase=deploy has slots, got %q", phase))
				}
			case strings.HasPrefix(arg, "--mode="):
				opts.Mode = strings.TrimPrefix(arg, "--mode=")
			case strings.HasPrefix(arg, "--env="):
				opts.Environment = strings.TrimPrefix(arg, "--env=")
			case strings.HasPrefix(arg, "--path="):
				opts.Path = strings.TrimPrefix(arg, "--path=")
			case strings.HasPrefix(arg, "--slot="):
				opts.Slot = strings.TrimPrefix(arg, "--slot=")
			default:
				argErrs = append(argErrs, fmt.Errorf("SLOTS-ARGUMENT-INVALID: unknown flag %q", arg))
			}
		}
		if len(argErrs) > 0 {
			exitSlotErrors(argErrs)
		}
		plan, errs := resolveSlotPlan(root, opts)
		if len(errs) > 0 {
			exitSlotErrors(errs)
		}
		if asJSON {
			out, _ := json.MarshalIndent(plan, "", "  ")
			fmt.Println(string(out))
			return
		}
		printSlotPlan(plan)
	case "validate":
		if len(args) != 1 {
			exitSlotErrors([]error{errors.New("SLOTS-ARGUMENT-INVALID: Usage: smaqit slots validate")})
		}
		if errs := validateSlotBindings(root); len(errs) > 0 {
			exitSlotErrors(errs)
		}
		if _, err := os.Stat(filepath.Join(root, slotsConfigPath)); os.IsNotExist(err) {
			fmt.Println("✓ No .smaqit/slots.yml; every slot uses its profile default")
			return
		}
		fmt.Println("✓ .smaqit/slots.yml is valid and every bound skill declares its slot")
	default:
		exitSlotErrors([]error{fmt.Errorf("SLOTS-ARGUMENT-INVALID: unknown slots command %q", args[0])})
	}
}

func exitSlotErrors(errs []error) {
	for _, err := range errs {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(1)
}

func printSlotPlan(plan *slotPlan) {
	config := plan.Config
	if config == "" {
		config = "none (profile defaults)"
	}
	fmt.Printf("Deployment slot plan: mode=%s env=%s path=%s config=%s\n\n", plan.Mode, plan.Environment, plan.Path, config)
	for _, entry := range plan.Slots {
		filler := "-"
		switch entry.Kind {
		case fillerSkill:
			filler = fmt.Sprintf("%s (%s)", entry.Skill, entry.Operation)
		case fillerFamily:
			var names []string
			for _, candidate := range entry.Candidates {
				names = append(names, candidate.Skill)
			}
			filler = fmt.Sprintf("%s* (%s) candidates: %s", entry.Family, entry.Operation, strings.Join(names, ", "))
		case fillerBuiltin:
			filler = "agent builtin: " + entry.Builtin
		}
		executor := entry.Executor
		if executor == "" {
			executor = "-"
		}
		fmt.Printf("%2d  %-20s %-8s %-6s %s\n", entry.Order, entry.Slot, entry.Source, executor, filler)
	}
	for _, warning := range plan.Warnings {
		fmt.Printf("\n⚠ %s", warning)
	}
	if len(plan.Warnings) > 0 {
		fmt.Println()
	}
}

// removeSmaqitDirPreservingSlots removes root/.smaqit but restores the
// project-owned slots.yml afterwards. It reports whether the file was kept.
func removeSmaqitDirPreservingSlots(root string) (bool, error) {
	path := filepath.Join(root, slotsConfigPath)
	content, readErr := os.ReadFile(path)
	if readErr != nil && !os.IsNotExist(readErr) {
		return false, readErr
	}
	if err := os.RemoveAll(filepath.Join(root, ".smaqit")); err != nil {
		return false, err
	}
	if readErr != nil {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return false, err
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		return false, err
	}
	return true, nil
}
