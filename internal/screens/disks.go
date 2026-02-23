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
		"\n(↑/↓)/(k/j): Scroll | Q: Menu")
}

func parentBlockDevice(devPath string) string {
	// devPath: "/dev/sda1" -> "/dev/sda"
	// "/dev/nvme0n1p2" -> "/dev/nvme0n1"
	// "/dev/mmcblk0p1" -> "/dev/mmcblk0"

	base := strings.TrimPrefix(devPath, "/dev/")
	if base == "" {
		return devPath
	}

	// nvme0n1p1 / mmcblk0p1: режем по последней 'p'
	if strings.HasPrefix(base, "nvme") || strings.HasPrefix(base, "mmcblk") {
		if i := strings.LastIndex(base, "p"); i > 0 {
			return "/dev/" + base[:i]
		}
	}

	// обычные sda1 / vda2: убираем хвостовые цифры
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
	if len(d.data.BlockDevs) == 0 && len(d.data.Disks) == 0 {
		return "Loading Disks data...\n\nPress ESC to return to menu."
	}

	// группируем партиции по родителю
	partsByParent := make(map[string][]models.DiskInfo)
	for _, p := range d.data.Disks {
		parent := parentBlockDevice(p.Device)
		partsByParent[parent] = append(partsByParent[parent], p)
	}

	var s strings.Builder

	// сначала печатаем физдиски (как lsblk верхний уровень)
	for _, bd := range d.data.BlockDevs {
		s.WriteString(fmt.Sprintf("%s  |  Model: %s [%s]  |  Size: %s\n",
			bd.Device, bd.Model, bd.Type, formatBytes(bd.Size),
		))

		// под ними партиции
		parts := partsByParent[bd.Device]
		if len(parts) == 0 {
			s.WriteString("  (no partitions found)\n\n")
			continue
		}

		for _, disk := range parts {
			header := fmt.Sprintf("  Part: %s  |  Mount: %s", disk.Device, disk.Mountpoint)
			s.WriteString(header + "\n")

			usageLine := fmt.Sprintf("    Usage: %s / %s (%.1f%%)",
				formatBytes(disk.Used),
				formatBytes(disk.Total),
				disk.UsedPercent)
			s.WriteString(usageLine + "\n")
			s.WriteString("    " + renderProgressBar(disk.UsedPercent, 40) + "\n")

			// Wear показывать логично только для NVMe, но у тебя он лежит в DiskInfo уже.
			if disk.Wear >= 0 {
				wearText := fmt.Sprintf("%.0f%%", disk.Wear)
				wearStyleLocal := okStyle
				if disk.Wear > 80 {
					wearStyleLocal = critStyle
				} else if disk.Wear > 50 {
					wearStyleLocal = warnStyle
				}
				s.WriteString(fmt.Sprintf("    Wear: %s  Status: %s  Cycles: %d  Hours: %d  Unsafe: %d\n",
					wearStyleLocal.Render(wearText),
					disk.Status,
					disk.PowerCycles,
					disk.PowerOnHours,
					disk.UnsafeShutdowns,
				))
			} else {
				s.WriteString("    Wear: N/A (Run as sudo or unsupported)\n")
			}

			s.WriteString("\n")
		}

		s.WriteString("\n")
	}

	// если вдруг есть партиции, чей родитель не попал в BlockDevs (редкие случаи), можно допечатать их отдельно
	return s.String()
}
