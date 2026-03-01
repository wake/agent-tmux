package ui

import (
	"github.com/wake/tmux-session-menu/internal/store"
	"github.com/wake/tmux-session-menu/internal/tmux"
)

// ItemType 區分列表項目的種類。
type ItemType int

const (
	ItemSession ItemType = iota
	ItemGroup
)

// ListItem 代表列表中的一個項目（session 或群組標頭）。
type ListItem struct {
	Type    ItemType
	Session tmux.Session
	Group   store.Group
}
