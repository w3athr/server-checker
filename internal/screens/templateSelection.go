package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// templateSelectionScreen - экран для выбора эталонной конфигурации.
type templateSelectionScreen struct {
	cursor    int
	options   []string
	inputMode bool            // режим ввода пути к файлу
	textInput textinput.Model // компонент ввода текста
	err       string          // сообщение об ошибке
}

func NewTemplateSelectionScreen() templateSelectionScreen {
	ti := textinput.New()
	ti.Placeholder = "/path/to/custom_template.yaml"
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 40

	return templateSelectionScreen{
		options: []string{
			"Рубеж-Н1",
			"Рубеж-С",
			"Рубеж-Т",
			"Custom Template",
		},
		textInput: ti,
	}
}

func (t templateSelectionScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (t templateSelectionScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Если в режиме ввода текста
		if t.inputMode {
			switch msg.String() {
			case "enter":
				path := strings.TrimSpace(t.textInput.Value())
				if path != "" {
					// Сохраняем выбор
					SelectedTemplateName = "Custom"
					SelectedTemplatePath = path
					// Возврат в меню отчетов
					return NewReportMenuScreen(), tea.ClearScreen
				}

			case "esc":
				// отмена ввода
				t.inputMode = false
				t.textInput.Reset()
				return t, tea.ClearScreen
			}

			t.textInput, cmd = t.textInput.Update(msg)
			return t, cmd
		}

		// Если в режиме выбор из меню
		switch msg.String() {
		case "up", "k":
			if t.cursor > 0 {
				t.cursor--
			}

		case "down", "j":
			if t.cursor < len(t.options)-1 {
				t.cursor++
			}

		case "enter":
			selected := t.options[t.cursor]
			if selected == "Custom Template" {
				t.inputMode = true // Включение режима ввода
				return t, tea.ClearScreen
			} else {
				// Сохранение выбора предустановленного шаблона
				SelectedTemplateName = selected
				SelectedTemplatePath = "" // Пустой путь, потом добавлю
				// Возврат в меню отчета
				return NewReportMenuScreen(), tea.ClearScreen
			}

		case "esc":
			return NewReportMenuScreen(), tea.ClearScreen
		}
	}

	return t, cmd
}

func (t templateSelectionScreen) View() string {
	s := "Select Template:\n\n"

	// Если включен режим ввода, отображать поле ввода
	if t.inputMode {
		s += "Enter path to custom YAML template:\n"
		s += t.textInput.View() + "\n\n"
		s += "(Esc to cancel, Enter to confirm)"
		return s
	}

	// Иначе показывать список опций
	for i, option := range t.options {
		cursor := " "
		if t.cursor == i {
			cursor = ">"
		}

		// Отмечаем текущий выбранный шаблон галочной если он был выбран
		checked := ""
		if SelectedTemplateName == option || (option == "Custom Template" && SelectedTemplateName == "Custom") {
			checked = " [x]"
		}

		s += fmt.Sprintf("%s %s%s\n", cursor, option, checked)
	}

	s += "\n(↑/↓)/(k/j): Select | Enter: Confirm | ESC: Back"

	return s
}
