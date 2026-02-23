package screens

import (
	"fmt"
	"server-checker/internal/models"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type systemScreen struct {
	data models.SystemInfo
}

func NewSystemScreen() systemScreen {
	return systemScreen{}
}

func (s systemScreen) Init() tea.Cmd {
	return nil
}

func (s systemScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		s.data = models.SystemInfo(msg)
		return s, nil
	}
	return s, nil
}

func (s systemScreen) View() string {
	if s.data.Host.OS == "" {
		return "Loading System data...\n\nPress Q to return to menu."
	}

	var b strings.Builder
	b.WriteString("System Information\n\n")

	// OS / Kernel / Uptime
	b.WriteString(fmt.Sprintf("OS: %s\n", s.data.Host.OS))
	b.WriteString(fmt.Sprintf("Kernel: %s\n", s.data.Host.Kernel))
	b.WriteString(fmt.Sprintf("Uptime: %s\n", formatUptime(s.data.Host.Uptime)))

	// Hardware vendor/model/SN (живут в CPU, потому что это DMI/платформа)
	vendor := s.data.CPU.HardwareVendor
	model := s.data.CPU.HardwareModel
	sn := s.data.CPU.HardwareSN

	if vendor == "" {
		vendor = "N/A"
	}
	if model == "" {
		model = "N/A"
	}
	if sn == "" {
		sn = "N/A"
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Hardware vendor: %s\n", vendor))
	b.WriteString(fmt.Sprintf("Hardware model: %s\n", model))
	b.WriteString(fmt.Sprintf("Hardware S/N: %s\n", sn))

	b.WriteString("\nQ: Menu")

	return b.String()
}

func formatUptime(sec uint64) string {
	// простой человеко-читаемый формат: 3d 12h 05m 07s
	days := sec / 86400
	sec %= 86400
	hours := sec / 3600
	sec %= 3600
	mins := sec / 60
	sec %= 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, mins, sec)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, mins, sec)
	}
	if mins > 0 {
		return fmt.Sprintf("%dm %ds", mins, sec)
	}
	return fmt.Sprintf("%ds", sec)
}
