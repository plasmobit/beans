---
# beans-p17z
title: beans next command
status: todo
type: task
priority: normal
created_at: 2025-12-27T21:44:04Z
updated_at: 2026-10-14T13:34:41Z
order: VVg
parent: beans-mmyp
---

Add `beans next`: a ranked list of beans to start next, with the lane and one reason line per row.

## Candidates

The `--ready` set (`beangraph.AddReadyFilter`), restricted to workable beans: no epics, no milestones, no beans with open children.

## Fronts

A front is a milestone or epic whose children get focus.

- **Explicit focus** (`next.focus` set): the listed IDs are the fronts and replace the in-progress ones. `<global>` makes root beans (no parent) a front.
- **Otherwise:** every in-progress milestone or epic is a front.
- The nearest front ancestor of a bean decides (an in-progress epic under an in-progress milestone uses the epic).

## Lanes

Each candidate goes into the first lane that matches:

1. **interrupt** – bug with priority high or critical, anywhere in the tree
2. **focus** – bean has a front among its ancestors
3. **rest** – shown only with `--all`

## Ordering inside a lane

1. front priority
2. front manual `order`
3. front `updated_at`, most recent first
4. bean priority
5. bean type: bug, feature, task
6. bean manual `order`
7. bean `created_at`, oldest first

Steps 1–3 apply only to beans that have a front. Each level uses one attribute: the priority shown on a bean is the priority used for it.

## Output

- Header line names the scope source, e.g. `scope: focus (.beans.local.yml)` or `scope: in-progress fronts`.
- One row per bean: lane, ID, title, reason (e.g. `focus: under epic beans-5txp (in-progress)`).
- Hint when more than one in-progress front exists and no focus is set: `3 active fronts – consider "beans focus set"`.
- If nothing is ready: suggest `beans list --is-blocked` or `beans list`.

Flags: `--limit N`, `--json`, `--all`, `--in-progress` (lists in-progress leaf beans to resume, same format).

## Focus setting

- Key `next.focus: [<id>|<global>, ...]`; each ID must be a milestone or epic.
- Read from `.beans.yml` and `.beans.local.yml`; the local value replaces the shared one.
- `.beans.local.yml` is per user and gitignored; `beans init` adds it to `.gitignore`.
- CLI writes only `.beans.local.yml`: `beans focus` (show, with source), `beans focus set <id>|<global>`, `beans focus clear`.

## Layout

- `pkg/bean/next.go` – ranking as a pure function (no I/O)
- `pkg/config` – merge `.beans.local.yml` over `.beans.yml`
- `internal/commands/next.go`, `internal/commands/focus.go` – CLI, calling the core resolver directly like `beans list`

## Todo

- [ ] Ranking in `pkg/bean/next.go` with table-driven tests (lanes, front selection, ordering, nested fronts)
- [ ] Config: `next.focus` and `.beans.local.yml` merge, with tests
- [ ] `beans init` adds `.beans.local.yml` to `.gitignore`
- [ ] `beans next` command with flags, scope header and multi-front hint
- [ ] `beans focus` / `focus set` / `focus clear`
- [ ] Manual check on this repo: the 3 high bugs under `beans-5txp` rank first

## Out of scope

TUI focus key and GraphQL exposure: see follow-up bean.
