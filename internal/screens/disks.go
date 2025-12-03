package screens

import tea "github.com/charmbracelet/bubbletea"

type disksScreen struct{}

func NewDisksScreen() disksScreen {
	return disksScreen{}
}

func (c disksScreen) Init() tea.Cmd {
	return nil
}

func (c disksScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
		return NewMenuScreen(), nil
	}
	return c, nil
}

func (c disksScreen) View() string {
	return "Disks Screen\n\nPress ESC to return to menu."
}
