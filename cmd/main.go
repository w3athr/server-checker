package main

import (
	"fmt"
	"os"

	"server-checker/internal/screens"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(screens.NewRoot(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
	}
}
