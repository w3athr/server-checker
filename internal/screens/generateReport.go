package screens

import (
	"fmt"
	"os"
	"path/filepath"
	"server-checker/internal/collector"
	"server-checker/internal/models"
	"server-checker/internal/templates"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"
)

// generateReportScreen - экран для генерации отчета.
type generateReportScreen struct {
	status string // Generating, Success, Error
	err    error
	done   bool
}

type reportResultMsg struct {
	path string
	err  error
}

func NewGenerateReportScreen() generateReportScreen {
	return generateReportScreen{
		status: "Generating report...",
		done:   false,
	}
}

func (g generateReportScreen) Init() tea.Cmd {
	return generateReportCmd
}

func (g generateReportScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case reportResultMsg:
		if msg.err != nil {
			g.status = fmt.Sprintf("Error: %v", msg.err)
			g.err = msg.err
		} else {
			g.status = fmt.Sprintf("Success! Report saved to:\n%s", msg.path)
		}
		return g, tea.ClearScreen

	case tea.KeyMsg:
		if g.done {
			if msg.String() == "esc" || msg.String() == "enter" {
				return NewReportMenuScreen(), tea.ClearScreen
			}
		} else {
			if msg.String() == "esc" {
				return NewReportMenuScreen(), tea.ClearScreen
			}
		}
	}
	return g, nil
}

func (g generateReportScreen) View() string {
	s := "Report Generation Status:\n\n"
	s += g.status + "\n\n"

	if g.done {
		s += "Press Esc/Enter to return to menu"
	} else {
		s += "Please wait..."
	}

	return s
}

func generateReportCmd() tea.Msg {
	// ОТЛАДКА: Пишем лог в файл, чтобы понять, где застревает
	f, _ := os.Create("debug_log.txt")
	defer f.Close()
	logger := func(msg string) {
		f.WriteString(time.Now().Format(time.RFC3339) + " " + msg + "\n")
	}

	logger("Starting generation...")

	// Получаем выбранный шаблон
	var tmpl models.SystemTemplate
	var ok bool

	if SelectedTemplateName == "" {
		return reportResultMsg{err: fmt.Errorf("no template selected")}
	}

	if SelectedTemplateName == "Custom" {
		// Читаем файл
		data, err := os.ReadFile(SelectedTemplatePath)
		if err != nil {
			return reportResultMsg{err: fmt.Errorf("failed to read template file: %w", err)}
		}

		if err := yaml.Unmarshal(data, &tmpl); err != nil {
			return reportResultMsg{err: fmt.Errorf("invalid YAML template: %w", err)}
		}
	} else {
		// Берем встроенный
		tmpl, ok = templates.GetEmbeddedTemplate(SelectedTemplateName)
		if !ok {
			return reportResultMsg{err: fmt.Errorf("unknown template: %s", SelectedTemplateName)}
		}
	}
	// Собираем текущие данные
	currentData, err := collector.CollectAll()
	if err != nil {
		return reportResultMsg{err: fmt.Errorf("failed to collect system data: %w", err)}
	}

	// Создаем директорию для отчета
	dirName := fmt.Sprintf("HealthCheck-%s", time.Now().Format("2006-01-02_15-04-05"))
	absPath, err := filepath.Abs(dirName)
	if err != nil {
		return reportResultMsg{err: fmt.Errorf("failed to get absolute path: %w", err)}
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return reportResultMsg{err: fmt.Errorf("failed to create directory: %w", err)}
	}

	// Генерируем отчеты
	if err := generateTXTReport(absPath, tmpl, currentData); err != nil {
		return reportResultMsg{err: err}
	}
	if err := generateYAMLReport(absPath, tmpl, currentData); err != nil {
		return reportResultMsg{err: err}
	}
	if err := generateHTMLReport(absPath, tmpl, currentData); err != nil {
		return reportResultMsg{err: err}
	}

	return reportResultMsg{path: absPath}
}

func generateTXTReport(dir string, tmpl models.SystemTemplate, data models.SystemInfo) error {
	content := fmt.Sprintf("Health Check Report\nTemplate: %s\nDate: %s\n\n", tmpl.Name, time.Now().Format(time.RFC1123))

	// Здесь будет логика сравнения...
	content += fmt.Sprintf("CPU Check:\n")
	content += fmt.Sprintf("  Template: Min %d Cores, %.1f GHz\n", tmpl.CPU.MinCores, tmpl.CPU.MinSpeed)
	content += fmt.Sprintf("  Actual:   %d Cores, %.1f GHz\n", data.CPU.Cores, data.CPU.Speed)

	// ... (добавить остальные проверки)

	return os.WriteFile(filepath.Join(dir, "report.txt"), []byte(content), 0644)
}

func generateYAMLReport(dir string, tmpl models.SystemTemplate, data models.SystemInfo) error {
	// Просто сохраняем текущие данные в YAML для начала
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "report.yaml"), bytes, 0644)
}

func generateHTMLReport(dir string, tmpl models.SystemTemplate, data models.SystemInfo) error {
	content := "<html><body><h1>Health Check Report</h1></body></html>" // Заглушка
	return os.WriteFile(filepath.Join(dir, "report.html"), []byte(content), 0644)
}
