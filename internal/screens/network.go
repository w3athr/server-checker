package screens

import (
	"bytes"
	"fmt"
	"server-checker/internal/collector"
	"server-checker/internal/models"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type networkScreen struct {
	data     models.SystemInfo
	cursor   int
	viewport viewport.Model
	ready    bool
	loaded   bool
}

func NewNetworkScreen() networkScreen {
	return networkScreen{}
}

func (n networkScreen) Init() tea.Cmd {
	return nil
}

func (n networkScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight, footerHeight := 3, 2
		if !n.ready {
			n.viewport = viewport.New(msg.Width, msg.Height-headerHeight-footerHeight)
			n.ready = true
		} else {
			n.viewport.Width = msg.Width
			n.viewport.Height = msg.Height - headerHeight - footerHeight
		}
		n.viewport.SetContent(n.renderContent())

	case TickMsg:
		n.data = models.SystemInfo(msg)
		n.loaded = true
		n.viewport.SetContent(n.renderContent())

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if n.cursor > 0 {
				n.cursor--
				n.syncViewport()
				n.viewport.SetContent(n.renderContent())
			}
		case "down", "j":
			if n.cursor < len(n.data.Network)-1 {
				n.cursor++
				n.syncViewport()
				n.viewport.SetContent(n.renderContent())
			}
		case "l":
			if len(n.data.Network) > 0 {
				ifaceName := n.data.Network[n.cursor].Interface
				cmds = append(cmds, func() tea.Msg {
					collector.BlinkInterface(ifaceName)
					return nil
				})
			}
		}
	}
	n.viewport, cmd = n.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return n, tea.Batch(cmds...)
}

func (n networkScreen) View() string {
	return fmt.Sprintf("%s\n%s\n%s",
		"NIC information\n",
		n.viewport.View(),
		"(↑/↓)/(k/j): Select | L: Blink LED (5s) | Q: Menu")
}

func (n networkScreen) renderContent() string {
	if !n.loaded {
		return "Loading Network data..."
	}

	if len(n.data.Network) == 0 {
		return "No Ethernet interfaces detected on this system."
	}

	var s strings.Builder

	for i, net := range n.data.Network {
		// Подсветка выбранного интерфейса
		cursorChar := "  "
		style := lipgloss.NewStyle().PaddingLeft(2)
		if n.cursor == i {
			cursorChar = "> "
			style = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(lipgloss.Color("#7D56F4")).
				PaddingLeft(1).
				Background(lipgloss.Color("#222222"))
		}

		// Цветовой индикатор статуса
		statusDot := "●"
		statusColor := "#FF0000" // Down
		if strings.ToLower(net.Status) == "up" {
			statusColor = "#00FF00" // Up
		}
		coloredDot := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Render(statusDot)

		// Формируем карточку интерфейса
		card := fmt.Sprintf("%s [%s] %s %s\n", cursorChar, net.Interface, coloredDot, net.Status)
		var buf bytes.Buffer
		tw := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

		fmt.Fprintf(tw, "   Model:\t%s\n", net.Model)
		fmt.Fprintf(tw, "   MAC:\t%s\n", net.MAC)

		cur := "N/A"
		if net.Speed > 0 {
			cur = fmt.Sprintf("%d Mbps", net.Speed)
		}
		fmt.Fprintf(tw, "   CurrentSpeed:\t%s\n", cur)

		max := "N/A"
		if net.MaxSpeedMbps > 0 {
			max = fmt.Sprintf("%d Mbps", net.MaxSpeedMbps)
		}
		fmt.Fprintf(tw, "   MaxSpeed:\t%s\n", max)

		fmt.Fprintf(tw, "   IP:\t%s\n", net.IP)

		_ = tw.Flush()
		card += buf.String()
		s.WriteString(style.Render(card) + "\n\n")
	}

	return s.String()
}

func (n *networkScreen) syncViewport() {
	itemHeight := 10 // Количество строк, которое занимает один интерфейс (4 текста + 2 отступа)

	// Вычисляем позицию верхней и нижней границы текущего элемента
	topOfItem := n.cursor * itemHeight
	bottomOfItem := topOfItem + itemHeight

	// Границы текущего "окна" вьюпорта
	viewportTop := n.viewport.YOffset
	viewportBottom := n.viewport.YOffset + n.viewport.Height

	// Если курсор ушел выше видимой области
	if topOfItem < viewportTop {
		n.viewport.YOffset = topOfItem
	}

	// Если курсор ушел ниже видимой области
	if bottomOfItem > viewportBottom {
		// Смещаем так, чтобы нижний край элемента был виден в самом низу
		n.viewport.YOffset = bottomOfItem - n.viewport.Height
	}
}
