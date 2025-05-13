package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/w3athr/server-checker/internal/collector"
)

const (
	stateSelectEquip   = "selectEquip"
	stateSelectProduct = "selectProduct"
	stateMainMenu      = "mainMenu"

	stateShowConfig = "showConfig"
	stateTestLEDs   = "testLEDs"
	stateShowLoad   = "showLoad"
	stateFillReport = "fillReport"
)

type configMsg struct {
	lines []string
}

type model struct {
	state string

	equipOptions  []string
	equipIndex    int
	equipmentType string

	productOptions []string
	productIndex   int
	productType    string

	menuOptions []string
	menuIndex   int

	configLines  []string
	SystemConfig *collector.SystemConfig
}

func initialModel() model {
	return model{
		state:        stateSelectEquip,
		equipOptions: []string{"рубеж-с", "рубеж-т", "рубеж-н", "бокс", "тринити", "супермикро"},
		equipIndex:   0,

		productOptions: []string{"ngfw", "ifw"},
		productIndex:   0,

		menuOptions: []string{
			"Показать конфигурацию оборудования",
			"Тест LED интерфейсов",
			"Показать нагрузку CPU/RAM",
			"Заполнить отчет",
		},
		menuIndex: 0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if cfg, ok := msg.(configMsg); ok {
		m.configLines = cfg.lines
		return m, nil
	}

	if key, ok := msg.(tea.KeyMsg); ok && (key.String() == "q" || key.String() == "ctrl+c") {
		switch m.state {
		case stateSelectEquip:
			return m, tea.Quit

		case stateSelectProduct:
			m.state = stateSelectEquip
			return m, nil

		case stateMainMenu:
			m.state = stateSelectProduct
			return m, nil

		case stateShowConfig:
			m.state = stateMainMenu
			return m, nil

		case stateTestLEDs, stateShowLoad, stateFillReport:
			m.state = stateMainMenu
			return m, nil
		}
	}

	switch m.state {

	case stateSelectEquip:
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "up", "k":
				if m.equipIndex > 0 {
					m.equipIndex--
				}
			case "down", "j":
				if m.equipIndex < len(m.equipOptions)-1 {
					m.equipIndex++
				}
			case "enter":
				m.equipmentType = m.equipOptions[m.equipIndex]
				m.state = stateSelectProduct
			}
		}

	case stateSelectProduct:
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "up", "k":
				if m.productIndex > 0 {
					m.productIndex--
				}
			case "down", "j":
				if m.productIndex < len(m.productOptions)-1 {
					m.productIndex++
				}
			case "enter":
				m.productType = m.productOptions[m.productIndex]
				m.state = stateMainMenu
			}
		}

	case stateMainMenu:
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "up", "k":
				if m.menuIndex > 0 {
					m.menuIndex--
				}
			case "down", "j":
				if m.menuIndex < len(m.menuOptions)-1 {
					m.menuIndex++
				}
			case "enter":
				switch m.menuIndex {
				case 0:
					// переключаемся в режим показа конфигурации
					m.state = stateShowConfig
					// запускаем Collector и ждём configMsg
					return m, func() tea.Msg {
						cfg, err := collector.CollectConfig()
						if err != nil {
							return configMsg{lines: []string{fmt.Sprintf("Error: %v", err)}}
						}
						return configMsg{lines: collector.FormatConfigLines(cfg)}
					}
				case 1:
					m.state = stateTestLEDs
				case 2:
					m.state = stateShowLoad
				case 3:
					m.state = stateFillReport
				}
			}
		}

	case stateShowConfig:
		// ничего не делаем — данные уже в m.configLines

	case stateTestLEDs:

	case stateShowLoad:

	case stateFillReport:

	}
	return m, nil
}

func (m model) View() string {
	switch m.state {
	case stateSelectEquip:
		s := "1) Выберите модель оборудования:\n\n"
		for i, opt := range m.equipOptions {
			cursor := " "
			if i == m.equipIndex {
				cursor = "> "
			}
			s += fmt.Sprintf("%s%s\n", cursor, opt)
		}
		s += "\n ↑/↓ (k/j) - навигация, Enter — подтвердить, q - выход.\n"
		return s

	case stateSelectProduct:
		s := fmt.Sprintf("Оборудование: %s\n\n", m.equipmentType)
		s += "2) Выберите продукт:\n\n"
		for i, opt := range m.productOptions {
			cursor := " "
			if i == m.productIndex {
				cursor = "> "
			}
			s += fmt.Sprintf("%s%s\n", cursor, opt)
		}
		s += "\n ↑/↓ (k/j) - навигация, Enter — подтвердить, q - назад.\n"
		return s

	case stateMainMenu:
		s := fmt.Sprintf("Оборудование: %s\nПродукт: %s\n\n", m.equipmentType, m.productType)
		s += "3) Главное меню:\n\n"
		for i, opt := range m.menuOptions {
			cursor := " "
			if i == m.menuIndex {
				cursor = "> "
			}
			s += fmt.Sprintf("%s%s\n", cursor, opt)
		}
		s += "\n↑/↓ — навигация; Enter — выбрать; q — назад\n"
		return s

	case stateShowConfig:
		s := fmt.Sprintf("Оборудование: %s\nПродукт: %s\n\n", m.equipmentType, m.productType)
		s += "4) Показать конфигурацию:\n\n"
		for _, line := range m.configLines {
			s += line + "\n"
		}
		return s

	case stateTestLEDs:
		return "\n<<< CONFIG: placeholder >>>\n\nq — назад\n"

	case stateShowLoad:
		return "\n<<< CONFIG: placeholder >>>\n\nq — назад\n"

	case stateFillReport:
		return "\n<<< CONFIG: placeholder >>>\n\nq — назад\n"
	}
	return "Unknown state\n"
}

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)
	if err := p.Start(); err != nil {
		fmt.Println("Ошибка запуска:", err)
		os.Exit(1)
	}
}
