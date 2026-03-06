package screens

import (
	"server-checker/internal/collector"
	"server-checker/internal/models"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type root struct {
	current tea.Model
	data    models.SystemInfo
	width   int
	height  int
}

type TickMsg models.SystemInfo

func doTick() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		data, _ := collector.CollectAll()
		return TickMsg(data)
	})
}

func NewRoot() root {
	return root{
		current: NewMenuScreen(),
	}
}

func (r root) Init() tea.Cmd {
	return tea.Batch(r.current.Init(), doTick())
}

func (r root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.width, r.height = msg.Width, msg.Height
	case TickMsg:
		r.data = models.SystemInfo(msg)
		cmds = append(cmds, doTick())
	case tea.KeyMsg:
		s := msg.String()
		if s == "ctrl+c" {
			return r, tea.Quit
		}
		// Возвращаем функционал Q: выход или возврат в меню
		if s == "q" || s == "Q" {
			if _, ok := r.current.(menu); ok {
				return r, tea.Quit
			}
			r.current = NewMenuScreen()
			return r, tea.Batch(
				func() tea.Msg { return tea.WindowSizeMsg{Width: r.width, Height: r.height} },
				tea.ClearScreen,
			)
		}
	}

	// Передаем сообщение в дочерний экран (включая ESC для отмены)
	newModel, cmd := r.current.Update(msg)
	r.current = newModel
	cmds = append(cmds, cmd)

	return r, tea.Batch(cmds...)
}

func (r root) View() string {
	return r.current.View()
}