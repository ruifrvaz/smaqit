# Skill Definition: smaqit.requirements-extract

## Identity
- **Name:** smaqit.requirements-extract
- **Version:** 1.0.0
- **Description:** Extract structured requirements candidates from raw project assets (code prototypes, documents, meeting notes) so that specification agents have a clean, enumerated input to work from.

## Steps
1. Read all files in `assets/raw/` (or the path specified by the user).
2. For each file, identify and extract the following categories:
   - **Data models** — entity shapes, field names, types, keys
   - **Business formulas** — scoring algorithms, calculation rules, thresholds (e.g. attrition risk tiers, score normalization)
   - **Navigation / UI flows** — screen names, tab lists, routing logic, onboarding gates
   - **State machine rules** — locking conditions, completion triggers, immutability constraints
   - **Integration contracts** — API shapes, response keys, 3rd-party SDKs referenced
   - **Constraints and anomalies** — range discrepancies, deviations from standard patterns, explicit exclusions
3. Write extracted items as a structured markdown inventory to `.smaqit/requirements-extract.md`. Use the categories above as H2 headings. Under each heading, list items as bullet points with a brief source annotation (filename + line or paragraph reference).
4. After extraction, flag any ambiguities found (e.g. conflicting ranges, undefined terms) in a `## Ambiguities` section at the end of the file.
5. Report: number of items extracted per category and number of ambiguities flagged.

## Output
- `.smaqit/requirements-extract.md` — structured extraction inventory

## Scope
- Does NOT write specifications. The output is raw material for specification agents (`smaqit.business`, `smaqit.functional`, etc.).
- Does NOT infer missing requirements. If something is not in the raw assets, it is not in the output.
- Does NOT process PDFs directly — use `smaqit.utils.read-pdf` first if the asset is a PDF.

## Gotchas
- Source files may use inconsistent naming (e.g., `CD-RISC` range 0–4 vs other scales 1–5 in HIM Corporate). Always annotate range values with the source file and do not normalise them — preserve the discrepancy as-is and flag it.
- `assets/raw/code.txt` in HIM Corporate was a React prototype (~1500 lines). Code files can contain formulas and state machine rules embedded in component logic — scan beyond just comments and data declarations.
- Do not treat the extraction output as final requirements. It is an extraction, not a spec. Functional agents resolve ambiguities.

## Completion
- [ ] All files in `assets/raw/` (or specified path) read
- [ ] Extraction inventory written to `.smaqit/requirements-extract.md`
- [ ] All 6 categories present in the output (even if empty — write "None identified" under empty categories)
- [ ] Ambiguities section present
- [ ] Report delivered to user (item count per category + ambiguity count)

## Failure Handling
| Situation | Action |
|-----------|--------|
| `assets/raw/` is empty or does not exist | Ask the user for the asset path before proceeding |
| File is a PDF | Invoke `smaqit.utils.read-pdf` first, then continue |
| Ambiguity count is high (>5) | Summarise the top 3 most impactful ambiguities in the report and recommend resolving them before spec generation |

## Examples
**Input:** User places `code.txt` (React prototype) in `assets/raw/` and invokes `/requirements.extract`.
**Output:** `.smaqit/requirements-extract.md` with 6 categorised sections including score formulas, 7 tab names, habit data model, baseline locking trigger, and 2 flagged range discrepancies.
