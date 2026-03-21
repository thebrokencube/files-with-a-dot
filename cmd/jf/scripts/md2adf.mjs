#!/usr/bin/env node
import { readFileSync } from 'fs';
import { markdownToAdf } from 'marklassian';

if (process.argv.length < 3) {
  process.stderr.write('Usage: md2adf.mjs <file>\n');
  process.exit(1);
}

const md = readFileSync(process.argv[2], 'utf8');
console.log(JSON.stringify(markdownToAdf(md)));
