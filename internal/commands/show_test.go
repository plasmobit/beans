package commands

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/hmans/beans/pkg/bean"
)

func TestFormatRelationships_ClosedAncestor(t *testing.T) {
	tests := []struct {
		name         string
		bean         *bean.Bean
		closedStatus string
		closedFrom   string
		want         string
	}{
		{
			name: "open parent",
			bean: &bean.Bean{ID: "t1", Parent: "e1"},
			want: "parent: e1",
		},
		{
			name:         "closed parent",
			bean:         &bean.Bean{ID: "t1", Parent: "e1"},
			closedStatus: "completed",
			closedFrom:   "e1",
			want:         "parent: e1 !completed",
		},
		{
			name:         "closed grandparent",
			bean:         &bean.Bean{ID: "t1", Parent: "e1"},
			closedStatus: "scrapped",
			closedFrom:   "m1",
			want:         "parent: e1\nancestor: m1 !scrapped",
		},
		{
			name:         "closed parent with blocking",
			bean:         &bean.Bean{ID: "t1", Parent: "e1", Blocking: []string{"t2"}},
			closedStatus: "completed",
			closedFrom:   "e1",
			want:         "parent: e1 !completed\nblocking: t2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ansi.Strip(formatRelationships(tt.bean, tt.closedStatus, tt.closedFrom))
			if got != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}
