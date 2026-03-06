package screens

import (
	"fmt"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PowerAction string

const (
	ActionShutdown PowerAction = "Shutdown"
	ActionReboot   PowerAction = "Reboot"
)

type powerActionScreen struct {
	action    PowerAction
	seconds   int
	width     int
	height    int
	executing bool
	err       string
}

// Объявляем тип сообщения для обработки ошибок
type powerErrorMsg string

func NewPowerActionScreen(action PowerAction) powerActionScreen {
	return powerActionScreen{
		action:  action,
		seconds: 10,
	}
}

func (p powerActionScreen) Init() tea.Cmd {
	return nil
}

func (p powerActionScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		if p.executing || p.err != "" {
			return p, nil
		}
		if p.seconds > 0 {
			p.seconds--
			return p, nil
		}
		p.executing = true
		return p, p.executeAction()

	case powerErrorMsg:
		p.executing = false
		p.err = string(msg)
		return p, nil

	case tea.KeyMsg:
		// Обработка ESC для возврата в меню (Q обработает root.go)
		if msg.String() == "esc" {
			return NewMenuScreen(), tea.Batch(
				func() tea.Msg { return tea.WindowSizeMsg{Width: p.width, Height: p.height} },
				tea.ClearScreen,
			)
		}
	}
	return p, nil
}

func (p powerActionScreen) executeAction() tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		if p.action == ActionShutdown {
			cmd = exec.Command("systemctl", "poweroff", "--force")
		} else {
			cmd = exec.Command("systemctl", "reboot", "--force")
		}

		out, err := cmd.CombinedOutput()
		if err != nil {
			// Запасной вариант если systemctl не сработал
			fallback := "reboot"
			if p.action == ActionShutdown {
				fallback = "poweroff"
			}
			cmd = exec.Command(fallback, "-f")
			out, err = cmd.CombinedOutput()
		}

		if err != nil {
			return powerErrorMsg(fmt.Sprintf("ERROR: %v\nOutput: %s", err, string(out)))
		}

		time.Sleep(2 * time.Second)
		return tea.Quit()
	}
}

func (p powerActionScreen) View() string {
	if p.err != "" {
		return lipgloss.Place(p.width, p.height, lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true).
				Render(p.err+"\n\nPress ESC/Q to return"))
	}

	status := fmt.Sprintf("System %s in: %d sec", p.action, p.seconds)
	if p.executing {
		status = "SENDING SYSTEM SIGNAL..."
	}

	content := fmt.Sprintf("!!! WARNING !!!\n\n%s\n\nPress ESC/Q to CANCEL", status)

	return lipgloss.Place(p.width, p.height, lipgloss.Center, lipgloss.Center,
		lipgloss.NewStyle().Bold(true).Render(content),
	)
}