package screens

import tea "github.com/charmbracelet/bubbletea"

type cpuScreen struct{}

func NewCPUScreen() cpuScreen {
	return cpuScreen{}
}

func (c cpuScreen) Init() tea.Cmd {
	return nil
}

func (c cpuScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
		return NewMenuScreen(), nil
	}
	return c, nil
}

func (c cpuScreen) View() string {
	return "CPU Screen\n\nPress ESC to return to menu."
}
