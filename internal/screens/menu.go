package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type menu struct {
	cursor     int
	items      []menuItem
	lastWidth  int
	lastHeight int
}

type menuItem struct {
	text    string
	onPress func() tea.Model
}

func NewMenuScreen() menu {
	return menu{
		items: []menuItem{
			{text: "CPU", onPress: func() tea.Model { return NewCPUScreen() }},
			{text: "RAM", onPress: func() tea.Model { return NewRAMScreen() }},
			{text: "Network", onPress: func() tea.Model { return NewNetworkScreen() }},
			{text: "Disks", onPress: func() tea.Model { return NewDisksScreen() }},
			{text: "Report", onPress: func() tea.Model { return NewReportMenuScreen() }},
		},
	}
}

func (m menu) Init() tea.Cmd {
	return nil
}

func (m menu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.lastWidth = msg.Width
		m.lastHeight = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			newModel := m.items[m.cursor].onPress()
			return newModel, func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.lastWidth, Height: m.lastHeight}
			}
		}
	}
	return m, nil
}

func (m menu) View() string {
	s := "Select an option:\n\n"

	for i, item := range m.items {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor
		}
		s += fmt.Sprintf("%s %s\n", cursor, item.text)
	}

	s += "\n(↑/↓)/(k/j): Select | Q: Exit\n"

	return s
}
