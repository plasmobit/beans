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

func TestPreviewView(t *testing.T) {
	b := &bean.Bean{
		ID:       "beans-test",
		Title:    "Test Bean",
		Status:   "todo",
		Type:     "feature",
		Priority: "high",
		Tags:     []string{"frontend", "design"},
		Body:     "## Summary\n\nThis is the body.",
	}

	preview := newPreviewModel(b, 60, 20)
	view := preview.View()

	// Should contain the title
	if !strings.Contains(view, "Test Bean") {
		t.Error("preview should contain bean title")
	}

	// Should contain the ID
	if !strings.Contains(view, "beans-test") {
		t.Error("preview should contain bean ID")
	}

	// Should contain status
	if !strings.Contains(view, "todo") {
		t.Error("preview should contain status")
	}

	// Should contain type
	if !strings.Contains(view, "feature") {
		t.Error("preview should contain type")
	}

	// Should contain body content
	if !strings.Contains(view, "Summary") {
		t.Error("preview should contain body")
	}
}

func TestPreviewViewEmpty(t *testing.T) {
	preview := newPreviewModel(nil, 60, 20)
	view := preview.View()

	if !strings.Contains(view, "No bean selected") {
		t.Error("empty preview should show 'No bean selected'")
	}
}

func TestPreviewViewWithTags(t *testing.T) {
	b := &bean.Bean{
		ID:     "beans-test",
		Title:  "Bean with Tags",
		Status: "in-progress",
		Type:   "bug",
		Tags:   []string{"urgent", "backend"},
		Body:   "Test body",
	}

	preview := newPreviewModel(b, 60, 20)
	view := preview.View()

	// Should show tags
	if !strings.Contains(view, "urgent") || !strings.Contains(view, "backend") {
		t.Error("preview should display tags")
	}
}

func TestPreviewViewWithPriority(t *testing.T) {
	b := &bean.Bean{
		ID:       "beans-test",
		Title:    "High Priority Bean",
		Status:   "todo",
		Type:     "task",
		Priority: "critical",
		Body:     "Important work",
	}

	preview := newPreviewModel(b, 60, 20)
	view := preview.View()

	// Should show priority
	if !strings.Contains(view, "critical") {
		t.Error("preview should display priority when not normal")
	}
}

func TestPreviewViewEmptyBody(t *testing.T) {
	b := &bean.Bean{
		ID:     "beans-test",
		Title:  "Bean without body",
		Status: "todo",
		Type:   "task",
		Body:   "",
	}

	preview := newPreviewModel(b, 60, 20)
	view := preview.View()

	// Should show placeholder for empty body
	if !strings.Contains(view, "No description") {
		t.Error("preview should show 'No description' for empty body")
	}
}

func TestPreviewWrapsBodyOnce(t *testing.T) {
	line80 := strings.Repeat("abcdefghi ", 7) + "abcdefghij"
	b := &bean.Bean{ID: "beans-wrap", Title: "Wrap", Status: "todo", Type: "task",
		Body: line80 + "\n" + strings.Repeat("word ", 60)}

	for _, width := range []int{RightPaneMaxWidth, 60} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			view := ansi.Strip(newPreviewModel(b, width, 40).View())
			for _, l := range strings.Split(view, "\n") {
				if w := lipgloss.Width(l); w > width {
					t.Errorf("line width %d exceeds pane width %d: %q", w, width, l)
				}
				if trimmed := strings.TrimSpace(strings.Trim(l, "│")); trimmed == "word" {
					t.Errorf("stray single-word line, body was wrapped twice\n%s", view)
				}
			}
			if width == RightPaneMaxWidth && !strings.Contains(view, line80) {
				t.Errorf("80-column line was broken\n%s", view)
			}
		})
	}
}

func longBodyBean(lines int) *bean.Bean {
	var body strings.Builder
	for i := range lines {
		fmt.Fprintf(&body, "- item %02d\n", i)
	}
	return &bean.Bean{ID: "beans-long", Title: "Long", Status: "todo", Type: "task", Body: body.String()}
}

func TestPreviewScroll(t *testing.T) {
	const height = 20

	tests := []struct {
		name        string
		scrolls     []int
		wantVisible []string
		wantHidden  []string
	}{
		{"initial", nil, []string{"item 00"}, []string{"item 30", "item 49"}},
		{"scrolled down", []int{10}, []string{"item 12"}, []string{"item 05", "item 49"}},
		{"clamped at end", []int{1000}, []string{"item 49"}, []string{"item 00", "..."}},
		{"one notch up from end", []int{1000, -previewScrollStep}, []string{"item 34", "..."}, []string{"item 49"}},
		{"clamped at start", []int{1000, -2000}, []string{"item 00", "..."}, []string{"item 49"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newPreviewModel(longBodyBean(50), RightPaneMaxWidth, height)
			for _, d := range tt.scrolls {
				p.scrollBy(d)
			}
			view := ansi.Strip(p.View())

			if got := lipgloss.Height(view); got != height {
				t.Errorf("height = %d, want %d", got, height)
			}
			if !strings.Contains(view, "beans-long") {
				t.Error("header scrolled away, want it fixed")
			}
			for _, s := range tt.wantVisible {
				if !strings.Contains(view, s) {
					t.Errorf("%q not visible\n%s", s, view)
				}
			}
			for _, s := range tt.wantHidden {
				if strings.Contains(view, s) {
					t.Errorf("%q visible, want hidden\n%s", s, view)
				}
			}
		})
	}
}

func TestStackedLayout(t *testing.T) {
	items := []ui.FlatItem{
		{Bean: longBodyBean(50), Matched: true},
		{Bean: &bean.Bean{ID: "beans-0002", Title: "B", Status: "todo", Type: "task"}, Matched: true},
	}
	defaultMin := config.DefaultStackedListHeight + StackedBelowListMinHeight

	tests := []struct {
		name          string
		listHeight    int // configured tui.stacked_list_height; 0 keeps the default
		width, height int
		wantPreview   bool
	}{
		{"narrow and tall", 0, 100, defaultMin, true},
		{"narrow and short", 0, 100, defaultMin - 1, false},
		{"just below two-column width", 0, TwoColumnMinWidth - 1, 50, true},
		{"two-column preview would shrink", 0, TwoColumnFullWidth - 1, defaultMin, true},
		{"configured taller list", 25, 100, 25 + StackedBelowListMinHeight, true},
		{"configured taller list, too short", 25, 100, 25 + StackedBelowListMinHeight - 1, false},
		{"configured shorter list", 8, 100, 8 + StackedBelowListMinHeight, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Default()
			cfg.TUI.StackedListHeight = tt.listHeight
			listHeight := cfg.GetStackedListHeight()
			a := New(nil, cfg)
			a.Update(tea.WindowSizeMsg{Width: tt.width, Height: tt.height})
			a.Update(beansLoadedMsg{items: items, idColWidth: 12})
			view := ansi.Strip(a.View())
			lines := strings.Split(view, "\n")

			gotPreview := strings.Contains(strings.Join(lines[min(listHeight, len(lines)):], "\n"), "item 00")
			if gotPreview != tt.wantPreview {
				t.Fatalf("preview below list = %v, want %v\n%s", gotPreview, tt.wantPreview, view)
			}
			if !tt.wantPreview {
				return
			}
			if got := len(lines); got != tt.height {
				t.Errorf("view height = %d, want %d", got, tt.height)
			}
			// The footer is excluded: its help line is not truncated in any layout.
			if got := lipgloss.Width(strings.Join(lines[:len(lines)-1], "\n")); got > tt.width {
				t.Errorf("panes width = %d, want <= %d", got, tt.width)
			}
			if got := lipgloss.Width(lines[listHeight]); got != min(tt.width, StackedPreviewMaxWidth) {
				t.Errorf("preview width = %d, want %d", got, min(tt.width, StackedPreviewMaxWidth))
			}

			wheel := func(y int) {
				a.Update(tea.MouseMsg{X: 5, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonWheelDown})
			}
			wheel(listHeight - 1)
			if a.preview.scroll != 0 {
				t.Errorf("wheel over list: preview scroll = %d, want 0", a.preview.scroll)
			}
			wheel(listHeight)
			if a.preview.scroll != previewScrollStep {
				t.Errorf("wheel over preview: scroll = %d, want %d", a.preview.scroll, previewScrollStep)
			}
		})
	}
}

func TestLayoutSelection(t *testing.T) {
	stackedMin := config.DefaultStackedListHeight + StackedBelowListMinHeight
	tests := []struct {
		name          string
		width, height int
		wantTwoColumn bool
		wantStacked   bool
	}{
		{"wide", TwoColumnFullWidth, stackedMin, true, false},
		{"preview would shrink, tall", TwoColumnFullWidth - 1, stackedMin, false, true},
		{"preview would shrink, short", TwoColumnFullWidth - 1, stackedMin - 1, true, false},
		{"narrow, tall", TwoColumnMinWidth - 1, stackedMin, false, true},
		{"narrow, short", TwoColumnMinWidth - 1, stackedMin - 1, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := New(nil, config.Default())
			a.width, a.height = tt.width, tt.height
			if got := a.isTwoColumnMode(); got != tt.wantTwoColumn {
				t.Errorf("isTwoColumnMode() = %v, want %v", got, tt.wantTwoColumn)
			}
			if got := a.isStackedMode(); got != tt.wantStacked {
				t.Errorf("isStackedMode() = %v, want %v", got, tt.wantStacked)
			}
			if tt.wantTwoColumn && tt.width >= TwoColumnFullWidth {
				if w, _ := a.previewSize(); w-previewChromeX < PreviewMinTextWidth {
					t.Errorf("preview text width = %d, want >= %d", w-previewChromeX, PreviewMinTextWidth)
				}
			}
		})
	}
}

func TestPreviewWheel(t *testing.T) {
	items := []ui.FlatItem{
		{Bean: longBodyBean(50), Matched: true},
		{Bean: &bean.Bean{ID: "beans-0002", Title: "B", Status: "todo", Type: "task"}, Matched: true},
	}
	wheelDown := func(x int) tea.MouseMsg {
		return tea.MouseMsg{X: x, Y: 5, Action: tea.MouseActionPress, Button: tea.MouseButtonWheelDown}
	}

	a := New(nil, config.Default())
	a.Update(tea.WindowSizeMsg{Width: 140, Height: 30})
	a.Update(beansLoadedMsg{items: items, idColWidth: 12})
	leftWidth, _ := calculatePaneWidths(140)

	a.Update(wheelDown(5))
	if a.preview.scroll != 0 || a.list.list.Index() != 0 {
		t.Errorf("wheel over list: preview scroll = %d, list index = %d, want both 0", a.preview.scroll, a.list.list.Index())
	}

	a.Update(wheelDown(leftWidth + 5))
	if a.preview.scroll != previewScrollStep {
		t.Errorf("wheel over preview: scroll = %d, want %d", a.preview.scroll, previewScrollStep)
	}

	a.Update(beansLoadedMsg{items: items, idColWidth: 12})
	if a.preview.scroll != previewScrollStep {
		t.Errorf("reload of the same bean: scroll = %d, want %d kept", a.preview.scroll, previewScrollStep)
	}
}

func TestPreviewPageKeys(t *testing.T) {
	var items []ui.FlatItem
	for i := range 100 {
		b := longBodyBean(50)
		b.ID = fmt.Sprintf("beans-%04d", i)
		items = append(items, ui.FlatItem{Bean: b, Matched: true})
	}
	pgDown := tea.KeyMsg{Type: tea.KeyPgDown}
	pgUp := tea.KeyMsg{Type: tea.KeyPgUp}

	tests := []struct {
		name          string
		width, height int
		wantScroll    bool // false: the keys page the list instead
	}{
		{"two columns", 140, 30, true},
		{"stacked", 100, 40, true},
		{"list only", 100, 30, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := New(nil, config.Default())
			a.Update(tea.WindowSizeMsg{Width: tt.width, Height: tt.height})
			a.Update(beansLoadedMsg{items: items, idColWidth: 12})

			a.Update(pgDown)
			a.Update(pgDown)
			if !tt.wantScroll {
				if a.list.list.Index() == 0 {
					t.Error("pgdown without preview: list index = 0, want a later page")
				}
				return
			}
			if a.list.list.Index() != 0 {
				t.Errorf("pgdown with preview: list index = %d, want 0", a.list.list.Index())
			}
			if a.preview.scroll != 2*previewScrollStep {
				t.Errorf("after 2x pgdown: scroll = %d, want %d", a.preview.scroll, 2*previewScrollStep)
			}
			a.Update(pgUp)
			if a.preview.scroll != previewScrollStep {
				t.Errorf("after pgup: scroll = %d, want %d", a.preview.scroll, previewScrollStep)
			}
			a.Update(pgUp)
			a.Update(pgUp)
			if a.preview.scroll != 0 {
				t.Errorf("pgup past the top: scroll = %d, want 0", a.preview.scroll)
			}
		})
	}
}
