---
# beans-6we9
title: Core methods modify stored beans in place
status: todo
type: bug
created_at: 2026-10-08T15:40:29Z
updated_at: 2026-10-08T15:40:29Z
---

Five `Core` methods modify beans stored in `c.beans` in place, under the write lock. `Core.Get` returns the stored pointer and releases its read lock, so a reader that holds that pointer reads the same fields without a lock. The two accesses race.

| Method | Writes |
|---|---|
| `Core.Archive` | `Path` |
| `Core.Unarchive` | `Path` |
| `Core.LoadAndUnarchive` | `Path` |
| `Core.RemoveLinksTo` (`pkg/beancore/links.go`) | `Parent`, `Blocking`, `BlockedBy` |
| `Core.FixBrokenLinks` (`pkg/beancore/links.go`) | `Parent`, `Blocking`, `BlockedBy` |

Readers without a lock: `CoreResolver.cloneForUpdate` (`pkg/beangraph/resolver.go`), the GraphQL field resolvers, `output.NewBean`.

Example: `deleteBean(Y)` and `updateBean(X)` run concurrently, and X blocks Y. `RemoveLinksTo` reassigns `X.Blocking` while `Bean.Clone` runs `slices.Clone(X.Blocking)`. The race detector flags it, and a torn slice header can read out of bounds.

beans-o1if made `Core.Update` a compare-and-swap with the invariant "no code modifies a bean stored in `c.beans`". These five methods are the remaining exceptions.

## Cross-bean cycle race

The relationship validators run `DetectCycleWith` before `Core.Update` takes the write lock. The compare-and-swap protects one bean only. Request A adds `X blocking Y` on X, request B adds `Y blocking X` on Y, concurrently: each check sees only its own pending edge, both swaps succeed, and a blocking cycle is stored.

## Todo

- [ ] The five methods clone each bean they change, modify the clone, and store the clone in `c.beans`
- [ ] Test under `-race`: `RemoveLinksTo` concurrent with `Clone` of a bean it changes
- [ ] Decide the cycle race: re-run the cycle check under the write lock (e.g. a validator callback in `Core.Update`), or accept and document it
- [ ] `mise test`
