---
name: merge-from-upstream
description: Merge upstream (tomasz-tomczyk/crit) into this crit fork — merge upstream/main into local main, commit and push to origin (talbor49/crit), then rebuild the installed binary at ~/.local/bin/crit. Use when the user says "merge from upstream", "sync the fork", "update from upstream", "pull upstream changes", "merge upstream and push", or "update and rebuild crit".
---

# Merge from upstream

Fork layout: `origin` = `talbor49/crit`, upstream = `tomasz-tomczyk/crit`, both default to `main`.
Merge strategy, never rebase — history is not rewritten, so `push` never needs `--force`.

Run the steps in order. Stop and report at the first failure; do not improvise past it.

## 1. Ensure the upstream remote

```bash
git remote get-url upstream 2>/dev/null || git remote add upstream https://github.com/tomasz-tomczyk/crit.git
```

## 2. Handle uncommitted work — ask, never assume

Several sessions edit this repo at once, so uncommitted changes are usually someone else's in-flight work.

```bash
git status --porcelain
git diff --stat HEAD
```

Show the user what's there and ask what to commit and with what message. Wait for the answer.
Never `git add -A` on your own initiative, never commit or stash a change you didn't make.

If they want it committed, commit with a one-line message and no `Co-Authored-By` line.
If they want it left alone, continue — a merge that doesn't touch those files succeeds fine on a dirty tree.

## 3. Merge upstream

Must be on `main`; if not, say so and stop.

```bash
git fetch upstream
git merge --no-edit upstream/main
```

On conflicts: report the conflicted paths and stop. Resolving another session's conflicts unprompted is worse than waiting.
If the merge refuses because a local modification would be overwritten, report which files and stop — that's step 2 coming back.

## 4. Push to the fork

```bash
git push origin main
```

## 5. Rebuild the installed binary

`crit` is installed as a plain binary at `~/.local/bin/crit` (first on PATH). It does not auto-update.

```bash
go build -o ~/.local/bin/crit ./cmd/crit
crit stop
crit --version
```

`crit stop` matters: a running daemon holds the old binary, so without it the next `crit` serves pre-merge code.

If the build fails, report the compile error verbatim — the merge and push already landed, so say plainly that the
installed binary is still the old one.

## Report

One block: upstream commits merged (`git log --oneline HEAD@{1}..HEAD`), what was committed and pushed, new version
from `crit --version`. Note anything skipped and why.
