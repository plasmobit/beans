---
# beans-qypq
title: Bean writes are not coordinated across processes
status: todo
type: bug
priority: high
created_at: 2026-10-12T08:42:53Z
updated_at: 2026-10-12T08:43:00Z
---

Bean writes are protected by `Core.mu`, a `sync.RWMutex`. It orders goroutines inside one process only. Several processes write the same `.beans/` files, and none of them coordinates with the others: there is no file lock and no single writer process. A bean can therefore change on disk between the moment a process reads it and the moment it writes it, and the compare-and-swap in `Core.Update` (beans-o1if) does not notice.

## Who writes the files

| Writer | Own `Core` | Learns of other writers by |
|---|---|---|
| every CLI call | yes: `beancore.New` + `Load()` in `internal/commands/root.go:56` | nothing; it loads once at start |
| `beans serve` | yes, long-lived | file watcher, 100 ms debounce (`pkg/beancore/watcher.go:15`) |
| TUI | yes | `core.Load()` after the editor closes |
| editors, `git checkout`/`rebase`/`merge`, worktree agents | no | — |

The CLI calls `CoreResolver` in its own process; it does not send mutations to a running `beans serve`.

`Core.Update` compares `base` with the bean in its own memory, never with the file. `saveToDisk` writes with `os.WriteFile` (`pkg/beancore/core.go:618`), which truncates and rewrites the file in place.

## Drift cases

Inferred from the code, not reproduced.

1. **Two CLI processes.** A and B load bean X. Each compares against its own load, both swaps succeed, the second write wins. `--if-match` does not help: the check runs against the load, not the file.
2. **CLI and `beans serve`.** The CLI writes X. Before the watcher reloads X, a UI mutation's compare-and-swap succeeds against the stale in-memory X and overwrites the CLI's change on disk.
3. **Torn reads.** The watcher, or another process's `Load()`, reads a file while `os.WriteFile` is still writing it. Parsing fails or reads partial content.
4. **Hand edits and git.** These ignore any lock beans could take. They can only be detected.

## Design

All of it lives in `beancore`, so the CLI, the TUI and `beans serve` get it alike.

1. **Lock.** Every bean write takes an exclusive `flock` on `.beans/.lock`. One lock per `.beans/` directory, not per file, so that invariants spanning beans hold too: the cycle check (beans-6we9), `RemoveLinksTo`, `FixBrokenLinks`.
2. **Compare against the file.** Under the lock, re-read the bean file and compare its ETag with `base`. On mismatch return `ETagMismatchError` and refresh the in-memory bean from the file.
3. **Atomic write.** Write a temp file in the same directory, then `rename` it over the bean file.
4. **Update memory, then release the lock.**

Limits:

- `flock` is advisory. Editors and git ignore it; step 2 detects their changes but cannot prevent them.
- A worktree has its own `.beans/`, hence its own lock. `beans serve`'s merged view stays as consistent as the worktree watcher keeps it.
- `flock` is unreliable on some network filesystems.

Alternative: `beans serve` as the single writer, the CLI sending its mutations over GraphQL (as `CLAUDE.md` describes). It needs a running server, and it still does not cover hand edits and git, so it does not replace the lock.

## Todo

- [ ] Reproduce case 1 with two concurrent CLI processes, and case 3 with a reader racing `saveToDisk`
- [ ] Decide the lock granularity: one lock per `.beans/` (proposed) or one per bean file
- [ ] Decide what the ETag of step 2 hashes: the rendered form of the re-read bean, or the file bytes (beans-o1if chose the rendered form for reads)
- [ ] `saveToDisk` and `saveToWorktree` write atomically (temp file + `rename`)
- [ ] `Core.Update` takes the lock and compares against the file
- [ ] `Create`, `Delete`, `Archive`, `Unarchive`, `LoadAndUnarchive`, `RemoveLinksTo`, `FixBrokenLinks`, `SaveDirty`, `SaveBean` take the lock
- [ ] Tests: two `Core`s on one directory, concurrent updates with the same `base`, exactly one succeeds; reader never sees a partial file
- [ ] `go test -race ./...`
