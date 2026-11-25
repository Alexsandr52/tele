package ui

import (
	"slices"

	tea "charm.land/bubbletea/v2"

	"github.com/sorokin-vladimir/tele/internal/ui/keys"
	"github.com/sorokin-vladimir/tele/internal/ui/screens"
)

// focusedContext maps the focused pane to its key-binding context.
func (m RootModel) focusedContext() keys.Context {
	switch m.focus {
	case FocusFolders:
		return keys.ContextFolders
	case FocusChatList:
		return keys.ContextChatList
	default:
		return keys.ContextChat
	}
}

func (m RootModel) focusPrev() (tea.Model, tea.Cmd) {
	hasFolders := m.folderBar != nil && m.folderBar.HasFolders()
	switch m.focus {
	case FocusChat:
		return m.focusPane(FocusChatList)
	case FocusChatList:
		if hasFolders {
			return m.focusPane(FocusFolders)
		}
		return m, nil
	case FocusFolders:
		return m.focusPane(FocusChat)
	}
	return m, nil
}

func (m RootModel) focusNext() (tea.Model, tea.Cmd) {
	hasFolders := m.folderBar != nil && m.folderBar.HasFolders()
	switch m.focus {
	case FocusFolders:
		return m.focusPane(FocusChatList)
	case FocusChatList:
		return m.focusPane(FocusChat)
	case FocusChat:
		if hasFolders {
			return m.focusPane(FocusFolders)
		}
		return m, nil
	}
	return m, nil
}

// The editor lives inside the chat pane. It has its own keyboard stop,
// without becoming a separate layout pane.
const tabComposer Focus = FocusChat + 1

func (m RootModel) cycleTab(step int) (tea.Model, tea.Cmd) {
	stops := make([]Focus, 0, 4)
	if m.folderBar != nil && m.folderBar.HasFolders() {
		stops = append(stops, FocusFolders)
	}
	stops = append(stops, FocusChatList, FocusChat)
	if m.currentChatID != 0 {
		stops = append(stops, tabComposer)
	}
	current := m.focus
	if m.focus == FocusChat && m.chat.ComposerFocused() {
		current = tabComposer
	}
	i := slices.Index(stops, current)
	next := stops[(i+step+len(stops))%len(stops)]
	if next == tabComposer {
		return m.focusComposer()
	}
	if current == tabComposer && next == FocusChat {
		cmd := m.blurComposer()
		return m, cmd
	}
	return m.focusPane(next)
}

func (m RootModel) focusComposer() (tea.Model, tea.Cmd) {
	next, paneCmd := m.focusPane(FocusChat)
	m = next.(RootModel)
	if m.chat.ComposerFocused() {
		return m, paneCmd
	}
	m.vimState.Mode = keys.ModeInsert
	m.statusBar.SetMode(keys.ModeInsert)
	return m, tea.Batch(paneCmd, m.chat.FocusComposer())
}

func (m *RootModel) blurComposer() tea.Cmd {
	m.vimState.Mode = keys.ModeNormal
	m.statusBar.SetMode(keys.ModeNormal)
	pane, cmd := m.chat.Update(keys.ActionMsg{Action: keys.ActionNormal})
	m.chat = pane.(*screens.ChatModel)
	return cmd
}

func (m RootModel) focusPane(target Focus) (tea.Model, tea.Cmd) {
	if target == m.focus {
		return m, nil
	}
	m.matcher.Reset()
	var blurCmd tea.Cmd
	if m.focus == FocusChat && m.vimState.Mode == keys.ModeInsert {
		blurCmd = m.blurComposer()
	}
	m.focus = target
	m.chatList.SetFocused(target == FocusChatList)
	if m.forum != nil {
		m.forum.SetFocused(target == FocusChatList)
	}
	m.chat.SetFocused(target == FocusChat)
	if m.folderBar != nil {
		m.folderBar.SetFocused(target == FocusFolders)
	}
	switch target {
	case FocusFolders:
		m.statusBar.SetActivePane("folders")
	case FocusChatList:
		m.statusBar.SetActivePane("chatlist")
	case FocusChat:
		m.statusBar.SetActivePane("chat")
		// A reaction that arrived while the pane was not focused is only being
		// looked at now, so this is when it counts as seen.
		if m.chatUnreadReactions > 0 {
			return m, tea.Batch(blurCmd, m.readReactionsCmd(m.currentHistory()))
		}
	}
	return m, blurCmd
}
