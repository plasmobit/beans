---
# beans-4vu3
title: Agent tests spawn the real Claude CLI instead of a mock
status: todo
type: bug
priority: critical
created_at: 2026-10-09T15:34:32Z
updated_at: 2026-10-09T15:44:13Z
blocking:
    - beans-eudm
---

The `internal/agent` tests start the real Claude Code CLI whenever a `claude` binary is on `PATH`. The tests assume none is installed ("SendMessage will try to spawn a process, which will fail since there's no claude binary in test", `describe_test.go:70`, `manager_test.go:464`). On a developer machine with Claude Code installed, that assumption is false.

## Observed

One run of `go test ./internal/agent/` started 10 `claude` processes:

- 6 × `claude --print --model haiku ...` with the working directory `internal/agent` inside this repository. They ran `git status`, `git log`, `git config user.name` there.
- 1 × an agent session with `--dangerously-skip-permissions` in `/tmp`.
- 3 × agent sessions in plan mode in `/tmp`.

The processes inherit the full environment (`buildClaudeEnv` removes only `CLAUDECODE`). Under `git rebase --exec` that includes `GIT_DIR` pointing at the real repository. They receive test prompts such as "Fix the auth bug" and "proceed" and may make real API calls.

## Spawn sites

- `internal/agent/claude.go:118`: `spawnAndRun`, reached from `Manager.SendMessage` (`manager.go:239`) and `claude.go:792`
- `internal/agent/quickreplies.go:49`: `generateQuickReplies`, reached from the output parser on turn completion (`claude.go:561`); sets no `cmd.Dir`
- `internal/agent/describe.go:53`: workspace description generation

## Fix

Tests must never execute a real agent. Two candidate designs, to be decided:

1. Make the agent command replaceable on `Manager` (default `claude`); tests set a stub.
2. `TestMain` puts a directory with a stub `claude` first on `PATH`; no production change.

- [ ] Decide between the two designs
- [ ] Introduce the agent mock and route all three spawn sites through it
- [ ] Guard test: running the package's tests starts no process named by the real `claude` path
- [ ] Update the test comments that assume a missing binary
