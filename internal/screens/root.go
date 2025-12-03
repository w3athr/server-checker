package screens

import tea "github.com/charmbracelet/bubbletea"

type root struct {
	current tea.Model // current screen model
}

func NewRoot() root {
	return root{
		current: NewMenuScreen(),
	}
}

func (r root) Init() tea.Cmd {
	return r.current.Init()
}

func (r root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	newModel, cmd := r.current.Update(msg)
	r.current = newModel
	return r, cmd
}

func (r root) View() string {
	return r.current.View()
}
