package screens

import tea "github.com/charmbracelet/bubbletea"

type reportScreen struct{}

func NewReportScreen() reportScreen {
	return reportScreen{}
}

func (c reportScreen) Init() tea.Cmd {
	return nil
}

func (c reportScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
		return NewMenuScreen(), nil
	}
	return c, nil
}

func (c reportScreen) View() string {
	return "Report Screen\n\nPress ESC to return to menu."
}
