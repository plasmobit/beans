---
# beans-xg8t
title: RemoveLinksTo and FixBrokenLinks bypass Update's routing and events
status: todo
type: bug
priority: normal
created_at: 2026-10-12T13:59:03Z
updated_at: 2026-10-12T13:59:03Z
---

`Core.RemoveLinksTo` and `Core.FixBrokenLinks` (`pkg/beancore/links.go`) bypass the routing that `Core.Update` applies.

1. **Dirty and worktree-linked beans are written to the main `.beans/`.** Both methods call `saveToDisk` for every bean they change. `Core.Update` instead writes a worktree-linked bean to its worktree and keeps a runtime-only (dirty) bean out of the main directory. Deleting a bean in `beans serve` therefore writes the runtime state of every dirty bean that linked to it to main.
2. **No subscriber events.** Neither method calls `fanOut`, so the Beans UI keeps showing the removed links until a watcher reload.

## Todo

- [ ] Route each changed bean like `Core.Update` does (worktree link, dirty flag, persist), or call `Update` per bean
- [ ] Send `EventUpdated` for every changed bean, after the lock is released
- [ ] Tests: a dirty bean and a worktree-linked bean that link to a deleted bean; a subscriber receives the update
- [ ] `go test ./...`
