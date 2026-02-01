package screens

import (
	"fmt"
	"server-checker/internal/models"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type ramScreen struct {
	data models.SystemInfo
}

func NewRAMScreen() ramScreen {
	return ramScreen{}
}

func (r ramScreen) Init() tea.Cmd {
	return nil
}

func (r ramScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		r.data = models.SystemInfo(msg)
	}
	return r, nil
}

func (r ramScreen) View() string {
	if r.data.RAM.Total == 0 { // если данные еще не пришли
		return "Loading RAM data...\n\nPress ESC to return to menu."
	}

	var s strings.Builder

	const gb = 1024 * 1024 * 1024 // для перевода байтов в гигабайты

	s.WriteString("RAM Information" + "\n\n")

	// Общая информация (общий, занято, свободно)
	totalGB := formatBytes(r.data.RAM.Total)
	usedGB := formatBytes(r.data.RAM.Used)
	freeGB := formatBytes(r.data.RAM.Free)

	fmt.Fprintf(&s, "Total: %.2f GB\n", totalGB)
	fmt.Fprintf(&s, "Used: %.2f GB (%.1f%%)\n", usedGB, r.data.RAM.UsedPercent)
	fmt.Fprintf(&s, "Free: %.2f GB\n", freeGB)

	// прогресс-бар использования RAM
	s.WriteString("Current Usage:\n")
	s.WriteString(renderProgressBar(r.data.RAM.UsedPercent, 30) + "\n\n")

	// список физ планок
	s.WriteString("Physical Memory Sticks:\n")
	if len(r.data.RAM.Sticks) == 0 {
		s.WriteString("No data available (commonly in VMs/WSL)\n")
	} else {
		for i, stick := range r.data.RAM.Sticks {
			stickGB := float64(stick.Capacity) / gb
			s.WriteString(fmt.Sprintf("[%d] Slot: %s | Capacity: %.2f GB | Vendor: %s | Speed: %d Mhz\n", i, stick.Slot, stickGB, stick.Model, stick.Speed))
		}
	}
	s.WriteString("\nPress ESC to return to menu.")
	return s.String()
}
