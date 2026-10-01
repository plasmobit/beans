package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/hmans/beans/internal/ui"
	"github.com/hmans/beans/pkg/bean"
	"github.com/hmans/beans/pkg/config"
)

func TestListRowsFitPane(t *testing.T) {
	var items []ui.FlatItem
	for i := range 5 {
		items = append(items, ui.FlatItem{
			Bean: &bean.Bean{ID: fmt.Sprintf("anvl-%04d", i), Status: "completed", Type: "task", Priority: "critical",
				Title: strings.Repeat("a long title that must be truncated ", 6)},
			Depth:      1,
			Matched:    true,
			TreePrefix: "├─",
		})
	}

	// The widths cover short and full type/status names.
	for _, width := range []int{80, 120, 124, 160} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			m := newListModel(nil, config.Default())
			m, _ = m.Update(tea.WindowSizeMsg{Width: width, Height: 40})
			m, _ = m.Update(beansLoadedMsg{items: items, idColWidth: 9 + 2 + 2})
			out := ansi.Strip(m.ViewConstrained(width, StackedListHeight))

			for _, line := range strings.Split(out, "\n") {
				if got := lipgloss.Width(line); got != width {
					t.Errorf("line width = %d, want %d: %q", got, width, line)
				}
				if strings.TrimSpace(strings.Trim(line, "│")) == "..." {
					t.Errorf("truncation marker wrapped onto its own line\n%s", out)
				}
			}
		})
	}
}

func TestListViewConstrained_ExactSize(t *testing.T) {
	// Children of a completed parent carry an implicit-status annotation;
	// deep tree prefixes shrink the title budget to a few columns.
	items := []ui.FlatItem{{
		Bean:    &bean.Bean{ID: "beans-root", Title: "Completed milestone", Status: "completed", Type: "milestone"},
		Matched: true,
	}}
	for i := range 30 {
		items = append(items, ui.FlatItem{
			Bean:           &bean.Bean{ID: fmt.Sprintf("beans-%04d", i), Title: "A child title long enough to be truncated", Status: "completed", Type: "task"},
			Depth:          1 + i%3,
			Matched:        true,
			TreePrefix:     strings.Repeat("│  ", i%3) + "├─",
			ImplicitStatus: "completed",
		})
	}
	const idColWidth = 10 + 2 + 3*3

	for _, sz := range [][2]int{{120, 27}, {120, 28}, {140, 40}} {
		width, height := sz[0], sz[1]
		t.Run(fmt.Sprintf("%dx%d", width, height), func(t *testing.T) {
			m := newListModel(nil, config.Default())
			m, _ = m.Update(tea.WindowSizeMsg{Width: width, Height: height})
			m, _ = m.Update(beansLoadedMsg{items: items, idColWidth: idColWidth})

			leftWidth, _ := calculatePaneWidths(width)
			contentHeight := height - 1
			out := m.ViewConstrained(leftWidth, contentHeight)

			if got := lipgloss.Height(out); got != contentHeight {
				t.Errorf("pane height = %d, want %d\n%s", got, contentHeight, out)
			}
			if got := lipgloss.Width(out); got != leftWidth {
				t.Errorf("pane width = %d, want %d", got, leftWidth)
			}
		})
	}
}
