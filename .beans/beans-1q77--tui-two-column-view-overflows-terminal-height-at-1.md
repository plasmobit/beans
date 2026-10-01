---
# beans-1q77
title: TUI two-column view overflows terminal height at 120x28
status: in-progress
type: bug
priority: normal
created_at: 2026-10-01T09:03:53Z
updated_at: 2026-10-01T09:45:18Z
---

At terminal size 120x28 the two-column TUI renders more lines than the terminal height; the top scrolls out of view. 120x27 and 119x28 render correctly.

- [x] Reproduce with a test rendering App.View() at 120x28
- [x] Identify the overflowing pane/row
- [x] Fix root cause
- [ ] Guard pane heights

## Root cause

`RenderBeanRow` appended the ` ↑<status>` implicit-status annotation after a title already truncated to the full title budget, so annotated rows were 11 columns too wide. In the 37-column left pane at width 120 they wrapped inside the border and grew the pane. Whether a page contains such rows depends on rows-per-page, hence on terminal height.

## Fix

The width fix in `RenderBeanRow` is superseded by beans-5fp6: the marker moved from a title suffix into the status column (`↑T`), so annotated rows have the same width as plain ones. `TestListViewConstrained_ExactSize` guards the pane size.

Open: whether `ViewConstrained` should additionally clip to its height (last todo).
