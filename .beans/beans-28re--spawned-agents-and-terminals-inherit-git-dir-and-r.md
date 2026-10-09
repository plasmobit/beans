---
# beans-28re
title: Spawned agents and terminals inherit GIT_DIR and related variables from beans serve
status: todo
type: bug
priority: high
created_at: 2026-10-09T14:49:33Z
updated_at: 2026-10-09T14:49:33Z
blocked_by:
    - beans-eudm
---

`beans serve` passes its own environment to the processes it spawns: Claude agents (`internal/agent/claude.go:119`), terminal PTYs (`internal/terminal/terminal.go:286`, `:340`), and the git commands in `internal/worktree`. They select their repository through `cmd.Dir`. If the server starts with `GIT_DIR`, `GIT_WORK_TREE`, `GIT_INDEX_FILE` or another variable from `git rev-parse --local-env-vars` set (for example from a hook, `git rebase --exec` or a `!`-alias), an agent working in a worktree acts on that repository instead.

- [ ] Remove the `--local-env-vars` variables from the environment of spawned agents, terminals and git commands
- [ ] Test: with `GIT_DIR` set, a spawned process's `git rev-parse --absolute-git-dir` reports the worktree
