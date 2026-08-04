#!/usr/bin/env node

import { createHash } from "node:crypto";
import { execFileSync } from "node:child_process";
import { cp, mkdir, readFile, readdir, rm, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const repoRoot = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const sourceDir = resolve(repoRoot, "tools/plantuml");
const outputDir = resolve(repoRoot, "installer/tools/plantuml-runtime");

execFileSync(process.platform === "win32" ? "npm.cmd" : "npm", ["ci", "--ignore-scripts", "--omit=dev", "--no-audit", "--no-fund"], {
  cwd: sourceDir,
  stdio: "inherit",
});

await rm(outputDir, { recursive: true, force: true });
await mkdir(outputDir, { recursive: true });
for (const name of ["package.json", "package-lock.json", "render-png.mjs", "THIRD_PARTY_NOTICES.md", "MPL-2.0.txt"]) {
  await cp(resolve(sourceDir, name), resolve(outputDir, name));
}
await cp(resolve(sourceDir, "node_modules"), resolve(outputDir, "node_modules"), { recursive: true, verbatimSymlinks: true });
// npm's .bin directory contains convenience symlinks. The installer invokes the
// concrete server entry point and rejects archive links as a traversal defense,
// so these aliases are deliberately excluded from the release payload.
await rm(resolve(outputDir, "node_modules/.bin"), { recursive: true, force: true });

const lock = await readFile(resolve(sourceDir, "package-lock.json"));
async function collectHashes(directory, prefix = "") {
  const hashes = {};
  const entries = await readdir(directory, { withFileTypes: true });
  entries.sort((a, b) => a.name.localeCompare(b.name));
  for (const entry of entries) {
    const relative = prefix ? `${prefix}/${entry.name}` : entry.name;
    const absolute = resolve(directory, entry.name);
    if (entry.isDirectory()) {
      Object.assign(hashes, await collectHashes(absolute, relative));
    } else if (entry.isFile()) {
      const content = await readFile(absolute);
      hashes[relative] = createHash("sha256").update(content).digest("hex");
    } else {
      throw new Error(`Unsupported runtime entry: ${relative}`);
    }
  }
  return hashes;
}
const manifest = {
  schema: 1,
  bundle_version: "plantuml-mcp-js-0.2.0_resvg-wasm-2.6.2_noto-sans-5.3.0",
  node_minimum_major: 22,
  plantuml_mcp: "0.2.0",
  resvg_wasm: "2.6.2",
  noto_sans: "5.3.0",
  package_lock_sha256: createHash("sha256").update(lock).digest("hex"),
  files: await collectHashes(outputDir),
};
await writeFile(resolve(outputDir, "manifest.json"), `${JSON.stringify(manifest, null, 2)}\n`);
