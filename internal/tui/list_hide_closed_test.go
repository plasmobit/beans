package tui

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hmans/beans/pkg/bean"
	"github.com/hmans/beans/pkg/beancore"
	"github.com/hmans/beans/pkg/beangraph"
	"github.com/hmans/beans/pkg/config"
)

func newHideClosedTestList(t *testing.T) listModel {
	t.Helper()
	beansDir := filepath.Join(t.TempDir(), ".beans")
	if err := os.MkdirAll(beansDir, 0755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	core := beancore.New(beansDir, cfg)
	if err := core.Load(); err != nil {
		t.Fatal(err)
	}
	for _, b := range []*bean.Bean{
		{ID: "open", Title: "Open", Status: "todo", Type: "task", Tags: []string{"x"}},
		{ID: "done", Title: "Done", Status: "completed", Type: "task", Tags: []string{"x"}},
		{ID: "dropped", Title: "Dropped", Status: "scrapped", Type: "task"},
		{ID: "done-epic", Title: "Done epic", Status: "completed", Type: "epic"},
		{ID: "orphan", Title: "Under done epic", Status: "todo", Type: "task", Parent: "done-epic"},
	} {
		b.Slug = bean.Slugify(b.Title)
		if err := core.Create(b); err != nil {
			t.Fatal(err)
		}
	}
	return newListModel(&beangraph.CoreResolver{Core: core}, cfg)
}

func loadedIDs(t *testing.T, m listModel) []string {
	t.Helper()
	msg, ok := m.loadBeans().(beansLoadedMsg)
	if !ok {
		t.Fatalf("loadBeans did not return beansLoadedMsg")
	}
	var ids []string
	for _, it := range msg.items {
		ids = append(ids, it.Bean.ID)
	}
	slices.Sort(ids)
	return ids
}

func TestListHideClosed(t *testing.T) {
	tests := []struct {
		name       string
		hideClosed bool
		tagFilter  string
		want       []string
	}{
		{"shows all by default", false, "", []string{"done", "done-epic", "dropped", "open", "orphan"}},
		{"hides archive statuses and their descendants", true, "", []string{"open"}},
		{"combines with tag filter", true, "x", []string{"open"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newHideClosedTestList(t)
			m.hideClosed = tt.hideClosed
			m.tagFilter = tt.tagFilter
			if got := loadedIDs(t, m); !slices.Equal(got, tt.want) {
				t.Errorf("ids = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListHideClosedToggle(t *testing.T) {
	m := newListModel(nil, config.Default())
	m.tagFilter = "x"

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	if !m.hideClosed || cmd == nil {
		t.Fatalf("h must enable hideClosed and trigger a reload (hideClosed=%v, cmd=%v)", m.hideClosed, cmd != nil)
	}
	if got, want := m.title(), "Beans (active) [tag: x]"; got != want {
		t.Errorf("title = %q, want %q", got, want)
	}

	m.clearFilter()
	if !m.hideClosed {
		t.Error("clearFilter must not reset hideClosed")
	}
	if got, want := m.title(), "Beans (active)"; got != want {
		t.Errorf("title = %q, want %q", got, want)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	if m.hideClosed {
		t.Error("second h must disable hideClosed")
	}
}
