package screens

import (
	"fmt"
	"server-checker/internal/models"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type disksScreen struct {
	data     models.SystemInfo
	viewport viewport.Model
	ready    bool
}

func NewDisksScreen() disksScreen {
	return disksScreen{}
}

func (d disksScreen) Init() tea.Cmd {
	return nil
}

func (d disksScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := 3
		footerHeight := 2

		if !d.ready {
			d.viewport = viewport.New(msg.Width, msg.Height-headerHeight-footerHeight)
			d.ready = true
		} else {
			d.viewport.Width = msg.Width
			d.viewport.Height = msg.Height - headerHeight - footerHeight
		}

	case TickMsg:
		d.data = models.SystemInfo(msg)
		d.viewport.SetContent(d.renderContent())
	}

	d.viewport, cmd = d.viewport.Update(msg)
	return d, cmd
}

func (d disksScreen) View() string {
	if !d.ready {
		return "Initializing Disks..."
	}
	return fmt.Sprintf("%s\n%s\n%s",
		"Disks information\n",
		d.viewport.View(),
		"\n(↑/↓)/(k/j): Scroll | ESC: Menu")
}

func (d disksScreen) renderContent() string {
	if len(d.data.Disks) == 0 {
		return "Loading Disks data...\n\nPress ESC to return to menu."
	}

	var s strings.Builder

	for _, disk := range d.data.Disks {
		// Заголовок: Путь и Точка монтирования
		header := fmt.Sprintf("ID: %s  |  Mount: %s", disk.Device, disk.Mountpoint)
		s.WriteString(header + "\n")

		// Информация о железе
		modelInfo := fmt.Sprintf("  Model: %s [%s]", disk.Model, disk.Type)
		s.WriteString(modelInfo + "\n")

		// Прогресс-бар использования места
		usageLine := fmt.Sprintf("  Usage: %s / %s (%.1f%%)",
			formatBytes(disk.Used),
			formatBytes(disk.Total),
			disk.UsedPercent)
		s.WriteString(usageLine + "\n")
		s.WriteString("  " + renderProgressBar(disk.UsedPercent, 40) + "\n")

		// Износ (Wear) - Самое важное для Олега
		if disk.Wear >= 0 {
			wearText := fmt.Sprintf("%.0f%%", disk.Wear)
			wearStyleLocal := okStyle
			if disk.Wear > 80 {
				wearStyleLocal = critStyle
			} else if disk.Wear > 50 {
				wearStyleLocal = warnStyle
			}
			s.WriteString(fmt.Sprintf("  Wear Level: %s (Status: %s)\n",
				wearStyleLocal.Render(wearText),
				disk.Status))
		} else {
			s.WriteString("  Wear Level: N/A (Run as sudo or unsupported)" + "\n")
		}

		s.WriteString("\n")
	}

	return s.String()
}
