#!/usr/bin/env node

import { readFile, writeFile } from "node:fs/promises";
import { resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { Resvg, initWasm } from "@resvg/resvg-wasm";

const runtimeDir = resolve(fileURLToPath(new URL(".", import.meta.url)));
const [svgPath, pngPath, requestedWidth = "1800"] = process.argv.slice(2);

if (!svgPath || !pngPath) {
  console.error("usage: render-png.mjs <input.svg> <output.png> [width]");
  process.exit(2);
}

const width = Number.parseInt(requestedWidth, 10);
if (!Number.isInteger(width) || width < 640 || width > 4096) {
  console.error("width must be an integer between 640 and 4096");
  process.exit(2);
}

const wasm = await readFile(resolve(runtimeDir, "node_modules/@resvg/resvg-wasm/index_bg.wasm"));
const font = await readFile(resolve(runtimeDir, "node_modules/@fontsource/noto-sans/files/noto-sans-latin-400-normal.woff2"));
await initWasm(wasm);

const svg = await readFile(svgPath);
const renderer = new Resvg(svg, {
  fitTo: { mode: "width", value: width },
  background: "#FFF9F0",
  font: {
    fontBuffers: [font],
    defaultFontFamily: "Noto Sans",
  },
});
const png = renderer.render().asPng();
await writeFile(pngPath, png);
