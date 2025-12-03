package screens

import tea "github.com/charmbracelet/bubbletea"

type networkScreen struct{}

func NewNetworkScreen() networkScreen {
	return networkScreen{}
}

func (c networkScreen) Init() tea.Cmd {
	return nil
}

func (c networkScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
		return NewMenuScreen(), nil
	}
	return c, nil
}

func (c networkScreen) View() string {
	return "Network Screen\n\nPress ESC to return to menu."
}
