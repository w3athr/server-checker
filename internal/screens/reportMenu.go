package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// reportMenu - это новый экран, который будет отображать подменю для отчетов.
type reportMenu struct {
	cursor     int
	items      []menuItem
	lastWidth  int
	lastHeight int
}

// NewReportMenuScreen создает и возвращает новый экземпляр reportMenu.
func NewReportMenuScreen() reportMenu {
	return reportMenu{
		items: []menuItem{
			// Здесь будут пункты "Выбрать эталон" и "Выгрузить отчет"
			{text: "Select Template", onPress: func() tea.Model { return NewTemplateSelectionScreen() }}, // Это будет следующий экран
			{text: "Generate Report", onPress: func() tea.Model { return NewGenerateReportScreen() }},    // И этот тоже
		},
	}
}

// Init вызывается при инициализации модели.
func (rm reportMenu) Init() tea.Cmd {
	return nil
}

// Update обрабатывает сообщения и обновляет состояние модели.
func (rm reportMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Сохраняем размеры окна для передачи дочерним моделям.
		rm.lastWidth = msg.Width
		rm.lastHeight = msg.Height
		return rm, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if rm.cursor > 0 {
				rm.cursor--
			}
		case "down", "j":
			if rm.cursor < len(rm.items)-1 {
				rm.cursor++
			}
		case "enter":
			// При нажатии Enter запускаем соответствующую модель из подменю.
			newModel := rm.items[rm.cursor].onPress()
			return newModel, func() tea.Msg {
				// Передаем размеры окна новой модели, чтобы она правильно инициализировалась.
				return tea.WindowSizeMsg{Width: rm.lastWidth, Height: rm.lastHeight}
			}
		case "esc":
			// Возврат в главное меню
			return NewMenuScreen(), func() tea.Msg {
				return tea.WindowSizeMsg{Width: rm.lastWidth, Height: rm.lastHeight}
			}
		}
	}
	return rm, nil
}

// View отображает текущее состояние модели.
func (rm reportMenu) View() string {
	s := "Report Options:\n\n"

	for i, item := range rm.items {
		cursor := " " // no cursor
		if rm.cursor == i {
			cursor = ">" // cursor
		}
		s += fmt.Sprintf("%s %s\n", cursor, item.text)
	}

	s += "\n(↑/↓)/(k/j): Select | ESC: Back to Main Menu\n"

	return s
}
