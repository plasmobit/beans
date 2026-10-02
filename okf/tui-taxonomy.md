# TUI taxonomy

This file lists the names of the terminal UI's elements, as the code uses them. It covers the
Bubbletea TUI (`beans-tui`, `internal/tui`, `internal/ui`), not the Beans UI in the browser. Each
name gives the code location that defines it.

## Views

A view fills the terminal. `App.state` (type `viewState`, `internal/tui/tui.go`) holds the active one.

| Name | Code | Opened by |
|---|---|---|
| list view | `viewList`, `listModel` (`list.go`) | start; `esc` from the detail view |
| detail view | `viewDetail`, `detailModel` (`detail.go`) | `enter` on a bean |
| tag picker | `viewTagPicker`, `tagPickerModel` | `g t` |
| parent picker | `viewParentPicker`, `parentPickerModel` | `p` |
| status picker | `viewStatusPicker`, `statusPickerModel` | `s` |
| type picker | `viewTypePicker`, `typePickerModel` | `t` |
| priority picker | `viewPriorityPicker`, `priorityPickerModel` | `P` |
| blocking picker | `viewBlockingPicker`, `blockingPickerModel` | `b` |
| create modal | `viewCreateModal`, `createModalModel` | `c` |
| help overlay | `viewHelpOverlay`, `helpOverlayModel` (`help.go`) | `?` |

The tag picker fills the terminal. The other pickers, the create modal and the help overlay are
**modals**: `ModalView` draws them over the view they were opened from (`App.previousState`).

## Layouts of the list view

| Name | Code | Arrangement |
|---|---|---|
| two-column layout | `isTwoColumnMode`, `renderTwoColumnView` | list pane left, preview pane right |
| stacked layout | `isStackedMode`, `renderStackedView` | preview pane below the list pane |
| single-column layout | neither of the above | list pane alone |

**Preview position** (`App.previewPosition`, config `tui.preview_position`) chooses between the
two-column and the stacked layout:

- `auto` selects by terminal size.
- `right` and `below` force a layout; if it does not fit the terminal, the other layout applies, and
  if neither fits, the single-column layout.
- The key `v` toggles between `right` and `below` until the TUI exits.

The constants `TwoColumnFullWidth`, `TwoColumnMinWidth`, `StackedBelowListMinHeight` and the config
`tui.stacked_list_height` set the size limits.

## Panes of the list view

A pane is a bordered rectangle; its **border title** (`withBorderTitle`, `styles.go`) sits in the
top border.

| Name | Code | Content |
|---|---|---|
| list pane (left pane in the two-column layout) | `listModel.viewContent` | the bean rows; border title "Beans", "Beans (active)" or "[tag: …]", replaced by the filter input while typing a filter |
| preview pane (right pane in the two-column layout) | `previewModel` (`preview.go`) | the highlighted bean; border title is its ID |
| footer | `listModel.Footer` | one unbordered line: key help, the selection count "(N selected)", or a status message |

## Columns of a bean row

`ui.RenderBeanRow` (`internal/ui/styles.go`) renders one row per bean, used by both the TUI list pane and
the CLI tree. `ui.CalculateResponsiveColumns` computes the column widths (`ResponsiveColumns`) from
the width it is given.

| Order | Name | Content |
|---|---|---|
| 1 | cursor | `▌` on the highlighted row |
| 2 | ID column | bean ID, preceded by the **tree prefix** (`├─`, `└─`) |
| 3 | type column | type, as one letter or in full (`UseFullTypeStatus`) |
| 4 | status column | status, as one letter or in full; a red `↑` marks an **implicit status** inherited from a closed ancestor |
| 5 | priority symbol | before the title, only for a priority other than normal |
| 6 | title column | title, cut with `...` at `MaxTitleWidth` |
| 7 | tags column | up to `MaxTags` tags, only when `ShowTags` is set |

States of a row:

- **highlighted** (`IsSelected`): the cursor is on it; the preview pane shows this bean.
- **marked** (`IsMarked`): chosen for a batch edit with `space`.
- **dimmed** (`Dimmed`): an ancestor shown for context that does not match the filter itself.

## Parts of the preview pane

| Name | Code | Content |
|---|---|---|
| header | `previewModel.renderHeader` | title, the metadata line "Status: … Type: … Priority: …", the tag line |
| body | `previewModel.renderBody` | the bean's markdown; scrolls (`scrollBy`) with the mouse wheel and `pgup`/`pgdn`, while the header stays |

## Parts of the detail view

| Name | Code | Content |
|---|---|---|
| header box | `detailModel.renderHeader` | title, ID, status badge, tags |
| links section | `detailModel.linkList` | related beans: parent, children, blocking, blocked by; takes focus with `tab` |
| body | `detailModel.viewport` | the bean's markdown |
| footer | `detailModel.View` | scroll percentage and key help |

The focused one of links section and body has a border in the primary color (`linksActive`).
