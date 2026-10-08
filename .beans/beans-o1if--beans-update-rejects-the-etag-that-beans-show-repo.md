---
# beans-o1if
title: beans update rejects the etag that beans show reports
status: todo
type: bug
priority: normal
created_at: 2026-10-08T13:51:47Z
updated_at: 2026-10-08T13:52:17Z
---

`beans update --if-match` rejects the etag that `beans show` reports, for every bean whose file differs from its rendered form. The cause is that `updateBean` modifies the stored bean before the etag check, and `Core.Update` works around that by hashing a different input than `Bean.ETag()`.

## Reproduction

A bean created without `-p` has no `priority:` line in its file:

```
$ beans create "probe" -t task          # -> beans-xxxx
$ beans update beans-xxxx --if-match "$(beans show beans-xxxx --etag-only)" -s todo
Error: etag mismatch: provided a685f981b655e800, current is 79b12faf276b6c8d
```

Observed on beans-4x0o (2026-10-07) and beans-q7ye (2026-10-06).

## Root cause

Two generators compute the etag of one bean:

| Path | Generator |
|---|---|
| reads: `Bean.ETag()` (`pkg/bean/bean.go:301`), used by `show`, `--json`, GraphQL | FNV-1a of `Render()` |
| `If-Match` check in `Core.Update` (`pkg/beancore/core.go:563-574`), clean beans | FNV-1a of the file bytes |

The loader fills defaults (`pkg/beancore/core.go:205-210`, e.g. `priority: normal`), so the rendered form and the file differ.

The file-byte hash exists because of the order in `UpdateBean` (`pkg/beangraph/mutations.go:109` onward):

1. `Core.Get` returns the pointer stored in `c.beans[id]` (`core.go:315`), not a copy.
2. `UpdateBean` writes the input fields into that object, without a lock.
3. `Core.Update` takes the write lock and only then compares the etag. `storedBean` is the modified object, so its `ETag()` can never match.

The check ran before the modification originally (PR #59, `validateETag` in `pkg/beangraph/resolver.go:40`). PR #72 moved it under the write lock to fix a race (two requests both passing the check, the second overwriting the first), which put it after the modification. `validateETag` has had no callers since; step 2 of the design reuses it.

Further effects of the same order, inferred from the code, not reproduced:

- The `If-Match` check on a dirty bean always fails: `core.go:576-579` hashes the already-modified stored object.
- A rejected update still changes `c.beans[id]`. In `beans serve`, queries and subscriptions show the rejected change until the bean is reloaded.
- Step 2 races with readers that hold `RLock`.

## Design

Optimistic concurrency with a compare-and-swap as the only transaction:

1. **Read and clone.** Deep-copy `c.beans[id]`, including `Tags`, `Blocking`, `BlockedBy`, `Properties`. `base := clone.ETag()`.
2. **Early reject.** If `ifMatch` is given and differs from `base`, return `ETagMismatchError`. Nothing has changed.
3. **Modify the clone.** No lock is held, so validations such as `DetectCycle` can call `Core`.
4. **Compare-and-swap.** `Core.Update(clone, base)` under the write lock: write only if `c.beans[id].ETag() == base`; save to disk (or worktree, or mark dirty) and store the clone in `c.beans[id]`. Otherwise return `ETagMismatchError`.

Step 4 compares against `base` also when the client sent no `ifMatch`, so concurrent updates without an etag no longer overwrite each other.

Invariant: no code modifies a bean stored in `c.beans`. Every writer uses clone, modify, swap.

Generator: both sides of step 4 come from the in-memory bean, so `Bean.ETag()` (rendered hash) serves reads and the check, and the file-byte branch goes away. It is weak with respect to the file: a hand edit that renders identically (whitespace, field order, a missing default) does not change the etag, and a hand edit stays invisible until the file watcher reloads the bean. The alternative closes both gaps: `beancore` records the hash of the bytes it last read or wrote, and a dirty bean uses its rendered hash.

## Todo

- [ ] Decide the generator: rendered hash everywhere, or recorded file hash
- [ ] Add a deep `Clone()` to `bean.Bean` with a test that mutating the clone leaves the original unchanged
- [ ] `Core.Get` returns a clone, or its callers are audited to treat the result as read-only
- [ ] Every caller of `Core.Update` clones, rejects early, and modifies the clone: `UpdateBean` and the five relationship mutations (`pkg/beangraph/mutations.go:236,306,340,356,389,405`), and the TUI (`internal/tui/tui.go:658`)
- [ ] `Core.Update`: compare-and-swap against `base` under the write lock; remove the file-byte branch (or replace it by the recorded hash)
- [ ] Step 2 (early reject) uses `validateETag` (`pkg/beangraph/resolver.go:40`), as `.claude/rules/backend.md` asks of resolvers
- [ ] Tests: `show` etag accepted by `update` for a file without `priority:`; dirty bean with `If-Match`; rejected update leaves the stored bean unchanged; two concurrent updates with the same `base`, exactly one succeeds
- [ ] `mise test`
