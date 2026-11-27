package ui

import (
	"image"
	"image/color"
	"image/draw"
	"os"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/sorokin-vladimir/tele/internal/core/project"
	"github.com/sorokin-vladimir/tele/internal/domain"
	"github.com/sorokin-vladimir/tele/internal/ui/media"
	"github.com/stretchr/testify/require"
)

func controlsFixture() RootModel {
	m := NewRootModel(false).WithScreen(ScreenMain).WithImageMode(media.ModeKitty)
	m.currentChatID = 1
	m.chatList.SetWindow(0, 2, []project.ChatRow{{ID: 1, Title: "Photo test"}, {ID: 2, Title: "Another chat"}})
	m.chat.SetKnownImages(m.imageCache)
	day := time.Now()
	m.chat.SetMessages([]domain.Message{
		{ID: 1, ChatID: 1, Date: day, Text: "First message"},
		{ID: 2, ChatID: 1, Date: day, Text: "Test pattern", Photo: &domain.PhotoRef{ID: 42}, Media: &domain.MediaRef{Kind: domain.MediaPhoto}},
		{ID: 3, ChatID: 1, Date: day, Text: "Last message"},
	})
	sized, _ := m.Update(tea.WindowSizeMsg{Width: 95, Height: 34})
	return sized.(RootModel)
}

func TestLocalNavigation(t *testing.T) {
	m := controlsFixture()
	key := func(code rune, mod tea.KeyMod) {
		next, _ := m.Update(tea.KeyPressMsg{Code: code, Mod: mod})
		m = next.(RootModel)
	}
	cursor := m.View().Cursor
	require.NotNil(t, cursor)
	require.True(t, cursor.Blink)
	key(tea.KeyDown, 0)
	require.Equal(t, cursor.Y+1, m.View().Cursor.Y)
	key(tea.KeyTab, 0)
	require.Equal(t, FocusChat, m.focus)
	require.NotNil(t, m.View().Cursor)
	key(tea.KeyUp, 0)
	require.Equal(t, 2, m.chat.SelectedMessageID())
	photoCursor := m.View().Cursor
	key(tea.KeyDown, 0)
	require.Equal(t, 3, m.chat.SelectedMessageID())
	require.Greater(t, m.View().Cursor.Y, photoCursor.Y)
	// Entering the text field does not require a separate insert-mode key.
	key(tea.KeyTab, 0)
	require.True(t, m.chat.ComposerFocused())
	require.Equal(t, "Message", m.chat.ComposerPlaceholder())
	inputCursor := m.View().Cursor
	require.NotNil(t, inputCursor)
	require.True(t, inputCursor.Blink)
	require.Greater(t, inputCursor.Y, photoCursor.Y)
	for _, r := range "Черновик" {
		next, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = next.(RootModel)
	}
	require.Equal(t, "Черновик", m.chat.ComposerValue())
	require.Equal(t, inputCursor.X+len([]rune("Черновик")), m.View().Cursor.X)
	key(tea.KeyTab, 0)
	require.Equal(t, FocusChatList, m.focus)
	require.False(t, m.chat.ComposerFocused())
	key(tea.KeyTab, tea.ModShift)
	require.True(t, m.chat.ComposerFocused())
	require.Equal(t, "Черновик", m.chat.ComposerValue())
	key(tea.KeyTab, tea.ModShift)
	require.Equal(t, FocusChat, m.focus)
	require.False(t, m.chat.ComposerFocused())
	require.Contains(t, m.chat.ComposerPlaceholder(), "Tab")
	m.folderBar.SetFolders([]domain.FolderFilter{{ID: 7, Title: "Folder"}})
	key(tea.KeyTab, 0)
	require.True(t, m.chat.ComposerFocused())
	key(tea.KeyTab, 0)
	require.Equal(t, FocusFolders, m.focus)
	require.NotNil(t, m.View().Cursor)
	key(tea.KeyTab, 0)
	require.Equal(t, FocusChatList, m.focus)
	key(tea.KeyTab, tea.ModShift)
	require.Equal(t, FocusFolders, m.focus)
	key(tea.KeyTab, tea.ModShift)
	require.True(t, m.chat.ComposerFocused())
	key(tea.KeyEsc, 0)
	require.False(t, m.chat.ComposerFocused())
	require.Equal(t, "Черновик", m.chat.ComposerValue())
	m.currentChatID = 0
	key(tea.KeyTab, 0)
	require.Equal(t, FocusFolders, m.focus)
	require.False(t, m.chat.ComposerFocused())
}

func TestLocalGraphicsFixture(t *testing.T) {
	if os.Getenv("TELE_GRAPHICS_FIXTURE") != "1" {
		t.Skip("requires an isolated graphical terminal")
	}
	m := controlsFixture()
	img := image.NewRGBA(image.Rect(0, 0, 320, 128))
	draw.Draw(img, image.Rect(0, 0, 160, 128), image.NewUniform(color.RGBA{246, 34, 107, 255}), image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(160, 0, 320, 128), image.NewUniform(color.RGBA{22, 225, 155, 255}), image.Point{}, draw.Src)
	p := tea.NewProgram(m)
	go func() { time.Sleep(300 * time.Millisecond); p.Send(PhotoReadyMsg{PhotoID: 42, Image: img}) }()
	_, err := p.Run()
	require.NoError(t, err)
}

// M hands the mouse to the terminal so text can be selected with a drag,
// and takes it back when pressed again.
func TestToggleMouseReleasesSelection(t *testing.T) {
	m := controlsFixture()
	require.Equal(t, tea.MouseModeCellMotion, m.View().MouseMode)
	press := func() {
		next, _ := m.Update(tea.KeyPressMsg{Code: 'M', Text: "M"})
		m = next.(RootModel)
	}
	press()
	require.Equal(t, tea.MouseModeNone, m.View().MouseMode)
	press()
	require.Equal(t, tea.MouseModeCellMotion, m.View().MouseMode)
}
