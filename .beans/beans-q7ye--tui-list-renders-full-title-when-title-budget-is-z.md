---
# beans-q7ye
title: TUI list renders full title when title budget is zero
status: todo
type: bug
created_at: 2026-10-06T13:19:37Z
updated_at: 2026-10-06T13:19:37Z
---

`RenderBeanRow` (`internal/ui/styles.go:585-589`) truncates a title only when `MaxTitleWidth > 0`. The list delegate computes the budget as `max(0, m.Width()-baseWidth)` (`internal/tui/list.go:72`). When the ID, status, type and tag columns already fill the row, the budget is 0, and the title is rendered in full.

Example: in the two-column layout at 120x27 the list is 37 columns wide. With tree depth 6 the ID column is 30 wide, `baseWidth` is 40, and the budget is 0. A 41-character title then makes the row 81 columns wide.

The pane clip in `viewContent` (beans-1q77) keeps the layout intact, but cuts such rows hard at the border without `...`.

## Todo

- [ ] Decide what a row shows when the title budget is 0 or less (no title, or a shrunk ID or tags column)
- [ ] Fix `RenderBeanRow` and the budget computation accordingly
- [ ] Table-driven test: deep tree in a narrow pane, every row exactly the list width and ending in `...` or a fully shown title
- [ ] `mise test`
