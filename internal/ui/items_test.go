package ui_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wake/tmux-session-menu/internal/store"
	"github.com/wake/tmux-session-menu/internal/tmux"
	"github.com/wake/tmux-session-menu/internal/ui"
)

func TestFlattenItems(t *testing.T) {
	groups := []store.Group{
		{ID: 1, Name: "dev", SortOrder: 0},
		{ID: 2, Name: "ops", SortOrder: 1},
	}
	sessions := []tmux.Session{
		{Name: "project-a", GroupName: "dev"},
		{Name: "project-b", GroupName: "dev"},
		{Name: "monitoring", GroupName: "ops"},
		{Name: "standalone"},
	}

	items := ui.FlattenItems(groups, sessions)

	// Ungrouped sessions first
	assert.Equal(t, ui.ItemSession, items[0].Type)
	assert.Equal(t, "standalone", items[0].Session.Name)

	// Then groups
	assert.Equal(t, ui.ItemGroup, items[1].Type)
	assert.Equal(t, "dev", items[1].Group.Name)

	assert.Equal(t, ui.ItemSession, items[2].Type)
	assert.Equal(t, "project-a", items[2].Session.Name)
	assert.Equal(t, ui.ItemSession, items[3].Type)
	assert.Equal(t, "project-b", items[3].Session.Name)

	assert.Equal(t, ui.ItemGroup, items[4].Type)
	assert.Equal(t, "ops", items[4].Group.Name)
	assert.Equal(t, ui.ItemSession, items[5].Type)
	assert.Equal(t, "monitoring", items[5].Session.Name)
}

func TestFlattenItems_CollapsedGroup(t *testing.T) {
	groups := []store.Group{
		{ID: 1, Name: "dev", SortOrder: 0, Collapsed: true},
	}
	sessions := []tmux.Session{
		{Name: "project-a", GroupName: "dev"},
		{Name: "project-b", GroupName: "dev"},
	}

	items := ui.FlattenItems(groups, sessions)
	assert.Len(t, items, 1)
	assert.Equal(t, ui.ItemGroup, items[0].Type)
}
