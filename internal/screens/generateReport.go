package screens

import tea "github.com/charmbracelet/bubbletea"

// generateReportScreen - экран для генерации отчета.
type generateReportScreen struct{}

func NewGenerateReportScreen() generateReportScreen {
	return generateReportScreen{}
}

func (g generateReportScreen) Init() tea.Cmd {
	return nil
}

func (g generateReportScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			// Возвращаемся в меню отчетов
			return NewReportMenuScreen(), nil
		}
	}
	return g, nil
}

func (g generateReportScreen) View() string {
	return "Generate Report Screen (Coming Soon!)\n\nPress ESC to return to Report Menu."
}
