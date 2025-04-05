package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type model struct {
	cpuUsage float64
	memUsed  uint64
	memTotal uint64
	err      error
}

type metricsMsg struct {
	cpu float64
	mem *mem.VirtualMemoryStat
	err error
}

func (m model) Init() tea.Cmd {
	return nil
}

func fetchMetrics() tea.Msg {
	cpuPercent, err1 := cpu.Percent(time.Second, false)
	memStats, err2 := mem.VirtualMemory()

	if err1 != nil || err2 != nil {
		return metricsMsg{err: fmt.Errorf("ошибка: %v %v", err1, err2)}
	}

	return metricsMsg{
		cpu: cpuPercent[0],
		mem: memStats,
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	
	case metricsMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.cpuUsage = msg.cpu
		m.memUsed = msg.mem.Used / (1024 * 1024)
		m.memTotal = msg.mem.Total / (1024 * 1024)
		return m, tea.Tick(time.Second*2, func(time.Time) tea.Msg {
			return fetchMetrics()
		})

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Произошла ошибка: %v\nНажми q для выхода", m.err)
	}

	return fmt.Sprintf(
		"Мониторинг сервера\n\nCPU загрузка: %.2f%%\nRAM: %d МБ\nТотал: %d\n\nНажми q для выхода.",
		m.cpuUsage, m.memUsed, m.memTotal,
	)
}

func main() {
	p := tea.NewProgram(model{})

	if err := p.Start(); err != nil {
		fmt.Println("Ошибка запуска:", err)
		os.Exit(1)
	}
}
