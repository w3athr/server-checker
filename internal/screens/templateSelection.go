package screens

import tea "github.com/charmbracelet/bubbletea"

// templateSelectionScreen - экран для выбора эталонной конфигурации.
type templateSelectionScreen struct{}

func NewTemplateSelectionScreen() templateSelectionScreen {
	return templateSelectionScreen{}
}

func (t templateSelectionScreen) Init() tea.Cmd {
	return nil
}

func (t templateSelectionScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			// Возвращаемся в меню отчетов
			return NewReportMenuScreen(), tea.ClearScreen
		}
	}
	return t, nil
}

func (t templateSelectionScreen) View() string {
	return "Template Selection Screen (Coming Soon!)\n\nPress ESC to return to Report Menu."
}
