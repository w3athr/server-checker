package screens

import tea "github.com/charmbracelet/bubbletea"

type ramScreen struct{}

func NewRAMScreen() cpuScreen {
	return cpuScreen{}
}

func (c ramScreen) Init() tea.Cmd {
	return nil
}

func (c ramScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
		return NewMenuScreen(), nil
	}
	return c, nil
}

func (c ramScreen) View() string {
	return "RAM Screen\n\nPress ESC to return to menu."
}
