# PlantUML Design Runtime Notices

The smaqit design runtime redistributes unmodified packages installed from the exact versions recorded in `package-lock.json`.

- `@plantuml/mcp-js` — MIT License — https://github.com/plantuml/plantuml/tree/master/plantuml-mcp-js
- `@resvg/resvg-wasm` and resvg — MPL-2.0 — https://github.com/thx/resvg-js — the full license is included as `MPL-2.0.txt`
- `@fontsource/noto-sans` and Noto Sans — SIL Open Font License 1.1 — https://github.com/fontsource/font-files
- `@modelcontextprotocol/sdk` and other transitive npm packages — see each installed package's `package.json` and license file in the embedded runtime.

Smaqit does not modify the redistributed PlantUML, resvg, or Noto Sans sources. Exact package integrity values and the complete dependency graph are retained in `package-lock.json` inside the shipped runtime.
