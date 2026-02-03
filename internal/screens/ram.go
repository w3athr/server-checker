package screens

import (
	"fmt"
	"server-checker/internal/models"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type ramScreen struct {
	data     models.SystemInfo
	viewport viewport.Model
	ready    bool
	lastSize tea.WindowSizeMsg
}

func NewRAMScreen() ramScreen {
	return ramScreen{}
}

func (r ramScreen) Init() tea.Cmd {
	return nil
}

func (r ramScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.lastSize = msg
		headerHeight, footerHeight := 3, 2
		if !r.ready {
			r.viewport = viewport.New(msg.Width, msg.Height-headerHeight-footerHeight)
			r.ready = true
		} else {
			r.viewport.Width = msg.Width
			r.viewport.Height = msg.Height - headerHeight - footerHeight
		}

	case TickMsg:
		r.data = models.SystemInfo(msg)
		r.viewport.SetContent(r.renderContent())
	}

	r.viewport, cmd = r.viewport.Update(msg)
	return r, cmd
}

func (r ramScreen) renderContent() string {
	if r.data.RAM.Total == 0 { // если данные еще не пришли
		return "Loading RAM data...\n\nPress ESC to return to menu."
	}

	var s strings.Builder

	fmt.Fprintf(&s, "Total: %s\n", formatBytes(r.data.RAM.Total))
	fmt.Fprintf(&s, "Used: %s (%.1f%%)\n", formatBytes(r.data.RAM.Used), r.data.RAM.UsedPercent)
	fmt.Fprintf(&s, "Available: %s\n", formatBytes(r.data.RAM.Available))

	// прогресс-бар использования RAM
	s.WriteString("Current Usage:\n")
	s.WriteString(renderProgressBar(r.data.RAM.UsedPercent, 30) + "\n\n")

	// список физ планок
	s.WriteString("Physical Memory Sticks:\n")
	for _, stick := range r.data.RAM.Sticks {
		s.WriteString(fmt.Sprintf("%s | %s | %d Mhz\n",
			stick.Slot, formatBytes(stick.Capacity), stick.Speed))
		s.WriteString("  Model: " + stick.Model + "\n")
	}
	return s.String()
}

func (r ramScreen) View() string {
	if !r.ready {
		return "Initializing RAM..."
	}
	return fmt.Sprintf("%s\n%s\n%s",
		"RAM information\n",
		r.viewport.View(),
		"\n(↑/↓)/(k/j): Scroll | ESC: Menu")
}
