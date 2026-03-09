package main

import (
	"fmt"
	"os"

	"server-checker/internal/screens"

	tea "github.com/charmbracelet/bubbletea"
)

// Инициализация корневого экрана для управления всеми экранами
func main() {
	p := tea.NewProgram(screens.NewRoot(), tea.WithAltScreen()) // WithAltScreen - работа в отдельном слое терминала, как у vim

	if _, err := p.Run(); err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
	}
}
