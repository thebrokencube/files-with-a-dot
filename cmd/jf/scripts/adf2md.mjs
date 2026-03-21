#!/usr/bin/env node
import { readFileSync } from 'fs';
import { AdfToMarkdownEngine } from 'extended-markdown-adf-parser';

if (process.argv.length < 3) {
  process.stderr.write('Usage: adf2md.mjs <file>\n');
  process.exit(1);
}

const json = readFileSync(process.argv[2], 'utf8');
const adf = JSON.parse(json);
const engine = new AdfToMarkdownEngine();
const md = engine.convert(adf);
process.stdout.write(md);
