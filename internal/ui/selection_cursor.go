package ui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/sorokin-vladimir/tele/internal/ui/components"
	"github.com/sorokin-vladimir/tele/internal/ui/keys"
	"github.com/sorokin-vladimir/tele/internal/ui/theme"
)

// Let the terminal blink a single cell beside the selected item. Using the
// terminal cursor avoids repainting the conversation while the user is idle.
func (m RootModel) selectionCursor() *tea.Cursor {
	if m.screen != ScreenMain || m.width < 3 || m.height < 4 ||
		m.mouseBlocked() || m.help != nil || m.settings != nil || m.profile != nil ||
		m.logOut != nil || m.confirmQuit != nil || m.noticeActive() {
		return nil
	}
	layout := computeLayout(m.width, m.height, m.chat.ComposerHeight(), m.folderBar != nil && m.folderBar.HasFolders())
	if m.focus == FocusChat && m.vimState.Mode == keys.ModeInsert {
		cursor := m.chat.ComposerCursor()
		if cursor == nil || !rectContains(components.Rect{Width: layout.composer.Width, Height: layout.composer.Height}, cursor.X, cursor.Y) {
			return nil
		}
		cursor.X += layout.composer.Left
		cursor.Y += layout.composer.Top
		cursor.Blink = true
		cursor.Color = theme.T().BorderComposerFocused
		return cursor
	}
	var pane components.Rect
	var row, col int
	switch m.focus {
	case FocusFolders:
		if !layout.hasFolders {
			return nil
		}
		pane, row = layout.folders, m.folderBar.Cursor()
	case FocusChatList:
		pane = layout.chatList
		if m.forum != nil {
			if _, ok := m.forum.CursorTopic(); !ok {
				return nil
			}
			row = m.forum.CursorViewportRow()
		} else {
			if _, ok := m.chatList.CursorChat(); !ok {
				return nil
			}
			row = m.chatList.CursorViewportRow()
		}
	case FocusChat:
		bubble, ok := m.chat.SelectedBubbleRect()
		if !ok || bubble.Height <= 0 || bubble.Top+bubble.Height <= 0 {
			return nil
		}
		pane, row, col = layout.messages, max(0, bubble.Top), max(0, bubble.Left)
	default:
		return nil
	}
	if row < 0 || row >= pane.Height || col >= pane.Width {
		return nil
	}
	cursor := tea.NewCursor(pane.Left+col, pane.Top+row)
	cursor.Color = theme.T().Indicator
	return cursor
}
