// internal/screens/disks.go
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
	loaded   bool
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
		d.viewport.SetContent(d.renderContent())

	case TickMsg:
		d.data = models.SystemInfo(msg)
		d.loaded = true
		if d.ready {
			d.viewport.SetContent(d.renderContent())
		}
	}

	d.viewport, cmd = d.viewport.Update(msg)
	return d, cmd
}

func (d disksScreen) View() string {
	return fmt.Sprintf("%s\n%s\n%s",
		"Disks information\n",
		d.viewport.View(),
		"\n(↑/↓)/(k/j): Scroll | Q: Menu")
}

func parentBlockDevice(devPath string) string {
	base := strings.TrimPrefix(devPath, "/dev/")
	if base == "" {
		return devPath
	}

	if strings.HasPrefix(base, "nvme") || strings.HasPrefix(base, "mmcblk") {
		if i := strings.LastIndex(base, "p"); i > 0 {
			return "/dev/" + base[:i]
		}
	}

	i := len(base)
	for i > 0 && base[i-1] >= '0' && base[i-1] <= '9' {
		i--
	}
	if i > 0 {
		return "/dev/" + base[:i]
	}

	return devPath
}

func (d disksScreen) renderContent() string {
	if !d.loaded {
		return "Loading Disks information..."
	}

	partsByParent := make(map[string][]models.DiskInfo)
	for _, p := range d.data.Disks {
		parent := parentBlockDevice(p.Device)
		partsByParent[parent] = append(partsByParent[parent], p)
	}

	var s strings.Builder

	for _, bd := range d.data.BlockDevs {
		s.WriteString(fmt.Sprintf("%s  |  Model: %s [%s]  |  Size: %s\n",
			bd.Device, bd.Model, bd.Type, formatBytes(bd.Size),
		))

		if bd.Wear >= 0 {
			wearText := fmt.Sprintf("%.0f%%", bd.Wear)
			wearStyleLocal := okStyle
			if bd.Wear > 80 {
				wearStyleLocal = critStyle
			} else if bd.Wear > 50 {
				wearStyleLocal = warnStyle
			}

			s.WriteString(fmt.Sprintf("  Wear: %s  Status: %s  Cycles: %d  Hours: %d  Unsafe: %d\n",
				wearStyleLocal.Render(wearText),
				bd.Status,
				bd.PowerCycles,
				bd.PowerOnHours,
				bd.UnsafeShutdowns,
			))
		} else {
			s.WriteString("  Wear: N/A (Run as sudo or unsupported)\n")
		}

		s.WriteString("\n")

		parts := partsByParent[bd.Device]
		if len(parts) == 0 {
			s.WriteString("  (no partitions found)\n\n")
			continue
		}

		for _, p := range parts {
			s.WriteString(fmt.Sprintf("  Part: %s  |  Mount: %s\n", p.Device, p.Mountpoint))
			s.WriteString(fmt.Sprintf("    Usage: %s / %s (%.1f%%)\n",
				formatBytes(p.Used), formatBytes(p.Total), p.UsedPercent,
			))
			s.WriteString("    " + renderProgressBar(p.UsedPercent, 40) + "\n\n")
		}

		s.WriteString("\n")
	}

	return s.String()
}
