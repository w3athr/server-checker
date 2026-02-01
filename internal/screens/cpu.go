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
	if c.data.CPU.Model == "" { // если данные еще не пришли
		return "Loading CPU data...\n\nPress ESC to return to menu."
	}

	var s strings.Builder

	s.WriteString("CPU Information" + "\n\n")

	s.WriteString(fmt.Sprintf("%s: %s\n", "Model", c.data.CPU.Model))
	s.WriteString(fmt.Sprintf("%s %d Cores / %d Threads\n", "Cores:", c.data.CPU.Cores, c.data.CPU.Threads))
	s.WriteString(fmt.Sprintf("%s %.2f GHz\n", "Speed:", c.data.CPU.Speed))

	load := c.data.CPU.LoadPercent // простой прогресс-бар для загрузки
	bar := renderProgressBar(load, 30)

	s.WriteString("\n" + "Current Load: " + fmt.Sprintf("%.2f%%", load) + "\n")
	s.WriteString(bar + "\n\n")

	s.WriteString("Press ESC to return to menu.")

	return s.String()
}
