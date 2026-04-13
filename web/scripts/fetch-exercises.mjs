#!/usr/bin/env node
/**
 * Downloads the latest exercises.json from free-exercise-db.
 * Run: node web/scripts/fetch-exercises.mjs
 */

import { createWriteStream, mkdirSync } from "node:fs";
import { pipeline } from "node:stream/promises";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const URL =
  "https://raw.githubusercontent.com/yuhonas/free-exercise-db/main/dist/exercises.json";

const __dirname = dirname(fileURLToPath(import.meta.url));
const OUT = join(__dirname, "../public/exercises.json");

console.log(`Fetching ${URL} ...`);

const res = await fetch(URL);
if (!res.ok) {
  console.error(`Failed: HTTP ${res.status} ${res.statusText}`);
  process.exit(1);
}

await pipeline(res.body, createWriteStream(OUT));

const bytes = (await import("node:fs")).statSync(OUT).size;
console.log(`Saved to web/public/exercises.json (${(bytes / 1024).toFixed(1)} KB)`);
