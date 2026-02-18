package screens

import (
	"fmt"
	"server-checker/internal/collector"
	"server-checker/internal/models"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type networkScreen struct {
	data     models.SystemInfo
	cursor   int
	viewport viewport.Model
	ready    bool
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
	if !n.ready {
		return "Initializing Network..."
	}

	return fmt.Sprintf("%s\n%s\n%s",
		"NIC information\n",
		n.viewport.View(),
		"(↑/↓)/(k/j): Select | L: Blink LED (10s) | Q: Menu")
}

func (n networkScreen) renderContent() string {
	if len(n.data.Network) == 0 {
		return "Loading Network data..."
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
		card += fmt.Sprintf("   Model: %s\n", net.Model)
		card += fmt.Sprintf("   MAC:   %s\n", net.MAC)

		speedStr := "N/A"
		if net.Speed > 0 {
			speedStr = fmt.Sprintf("%d Mbps", net.Speed)
		}
		card += fmt.Sprintf("   Speed: %s", speedStr)

		s.WriteString(style.Render(card) + "\n\n")
	}

	return s.String()
}

func (n *networkScreen) syncViewport() {
	itemHeight := 8 // Количество строк, которое занимает один интерфейс (4 текста + 2 отступа)

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
