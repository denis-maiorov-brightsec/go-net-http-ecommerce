#!/usr/bin/env node

import { existsSync, mkdirSync, readdirSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { spawnSync } from "node:child_process";

const repoRoot = process.cwd();
const defaultIndexPath = resolve(repoRoot, "docs/SPECS_INDEX.md");
const defaultSpecsDir = resolve(repoRoot, "docs/specs");
const defaultStackProfilePath = resolve(repoRoot, "docs/STACK_PROFILE.md");
const defaultImplementerPromptPath = resolve(repoRoot, "prompts/02-implement-next-ready-spec.md");
const defaultReviewerPromptPath = resolve(repoRoot, "prompts/03-review-and-fix-last-spec.md");

function printUsage() {
  console.log(`Usage: node scripts/run-specs-harness.mjs [options]

Runs specs sequentially with two Codex passes per spec:
1) implementer agent
2) reviewer/fixer agent

Options:
  --dry-run                 Show planned execution order only
  --max-specs <n>           Process at most n specs in this run
  --from <id>               Start from spec id (e.g. 003)
  --to <id>                 Stop at spec id (inclusive)
  --model <name>            Pass model to codex exec
  --unsafe                  Use --dangerously-bypass-approvals-and-sandbox
  --allow-dirty             Allow starting with dirty git working tree
  --allow-non-main          Allow running from a branch other than main
  --codex-bin <path>        Codex binary (default: codex)
  --index-path <path>       Specs index path (default: docs/SPECS_INDEX.md)
  --specs-dir <path>        Specs directory path (default: docs/specs)
  --stack-profile <path>    Stack profile path (default: docs/STACK_PROFILE.md)
  --implementer-prompt <path>
                            Implementer prompt path (default: prompts/02-implement-next-ready-spec.md)
  --reviewer-prompt <path>  Reviewer prompt path (default: prompts/03-review-and-fix-last-spec.md)

  -h, --help                Show this help
`);
}

function parseArgs(argv) {
  const args = {
    dryRun: false,
    maxSpecs: Number.POSITIVE_INFINITY,
    from: undefined,
    to: undefined,
    model: undefined,
    unsafe: false,
    allowDirty: false,
    allowNonMain: false,
    codexBin: "codex",
    indexPath: defaultIndexPath,
    specsDir: defaultSpecsDir,
    stackProfilePath: defaultStackProfilePath,
    implementerPromptPath: defaultImplementerPromptPath,
    reviewerPromptPath: defaultReviewerPromptPath,
  };

  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];

    if (arg === "--") {
      continue;
    }

    if (arg === "--dry-run") {
      args.dryRun = true;
      continue;
    }

    if (arg === "--unsafe") {
      args.unsafe = true;
      continue;
    }

    if (arg === "--allow-dirty") {
      args.allowDirty = true;
      continue;
    }

    if (arg === "--allow-non-main") {
      args.allowNonMain = true;
      continue;
    }

    if (arg === "--max-specs") {
      args.maxSpecs = Number(argv[i + 1]);
      i += 1;
      continue;
    }

    if (arg === "--from") {
      args.from = normalizeSpecId(argv[i + 1]);
      i += 1;
      continue;
    }

    if (arg === "--to") {
      args.to = normalizeSpecId(argv[i + 1]);
      i += 1;
      continue;
    }

    if (arg === "--model") {
      args.model = argv[i + 1];
      i += 1;
      continue;
    }

    if (arg === "--codex-bin") {
      args.codexBin = argv[i + 1];
      i += 1;
      continue;
    }

    if (arg === "--index-path") {
      args.indexPath = resolve(repoRoot, argv[i + 1]);
      i += 1;
      continue;
    }

    if (arg === "--specs-dir") {
      args.specsDir = resolve(repoRoot, argv[i + 1]);
      i += 1;
      continue;
    }

    if (arg === "--stack-profile") {
      args.stackProfilePath = resolve(repoRoot, argv[i + 1]);
      i += 1;
      continue;
    }

    if (arg === "--implementer-prompt") {
      args.implementerPromptPath = resolve(repoRoot, argv[i + 1]);
      i += 1;
      continue;
    }

    if (arg === "--reviewer-prompt") {
      args.reviewerPromptPath = resolve(repoRoot, argv[i + 1]);
      i += 1;
      continue;
    }

    if (arg === "-h" || arg === "--help") {
      printUsage();
      process.exit(0);
    }

    throw new Error(`Unknown argument: ${arg}`);
  }

  if (args.maxSpecs !== Number.POSITIVE_INFINITY && (!Number.isFinite(args.maxSpecs) || args.maxSpecs <= 0)) {
    throw new Error("--max-specs must be a positive number");
  }

  if (args.from && args.to && Number(args.from) > Number(args.to)) {
    throw new Error("--from must be less than or equal to --to");
  }

  return args;
}

function normalizeSpecId(value) {
  if (!value) {
    throw new Error("Missing spec id value");
  }

  const trimmed = value.trim();
  const numeric = Number(trimmed);
  if (!Number.isInteger(numeric) || numeric <= 0) {
    throw new Error(`Invalid spec id: ${value}`);
  }

  return String(numeric).padStart(3, "0");
}

function runCommand(command, args, options = {}) {
  const result = spawnSync(command, args, {
    cwd: repoRoot,
    encoding: "utf8",
    stdio: options.stdio ?? "pipe",
    input: options.input,
  });

  if (result.error) {
    throw result.error;
  }

  return result;
}

function ensureGitCleanOrThrow(allowDirty) {
  if (allowDirty) {
    return;
  }

  const status = runCommand("git", ["status", "--porcelain"]);
  if (status.status !== 0) {
    throw new Error(`Failed to check git status:\n${status.stderr}`);
  }

  if (status.stdout.trim() !== "") {
    throw new Error("Git working tree is not clean. Commit/stash changes or pass --allow-dirty.");
  }
}

function getHeadSha() {
  const head = runCommand("git", ["rev-parse", "HEAD"]);
  if (head.status !== 0) {
    throw new Error(`Failed to read HEAD:\n${head.stderr}`);
  }

  return head.stdout.trim();
}

function getCurrentBranch() {
  const branch = runCommand("git", ["branch", "--show-current"]);
  if (branch.status !== 0) {
    throw new Error(`Failed to determine current branch:\n${branch.stderr}`);
  }

  return branch.stdout.trim();
}

function parseSpecsIndex(indexPath, specsDir) {
  const content = readFileSync(indexPath, "utf8");
  const lines = content.split(/\r?\n/);
  const specs = [];

  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed.startsWith("|")) {
      continue;
    }

    const columns = trimmed.split("|").slice(1, -1).map(value => value.trim());
    if (columns.length < 4) {
      continue;
    }

    if (!/^\d{3}$/.test(columns[0])) {
      continue;
    }

    const [id, title, depsRaw, status] = columns;
    const deps = depsRaw === "-" ? [] : depsRaw.split(",").map(dep => dep.trim()).filter(Boolean);

    specs.push({
      id,
      title,
      deps,
      status,
    });
  }

  if (specs.length === 0) {
    throw new Error(`Could not parse any specs from ${indexPath}`);
  }

  const specFiles = readdirSync(specsDir);
  for (const spec of specs) {
    const file = specFiles.find(name => name.startsWith(`${spec.id}-`) && name.endsWith(".md"));
    if (!file) {
      throw new Error(`Missing spec file for ${spec.id} in ${specsDir}`);
    }

    spec.file = file;
    spec.path = resolve(specsDir, file).replace(`${repoRoot}/`, "");
  }

  return specs;
}

function readStackProfile(stackProfilePath) {
  if (!existsSync(stackProfilePath)) {
    return "";
  }

  return readFileSync(stackProfilePath, "utf8").trim();
}

function readPromptFile(promptPath, label) {
  if (!existsSync(promptPath)) {
    throw new Error(`Missing ${label} prompt: ${promptPath}`);
  }

  return readFileSync(promptPath, "utf8").trim();
}

function buildStackContext(stackProfilePath, stackProfile) {
  const lines = [
    "Stack context:",
    `- Stack profile path: ${stackProfile ? stackProfilePath.replace(`${repoRoot}/`, "") : "<missing>"}`,
  ];

  if (!stackProfile) {
    lines.push("- Stack profile content: <missing>");
    return lines.join("\n");
  }

  lines.push("", "Stack profile content:", stackProfile);
  return lines.join("\n");
}

function buildImplementerPrompt(spec, stackContext, basePrompt) {
  return `
${basePrompt}

Explicit target for this run:
- Spec ID: ${spec.id}
- Spec title: ${spec.title}
- Spec file: ${spec.path}

Run-specific constraints:
1. Work on this target spec only.
2. Do not pick another Ready spec.
3. Use AGENTS.md commit style with spec id ${spec.id}.
4. Keep commits small and reviewable.
5. Stay on the current branch and do not create or switch branches.

${stackContext}
`.trim();
}

function buildReviewerPrompt(spec, stackContext, basePrompt) {
  return `
${basePrompt}

Explicit review target for this run:
- Spec ID: ${spec.id}
- Spec title: ${spec.title}
- Spec file: ${spec.path}

Run-specific constraints:
1. Review only this target spec implementation.
2. Do not start implementing a different Ready spec.
3. If fixes are made, use AGENTS.md commit style with spec id ${spec.id}.
4. Stay on the current branch and do not create or switch branches.

${stackContext}
`.trim();
}

function runCodexExec({ codexBin, model, unsafe, prompt, outputFile }) {
  const args = ["exec"];

  if (unsafe) {
    args.push("--dangerously-bypass-approvals-and-sandbox");
  }
  else {
    args.push("--full-auto");
  }

  if (model) {
    args.push("--model", model);
  }

  args.push("--cd", repoRoot, "--output-last-message", outputFile, "-");

  const result = runCommand(codexBin, args, {
    stdio: ["pipe", "inherit", "inherit"],
    input: prompt,
  });
  return result.status ?? 1;
}

function inRange(spec, from, to) {
  if (from && Number(spec.id) < Number(from)) {
    return false;
  }

  if (to && Number(spec.id) > Number(to)) {
    return false;
  }

  return true;
}

function computeNextSpec(specs, doneSet, from, to) {
  return specs.find(spec =>
    inRange(spec, from, to)
    && !doneSet.has(spec.id)
    && spec.deps.every(dep => doneSet.has(dep)));
}

function getPendingInRange(specs, doneSet, from, to) {
  return specs.filter(spec => inRange(spec, from, to) && !doneSet.has(spec.id));
}

function main() {
  const args = parseArgs(process.argv.slice(2));

  if (!existsSync(args.indexPath)) {
    throw new Error(`Missing specs index: ${args.indexPath}`);
  }

  if (!existsSync(args.specsDir)) {
    throw new Error(`Missing specs directory: ${args.specsDir}`);
  }

  ensureGitCleanOrThrow(args.allowDirty);
  const currentBranch = getCurrentBranch();

  if (!args.allowNonMain && currentBranch !== "main") {
    throw new Error(`Harness must run on 'main'. Current branch: '${currentBranch}'. Use --allow-non-main to override.`);
  }

  const specs = parseSpecsIndex(args.indexPath, args.specsDir);
  const doneSet = new Set(specs.filter(spec => spec.status.toLowerCase() === "done").map(spec => spec.id));
  const stackProfile = readStackProfile(args.stackProfilePath);
  const implementerPromptBase = readPromptFile(args.implementerPromptPath, "implementer");
  const reviewerPromptBase = readPromptFile(args.reviewerPromptPath, "reviewer");
  const stackContext = buildStackContext(args.stackProfilePath, stackProfile);

  const runId = new Date().toISOString().replaceAll(":", "-");
  const runDir = resolve(repoRoot, ".codex-runs", runId);
  mkdirSync(runDir, { recursive: true });

  let processed = 0;

  while (processed < args.maxSpecs) {
    const next = computeNextSpec(specs, doneSet, args.from, args.to);
    if (!next) {
      break;
    }

    processed += 1;
    console.log(`\n=== [${processed}] Spec ${next.id}: ${next.title} ===`);

    if (args.dryRun) {
      doneSet.add(next.id);
      continue;
    }

    ensureGitCleanOrThrow(args.allowDirty);

    const beforeImpl = getHeadSha();
    const implStatus = runCodexExec({
      codexBin: args.codexBin,
      model: args.model,
      unsafe: args.unsafe,
      prompt: buildImplementerPrompt(next, stackContext, implementerPromptBase),
      outputFile: resolve(runDir, `${next.id}-implementer.md`),
    });

    if (implStatus !== 0) {
      throw new Error(`Implementer agent failed for spec ${next.id}`);
    }

    const afterImpl = getHeadSha();
    if (afterImpl === beforeImpl) {
      throw new Error(`Implementer agent did not create a commit for spec ${next.id}`);
    }

    const dirtyAfterImpl = runCommand("git", ["status", "--porcelain"]);
    if (dirtyAfterImpl.status !== 0) {
      throw new Error(`Failed to read git status after implementer stage for ${next.id}`);
    }
    if (dirtyAfterImpl.stdout.trim() !== "") {
      throw new Error(`Working tree is dirty after implementer stage for ${next.id}`);
    }

    const beforeReview = getHeadSha();
    const reviewStatus = runCodexExec({
      codexBin: args.codexBin,
      model: args.model,
      unsafe: args.unsafe,
      prompt: buildReviewerPrompt(next, stackContext, reviewerPromptBase),
      outputFile: resolve(runDir, `${next.id}-reviewer.md`),
    });

    if (reviewStatus !== 0) {
      throw new Error(`Reviewer agent failed for spec ${next.id}`);
    }

    const afterReview = getHeadSha();
    if (afterReview !== beforeReview) {
      console.log(`Reviewer committed follow-up fixes for spec ${next.id}.`);
    }
    else {
      console.log(`Reviewer made no commit for spec ${next.id}.`);
    }

    const dirtyAfterReview = runCommand("git", ["status", "--porcelain"]);
    if (dirtyAfterReview.status !== 0) {
      throw new Error(`Failed to read git status after reviewer stage for ${next.id}`);
    }
    if (dirtyAfterReview.stdout.trim() !== "") {
      throw new Error(`Working tree is dirty after reviewer stage for ${next.id}`);
    }

    doneSet.add(next.id);
  }

  const pending = getPendingInRange(specs, doneSet, args.from, args.to);

  console.log("\n=== Run Summary ===");
  console.log(`Processed specs: ${processed}`);
  console.log(`Run logs: ${resolve(".codex-runs", runId)}`);

  if (pending.length === 0) {
    console.log("All eligible specs in range are marked complete for this run.");
    return;
  }

  console.log("Remaining specs not processed:");
  for (const spec of pending) {
    const missing = spec.deps.filter(dep => !doneSet.has(dep));
    const reason = missing.length > 0 ? `blocked by: ${missing.join(", ")}` : "not selected";
    console.log(`- ${spec.id} ${spec.title} (${reason})`);
  }
}

try {
  main();
}
catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  console.error(`Harness failed: ${message}`);
  process.exit(1);
}
