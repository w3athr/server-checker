package screens

import (
	"server-checker/internal/collector"
	"server-checker/internal/models"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type root struct {
	current tea.Model
	data    models.SystemInfo // добавлено для хранения собранной информации
	width   int               // ширина для вьюпорта
	height  int               // высота для вьюпорта
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
	return tea.Batch(r.current.Init(), doTick()) // запуск сбора информации при отрисовке корневого экрана
}

func (r root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.width = msg.Width
		r.height = msg.Height
		r.current, cmd = r.current.Update(msg)
		return r, cmd

	case TickMsg:
		r.data = models.SystemInfo(msg)        // обновление данных в корне при получении TickMsg
		r.current, cmd = r.current.Update(msg) // передача сообщения текущему экрану
		return r, doTick()                     // повторный запуск таймера

	case tea.KeyMsg: // обработка клавиш
		switch msg.String() {
		case "enter":
			if m, ok := r.current.(menu); ok {
				selectedScreen := m.items[m.cursor].onPress()
				r.current, cmd = selectedScreen.Update(tea.WindowSizeMsg{
					Width:  r.width,
					Height: r.height,
				})
			}
		case "esc":
			r.current = NewMenuScreen()
			return r, func() tea.Msg {
				return tea.WindowSizeMsg{Width: r.width, Height: r.height}
			}
		case "ctrl+c":
			return r, tea.Quit
		case "q", "Q":
			if _, ok := r.current.(menu); ok {
				return r, tea.Quit
			} else {
				r.current = NewMenuScreen()
				return r, nil
			}
		}
	}

	r.current, cmd = r.current.Update(msg)
	return r, cmd
}

func (r root) View() string {
	return r.current.View()
}
