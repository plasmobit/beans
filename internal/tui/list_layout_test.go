package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hmans/beans/internal/ui"
	"github.com/hmans/beans/pkg/bean"
	"github.com/hmans/beans/pkg/config"
)

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
