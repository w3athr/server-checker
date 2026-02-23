package screens

import (
	"fmt"
	"server-checker/internal/models"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type cpuScreen struct {
	data models.SystemInfo
}

func NewCPUScreen() cpuScreen {
	return cpuScreen{}
}

func (c cpuScreen) Init() tea.Cmd {
	return nil
}

func (c cpuScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		c.data = models.SystemInfo(msg)
		return c, nil
	}
	return c, nil
}

func (c cpuScreen) View() string {
	if c.data.CPU.Model == "" {
		return "Loading CPU data...\n\nPress Q to return to menu."
	}

	var s strings.Builder

	s.WriteString("CPU and MotherBoard Information\n\n")

	s.WriteString(fmt.Sprintf("Model: %s\n", c.data.CPU.Model))
	s.WriteString(fmt.Sprintf("Cores: %d Cores / %d Threads\n", c.data.CPU.Cores, c.data.CPU.Threads))
	s.WriteString(fmt.Sprintf("Speed: %.2f GHz\n", c.data.CPU.Speed))

	if c.data.CPU.Motherboard != "" {
		s.WriteString(fmt.Sprintf("\nMotherBoard: %s\n", c.data.CPU.Motherboard))
	} else {
		s.WriteString("\nMotherBoard: N/A\n")
	}

	if c.data.CPU.MotherboardSerial != "" {
		s.WriteString(fmt.Sprintf("Serial Number: %s\n", c.data.CPU.MotherboardSerial))
	} else {
		s.WriteString("Serial Number: N/A\n")
	}

	load := c.data.CPU.LoadPercent
	bar := renderProgressBar(load, 30)

	s.WriteString("\nCurrent Load: " + fmt.Sprintf("%.2f%%", load) + "\n")
	s.WriteString(bar + "\n\n")

	s.WriteString("Q: Menu")

	return s.String()
}
