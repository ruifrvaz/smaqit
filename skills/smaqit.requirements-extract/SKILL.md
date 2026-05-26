---
name: smaqit.requirements-extract
description: Use this skill when the user wants to extract structured requirements candidates from raw project assets such as code prototypes, documents, or meeting notes. Produces `.smaqit/requirements-extract.md` — an enumerated, categorised inventory that specification agents can consume directly. Also use when the user asks to analyse assets, mine files for requirements, or prepare input for any spec workflow.
metadata:
  version: "1.0.0"
---

# Requirements Extraction

## Steps

1. **Determine the asset path.** Use `assets/raw/` unless the user specifies another path. If the directory is absent or empty, ask for the path before proceeding.

2. **Read every file in the directory.** If any file has a `.pdf` extension, invoke `smaqit.utils.read-pdf` first and work from the sidecar `.extracted.txt` file it produces.

3. **Scan each file in full** — including code logic, not just comments and data declarations — and extract items into these six fixed categories:

   | Category | What to capture |
   |---|---|
   | **Data models** | Entity shapes, field names, types, keys |
   | **Business formulas** | Scoring algorithms, calculation rules, thresholds (e.g. attrition risk tiers, score normalisation) |
   | **Navigation / UI flows** | Screen names, tab lists, routing logic, onboarding gates |
   | **State machine rules** | Locking conditions, completion triggers, immutability constraints |
   | **Integration contracts** | API shapes, response keys, 3rd-party SDKs referenced |
   | **Constraints and anomalies** | Range discrepancies, deviations from standard patterns, explicit exclusions |

   For each item, record a **source annotation**: filename + line number or paragraph reference. Do **not** normalise conflicting values — preserve discrepancies exactly as found.

4. **Write the inventory to `.smaqit/requirements-extract.md`** using this structure:
   - One `##` heading per category (all six must be present, in the order above)
   - Bullet items under each heading; each bullet ends with its source annotation in parentheses
   - Write `None identified.` under any empty category — do not omit the heading
   - `## Ambiguities` section at the end listing every conflicting value, undefined term, or unclear constraint found; each entry states what conflicts and which source files are involved

5. **Report to the user:** item count per category and total ambiguity count. If the ambiguity count exceeds 5, name the three most impactful and recommend resolving them before spec generation.

## Output

`.smaqit/requirements-extract.md` — structured inventory with six categorised sections and an `## Ambiguities` section.

## Scope

- Does NOT write specifications. Output is raw material for specification agents (`smaqit.business`, `smaqit.functional`, etc.).
- Does NOT infer requirements absent from the source files.
- Does NOT process PDFs directly — invoke `smaqit.utils.read-pdf` first.

## Examples

**Input:** User places `code.txt` (React prototype, ~1500 lines) in `assets/raw/` and invokes `/requirements.extract`.

**Output:** `.smaqit/requirements-extract.md` containing score formulas, 7 tab names, habit data model, baseline locking trigger, and 2 flagged range discrepancies across 6 categorised sections.

## Gotchas

- Source files may use inconsistent ranges for overlapping concepts. In HIM Corporate, `CD-RISC` uses a 0–4 range while other instrument scales use 1–5. Annotate range values with their source file and do not normalise — preserve the discrepancy as-is and flag it in `## Ambiguities`.
- Code files (e.g. React components) embed formulas and state machine rules inside component logic, not just in data declarations or comments. Scan the full file body.
- The extraction output is not a specification. Functional specification agents resolve ambiguities — this skill only surfaces them.

## Completion

- [ ] All files in the asset directory read (PDF files processed via `smaqit.utils.read-pdf` first)
- [ ] `.smaqit/requirements-extract.md` written
- [ ] All six category headings present (empty categories contain "None identified.")
- [ ] `## Ambiguities` section present
- [ ] Report delivered: item count per category + ambiguity count

## Failure Handling

| Situation | Action |
|---|---|
| Required input not provided | Request the missing information before proceeding |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification |
| Subagent invocation fails | Report the failure with context; do not silently retry |
| Output artifact already exists | Confirm with user before overwriting |
| `assets/raw/` is absent or empty | Ask the user for the asset path before proceeding |
| File is a PDF | Invoke `smaqit.utils.read-pdf` first, then continue with the sidecar |
| Ambiguity count exceeds 5 | Summarise the top 3 most impactful in the report; recommend resolving before spec generation |
