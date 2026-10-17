#!/usr/bin/env node
// Lays out a folio sketch's graph diagrams with ELK and splices the SVG into the page.
// Source of truth is diagrams.json beside index.html; every <svg data-elk="ID"> is regenerated.
import ELK from "elkjs/lib/elk.bundled.js";
import { readFileSync, writeFileSync, existsSync } from "node:fs";
import { dirname, join, resolve } from "node:path";

// Run with node; the bundled elkjs crashes under bun.
const USAGE = "usage: node sketch-diagrams.mjs <index.html> [--spec diagrams.json] [--check]";
const args = process.argv.slice(2);
if (!args.length || args.includes("-h") || args.includes("--help")) {
  console.log(USAGE);
  process.exit(args.length ? 0 : 2);
}
const page = resolve(args[0]);
const specArg = args.indexOf("--spec");
const specPath = specArg > 0 ? resolve(args[specArg + 1]) : join(dirname(page), "diagrams.json");
const checkOnly = args.includes("--check");
for (const p of [page, specPath]) if (!existsSync(p)) fail(`not found: ${p}`);

const MONO = 0.6;
const PAD = 12;
const TONES = new Set(["a", "b", "c", "d", "e", "muted", "ink"]);
const WIDE = 960;

const DEFAULTS = {
  "elk.algorithm": "layered",
  "elk.direction": "DOWN",
  "elk.edgeRouting": "ORTHOGONAL",
  "elk.layered.cycleBreaking.strategy": "MODEL_ORDER",
  "elk.layered.considerModelOrder.strategy": "NODES_AND_EDGES",
  "elk.layered.nodePlacement.strategy": "NETWORK_SIMPLEX",
  "elk.spacing.nodeNode": "28",
  "elk.layered.spacing.nodeNodeBetweenLayers": "46",
  "elk.spacing.edgeNode": "16",
  "elk.spacing.edgeEdge": "10",
  "elk.spacing.edgeLabel": "4",
  "elk.layered.spacing.edgeNodeBetweenLayers": "18",
  "elk.edgeLabels.placement": "CENTER",
  "elk.padding": "[top=14,left=14,bottom=14,right=14]",
};
const GROUP_PAD = "[top=30,left=14,bottom=14,right=14]";

const errors = [];
const warnings = [];
function fail(msg) {
  console.error(`sketch-diagrams: ${msg}`);
  process.exit(2);
}

const esc = (s) => String(s).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
const tw = (s, size, bold = false) => Math.ceil(String(s).length * size * (bold ? MONO + 0.02 : MONO));
const tone = (t) => (t === "muted" ? "var(--muted)" : t === "ink" ? "var(--ink)" : `var(--${t})`);
const label = (text, extra = {}) => ({ text: String(text), width: tw(text, 10.5) + 6, height: 14, ...extra });

const LINE_H = 15;
const flowTextTop = (n) => (n.sub ? 36 : 20);
function sizeFlowNode(n) {
  const lines = n.lines || [];
  const w = Math.max(tw(n.title, 12, true), n.sub ? tw(n.sub, 10.5) : 0, ...lines.map((l) => tw(l, 10.5))) + PAD * 2 + 6;
  const h = lines.length ? flowTextTop(n) + lines.length * LINE_H + 8 : n.sub ? 46 : 32;
  return { width: Math.max(w, 110), height: h };
}

const ER_HEAD = 24;
const ER_ROW = 17;
const ER_TAGS = { new: "a", changed: "d", existing: "muted", "code-registry": "muted", future: "muted" };
const isFuture = (n) => n.future === true || n.tag === "future";
function sizeErNode(n) {
  const rowW = Math.max(0, ...n.rows.map(([name, type]) => tw(name, 10.5) + (type ? tw(type, 10.5) + 16 : 0)));
  const head = tw(n.title, 11.5, true) + tw(n.tag, 9) + 30;
  return { width: Math.max(head, rowW) + PAD * 2, height: ER_HEAD + n.rows.length * ER_ROW + 6 };
}

function validate(id, d) {
  if (!["flow", "er"].includes(d.type)) errors.push(`${id}: type must be flow or er`);
  const ids = new Set();
  for (const n of d.nodes || []) {
    if (!n.id || ids.has(n.id)) errors.push(`${id}: missing or duplicate node id ${n.id}`);
    ids.add(n.id);
    if (n.tone && !TONES.has(n.tone)) errors.push(`${id}/${n.id}: unknown tone ${n.tone}`);
    if (d.type === "er" && !ER_TAGS[n.tag]) errors.push(`${id}/${n.id}: tag must be one of ${Object.keys(ER_TAGS)}`);
    if (n.group && !(d.groups || []).some((g) => g.id === n.group)) errors.push(`${id}/${n.id}: unknown group ${n.group}`);
    if (n.lines !== undefined && (!Array.isArray(n.lines) || n.lines.length < 1 || n.lines.length > 4 || !n.lines.every((l) => typeof l === "string"))) errors.push(`${id}/${n.id}: lines must be 1-4 strings`);
  }
  for (const e of d.edges || []) {
    if (!ids.has(e.from) || !ids.has(e.to)) errors.push(`${id}: edge ${e.from}→${e.to} references a missing node`);
  }
}

function buildGraph(id, d) {
  const er = d.type === "er";
  const layoutOptions = { ...DEFAULTS, ...(er ? { "elk.layered.spacing.nodeNodeBetweenLayers": "56" } : {}), ...(d.layout || {}) };
  const partitioned = er && d.nodes.some((n) => n.partition !== undefined);
  // MODEL_ORDER cycle breaking crashes inside elkjs 0.12 when partitioning meets a labelled edge.
  if (partitioned) Object.assign(layoutOptions, { "elk.partitioning.activate": "true", "elk.layered.cycleBreaking.strategy": "GREEDY", ...(d.layout || {}) });
  const groups = d.groups || [];
  if (groups.length) layoutOptions["elk.hierarchyHandling"] = "INCLUDE_CHILDREN";

  const nodeOf = new Map();
  const mk = (n) => {
    const size = er ? sizeErNode(n) : sizeFlowNode(n);
    const node = { id: n.id, ...size, _spec: n };
    if (partitioned) node.layoutOptions = { "elk.partitioning.partition": String(n.partition ?? 0) };
    nodeOf.set(n.id, node);
    return node;
  };
  const groupNodes = groups.map((g) => ({
    id: `group:${g.id}`,
    _group: g,
    layoutOptions: { "elk.padding": GROUP_PAD },
    labels: [{ text: g.title, width: tw(g.title, 11, true), height: 14 }],
    children: d.nodes.filter((n) => n.group === g.id).map(mk),
    edges: [],
  }));
  // Root children follow the spec's node order (a group sits where its first member does): MODEL_ORDER
  // cycle breaking reverses edges against that order.
  const children = [];
  for (const n of d.nodes) {
    if (!n.group) children.push(mk(n));
    else if (!children.some((c) => c._group?.id === n.group)) children.push(groupNodes.find((g) => g._group.id === n.group));
  }
  for (const g of groupNodes) if (!children.includes(g)) children.push(g);
  const root = { id, layoutOptions, children, edges: [] };

  d.edges.forEach((e, i) => {
    const labels = [];
    if (e.label) labels.push(label(e.label));
    if (e.tail) labels.push(label(e.tail, { layoutOptions: { "elk.edgeLabels.placement": "TAIL" } }));
    if (e.head) labels.push(label(e.head, { layoutOptions: { "elk.edgeLabels.placement": "HEAD" } }));
    const edge = { id: `e${i}`, sources: [e.from], targets: [e.to], labels, _spec: e };
    const gFrom = nodeOf.get(e.from)._spec.group;
    const gTo = nodeOf.get(e.to)._spec.group;
    if (gFrom && gFrom === gTo) groupNodes.find((g) => g._group.id === gFrom).edges.push(edge);
    else root.edges.push(edge);
  });
  return root;
}

// Flatten ELK output to absolute coordinates; ELK reports children and an edge's points relative to
// the node that contains them.
function flatten(g) {
  const nodes = [];
  const groups = [];
  const edges = [];
  const walk = (n, ox, oy) => {
    for (const c of n.children || []) {
      const abs = { ...c, x: c.x + ox, y: c.y + oy };
      if (c._group) {
        groups.push(abs);
        walk(c, abs.x, abs.y);
      } else nodes.push(abs);
    }
    for (const e of n.edges || []) {
      const shift = (p) => ({ x: p.x + ox, y: p.y + oy });
      const s = e.sections?.[0];
      edges.push({
        ...e,
        points: s ? [s.startPoint, ...(s.bendPoints || []), s.endPoint].map(shift) : null,
        labels: (e.labels || []).map((l) => ({ ...l, x: l.x + ox, y: l.y + oy })),
      });
    }
  };
  walk(g, 0, 0);
  return { nodes, groups, edges };
}

function check(id, g, flat) {
  const byId = Object.fromEntries(flat.nodes.map((n) => [n.id, n]));
  const onBorder = (p, n) => {
    const dx = Math.max(n.x - p.x, 0, p.x - (n.x + n.width));
    const dy = Math.max(n.y - p.y, 0, p.y - (n.y + n.height));
    const inside = p.x > n.x + 1 && p.x < n.x + n.width - 1 && p.y > n.y + 1 && p.y < n.y + n.height - 1;
    return !inside && Math.hypot(dx, dy) <= 2;
  };
  const crosses = (a, b, n) => {
    const [x1, x2] = [Math.min(a.x, b.x), Math.max(a.x, b.x)];
    const [y1, y2] = [Math.min(a.y, b.y), Math.max(a.y, b.y)];
    return x2 > n.x + 2 && x1 < n.x + n.width - 2 && y2 > n.y + 2 && y1 < n.y + n.height - 2;
  };
  for (const e of flat.edges) {
    const name = `${e._spec.from}→${e._spec.to}`;
    if (!e.points) { errors.push(`${id}: ${name} has no route`); continue; }
    if (!onBorder(e.points[0], byId[e.sources[0]])) errors.push(`${id}: ${name} starts off its source`);
    if (!onBorder(e.points.at(-1), byId[e.targets[0]])) errors.push(`${id}: ${name} ends off its target`);
    for (const p of e.points) if (p.x < 0 || p.y < 0 || p.x > g.width || p.y > g.height) errors.push(`${id}: ${name} leaves the viewBox`);
    for (const n of flat.nodes) {
      if (n.id === e.sources[0] || n.id === e.targets[0]) continue;
      for (let i = 1; i < e.points.length; i++) if (crosses(e.points[i - 1], e.points[i], n)) { warnings.push(`${id}: ${name} passes through ${n.id}`); break; }
    }
  }
  for (const n of flat.nodes) if (n.x + n.width > g.width + 0.5 || n.y + n.height > g.height + 0.5) errors.push(`${id}: ${n.id} leaves the viewBox`);
  if (g.width > WIDE) warnings.push(`${id}: ${Math.ceil(g.width)}px wide; rendered full-bleed`);
}

function markerDefs(id, tones) {
  return `<defs>${[...tones].map((t) => `<marker id="elk-${id}-${t}" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse"><path d="M0,0 L10,5 L0,10 z" style="fill:${tone(t)}"/></marker>`).join("")}</defs>`;
}

const DASH = { solid: "", dashed: ";stroke-dasharray:5 4", dotted: ";stroke-dasharray:2 3" };
function renderEdge(id, e, er, future) {
  const s = e._spec;
  const t = s.tone || "muted";
  const arrow = s.arrow ?? !er;
  const d = e.points.map((p, i) => `${i ? "L" : "M"}${p.x.toFixed(1)},${p.y.toFixed(1)}`).join(" ");
  const style = s.style || (future.has(s.to) ? "dotted" : "solid");
  let out = `<path d="${d}" style="fill:none;stroke:${tone(t)};stroke-width:1.5${DASH[style] ?? ""}"${arrow ? ` marker-end="url(#elk-${id}-${t})"` : ""}/>`;
  for (const l of e.labels) {
    out += `<text x="${(l.x + l.width / 2).toFixed(1)}" y="${(l.y + 11).toFixed(1)}" text-anchor="middle" style="fill:${tone(s.tone && s.tone !== "muted" ? s.tone : "muted")};font:10.5px var(--mono)">${esc(l.text)}</text>`;
  }
  return out;
}

function renderFlowNode(n) {
  const s = n._spec;
  const k = tone(s.tone || "a");
  const { x, y, width: w, height: h } = n;
  const ink = s.future ? "var(--muted)" : "var(--ink)";
  const fill = s.future ? 4 : s.strong ? 16 : 10;
  const extra = `${s.strong ? ";stroke-width:2" : ""}${s.future ? ";stroke-dasharray:5 3" : ""}`;
  return `<rect x="${x}" y="${y}" width="${w}" height="${h}" rx="8" style="fill:color-mix(in srgb, ${k} ${fill}%, transparent);stroke:color-mix(in srgb, ${k} ${s.future ? 45 : 65}%, transparent)${extra}"/>` +
    `<text x="${x + w / 2}" y="${y + (s.sub ? 19 : 20)}" text-anchor="middle" style="fill:${ink};font:700 12px var(--mono)">${esc(s.title)}</text>` +
    (s.sub ? `<text x="${x + w / 2}" y="${y + 36}" text-anchor="middle" style="fill:var(--muted);font:10.5px var(--mono)">${esc(s.sub)}</text>` : "") +
    (s.lines || []).map((l, i) => `<text x="${x + PAD}" y="${y + flowTextTop(s) + (i + 1) * LINE_H}" style="fill:var(--muted);font:10.5px var(--mono)">${esc(l)}</text>`).join("");
}

function renderErNode(n) {
  const s = n._spec;
  const k = tone(ER_TAGS[s.tag]);
  const { x, y, width: w, height: h } = n;
  const reg = s.tag === "code-registry";
  const fut = s.tag === "future";
  const fill = reg ? "none" : `color-mix(in srgb, ${k} ${s.tag === "existing" ? 4 : fut ? 3 : 9}%, transparent)`;
  const stroke = `stroke:color-mix(in srgb, ${k} ${fut ? 50 : 70}%, transparent)${s.tag === "new" ? ";stroke-width:2" : ""}${reg || fut ? ";stroke-dasharray:5 3" : ""}`;
  let out = `<rect x="${x}" y="${y}" width="${w}" height="${h}" rx="7" style="fill:${fill};${stroke}"/>`;
  out += `<line x1="${x}" y1="${y + ER_HEAD}" x2="${x + w}" y2="${y + ER_HEAD}" style="stroke:color-mix(in srgb, ${k} 45%, transparent)"/>`;
  out += `<text x="${x + PAD}" y="${y + 16}" style="fill:var(--ink);font:700 11.5px var(--mono)">${esc(s.title)}</text>`;
  out += `<text x="${x + w - PAD}" y="${y + 16}" text-anchor="end" style="fill:var(--muted);font:600 9px var(--mono);letter-spacing:.06em">${esc(s.tag.toUpperCase())}</text>`;
  s.rows.forEach(([name, type, mod], i) => {
    const ry = y + ER_HEAD + 13 + i * ER_ROW;
    if (mod === "new") out += `<rect x="${x + 1}" y="${ry - 11}" width="${w - 2}" height="16" style="fill:color-mix(in srgb, var(--a) 12%, transparent)"/>`;
    const st = mod === "del" ? "fill:var(--e);text-decoration:line-through" : mod === "pk" ? "fill:var(--a)" : "fill:var(--ink)";
    out += `<text x="${x + PAD}" y="${ry}" style="${st};font:${mod === "pk" ? "700 " : ""}10.5px var(--mono)">${esc(name)}</text>`;
    if (type) out += `<text x="${x + w - PAD}" y="${ry}" text-anchor="end" style="fill:var(--muted);font:10.5px var(--mono)">${esc(type)}</text>`;
  });
  return out;
}

function renderGroup(gn) {
  const g = gn._group;
  const k = tone(g.tone || "muted");
  return `<rect x="${gn.x}" y="${gn.y}" width="${gn.width}" height="${gn.height}" rx="10" style="fill:color-mix(in srgb, ${k} 6%, transparent);stroke:color-mix(in srgb, ${k} 40%, transparent)"/>` +
    `<text x="${gn.x + 12}" y="${gn.y + 19}" style="fill:var(--muted);font:600 11px var(--mono)">${esc(g.title)}</text>`;
}

const elk = new ELK();
let current = null;
// elkjs runs layout on a timer, so its exceptions surface as uncaught rather than as a rejection.
process.on("uncaughtException", (err) => fail(`${current ?? "?"}: ELK layout failed: ${err.message}`));
async function render(id, d) {
  const er = d.type === "er";
  current = id;
  const g = await elk.layout(buildGraph(id, d));
  const flat = flatten(g);
  check(id, g, flat);
  const W = Math.ceil(g.width);
  const H = Math.ceil(g.height);
  const tones = new Set(flat.edges.filter((e) => e._spec.arrow ?? !er).map((e) => e._spec.tone || "muted"));
  const future = new Set(d.nodes.filter(isFuture).map((n) => n.id));
  const body = flat.groups.map(renderGroup).join("") +
    flat.edges.filter((e) => e.points).map((e) => renderEdge(id, e, er, future)).join("") +
    flat.nodes.map(er ? renderErNode : renderFlowNode).join("");
  const cls = W > WIDE ? "elk elk-wide" : "elk";
  return `<svg data-elk="${id}" class="${cls}" viewBox="0 0 ${W} ${H}" style="--elk-w:${W}px" role="img" aria-label="${esc(d.label || id)}">${tones.size ? markerDefs(id, tones) : ""}${body}</svg>`;
}

const CSS = `/* sketch-diagrams */ .elk{display:block;width:100%;max-width:var(--elk-w);height:auto;margin:0 auto} .elk-wide{width:min(var(--elk-w),calc(100vw - 32px));max-width:none;position:relative;left:50%;transform:translateX(-50%)}`;

const spec = JSON.parse(readFileSync(specPath, "utf8"));
const diagrams = spec.diagrams || {};
for (const [id, d] of Object.entries(diagrams)) validate(id, d);
if (errors.length) fail(`invalid spec:\n  ${errors.join("\n  ")}`);

let html = readFileSync(page, "utf8");
const rendered = {};
for (const [id, d] of Object.entries(diagrams)) rendered[id] = await render(id, d);

for (const id of Object.keys(diagrams)) {
  const re = new RegExp(`<svg\\b[^>]*\\bdata-elk="${id.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}"[^>]*>[\\s\\S]*?</svg>`);
  if (!re.test(html)) errors.push(`${id}: no <svg data-elk="${id}"> placeholder in the page`);
  else html = html.replace(re, () => rendered[id]);
}
const orphans = [...html.matchAll(/data-elk="([^"]+)"/g)].map((m) => m[1]).filter((id) => !diagrams[id]);
for (const id of orphans) errors.push(`${id}: placeholder has no diagram in ${specPath}`);
if (!html.includes("/* sketch-diagrams */")) html = html.replace("</style>", `  ${CSS}\n</style>`);

for (const w of warnings) console.error(`warn: ${w}`);
if (errors.length) {
  for (const e of errors) console.error(`error: ${e}`);
  process.exit(1);
}
if (!checkOnly) writeFileSync(page, html);
console.log(`${checkOnly ? "checked" : "rendered"} ${Object.keys(diagrams).length} diagram(s) into ${page}`);
