package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	stateSelectEquip   = "selectEquip"
	stateSelectProduct = "selectProduct"
	stateMainMenu      = "mainMenu"
)

type model struct {
	state         string
	equipOptions  []string
	equipIndex    int
	equipmentType string
	
	productOptions []string
	productIndex   int
	productType    string
}

func initialModel() model {
	return model{
		state:        stateSelectEquip,
		equipOptions: []string{"рубеж-с", "рубеж-т", "рубеж-н", "бокс", "тринити", "супермикро"},
		equipIndex:   0,

		productOptions: []string{"ngfw", "ifw"},
		productIndex:   0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		if key.String() == "q" || key.String() == "ctrl+c" {
			return m, tea.Quit
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
		s := fmt.Sprintf("Вы выбрали оборудование: %s\n\n", m.equipmentType)
		s += "2) Выберите продукт:\n\n"
		for i, opt := range m.productOptions {
			cursor := " "
			if i == m.productIndex {
				cursor = "> "
			}
			s += fmt.Sprintf("%s%s\n", cursor, opt)
		}
		s += "\n ↑/↓ (k/j) - навигация, Enter — подтвердить, q - выход.\n" 
		return s
		
	case stateMainMenu:
		return fmt.Sprintf(
			"Оборудование: %s\nПродукт: %s\n\n3) Главное меню (здесь будут пункты 1-4)\n\nq - выход\n", 
			m.equipmentType, m.productType,
		)
	}
	return "Неизвестное состояние\n"
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Println("Ошибка запуска:", err)
		os.Exit(1)
	}
}
