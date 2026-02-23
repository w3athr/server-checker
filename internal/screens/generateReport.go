package screens

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"server-checker/internal/collector"
	"server-checker/internal/models"
	"server-checker/internal/templates"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"
)

// generateReportScreen - экран для генерации отчета.
type generateReportScreen struct {
	status string // Generating, Success, Error
	err    error
	done   bool
	width  int
	height int
}

type reportResultMsg struct {
	path string
	err  error
}

type checkResult struct {
	Name     string `yaml:"parameter"`
	Expected string `yaml:"expected"`
	Actual   string `yaml:"actual"`
	Passed   bool   `yaml:"is_passed"`
}

type yamlReport struct {
	TemplateName string            `yaml:"template_name"`
	Timestamp    string            `yaml:"timestamp"`
	Summary      string            `yaml:"summary"`
	Checks       []checkResult     `yaml:"checks"`
	RawData      models.SystemInfo `yaml:"raw_system_data"`
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
	case tea.WindowSizeMsg:
		g.width = msg.Width
		g.height = msg.Height
		return g, nil

	case reportResultMsg:
		g.done = true
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
				return NewReportMenuScreen(), tea.Batch(
					tea.ClearScreen,
					func() tea.Msg { return tea.WindowSizeMsg{Width: g.width, Height: g.height} },
				)
			}
		} else {
			if msg.String() == "esc" {
				return NewReportMenuScreen(), tea.Batch(
					tea.ClearScreen,
					func() tea.Msg { return tea.WindowSizeMsg{Width: g.width, Height: g.height} },
				)
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
	// Получаем выбранный шаблон
	var tmpl models.SystemTemplate
	var ok bool

	if SelectedTemplateName == "" {
		return reportResultMsg{err: fmt.Errorf("no template selected")}
	}

	if SelectedTemplateName == "Custom" {
		data, err := os.ReadFile(SelectedTemplatePath)
		if err != nil {
			return reportResultMsg{err: fmt.Errorf("failed to read template file: %w", err)}
		}

		if err := yaml.Unmarshal(data, &tmpl); err != nil {
			return reportResultMsg{err: fmt.Errorf("invalid YAML template: %w", err)}
		}
	} else {
		tmpl, ok = templates.GetEmbeddedTemplate(SelectedTemplateName)
		if !ok {
			return reportResultMsg{err: fmt.Errorf("unknown template: %s", SelectedTemplateName)}
		}
	}

	currentData, err := collector.CollectAll()
	if err != nil {
		return reportResultMsg{err: fmt.Errorf("failed to collect system data: %w", err)}
	}

	dirName := fmt.Sprintf("HealthCheck-%s", time.Now().Format("2006-01-02_15-04-05"))
	absPath, err := filepath.Abs(dirName)
	if err != nil {
		return reportResultMsg{err: fmt.Errorf("failed to get absolute path: %w", err)}
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return reportResultMsg{err: fmt.Errorf("failed to create directory: %w", err)}
	}

	if err := generateDiagnosticsLogs(absPath); err != nil {
		return reportResultMsg{err: err}
	}

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

func runHealthCheck(tmpl models.SystemTemplate, data models.SystemInfo) []checkResult {
	var results []checkResult

	// 1) CPU Cores
	results = append(results, checkResult{
		Name:     "CPU Cores",
		Expected: fmt.Sprintf(">= %d", tmpl.CPU.MinCores),
		Actual:   fmt.Sprintf("%d", data.CPU.Cores),
		Passed:   data.CPU.Cores >= tmpl.CPU.MinCores,
	})

	// 2) RAM
	totalRAMGB := int(data.RAM.Total / 1024 / 1024 / 1024)
	results = append(results, checkResult{
		Name:     "Total RAM",
		Expected: fmt.Sprintf(">= %d GB", tmpl.RAM.MinTotalGB),
		Actual:   fmt.Sprintf("%d GB", totalRAMGB),
		Passed:   totalRAMGB >= tmpl.RAM.MinTotalGB,
	})

	// 3) Disks: проверяем по физическим блочным устройствам (BlockDevs), а не по партициям (Disks)
	for _, dt := range tmpl.Disks {
		found := false
		actualSizeGB := 0

		for _, bd := range data.BlockDevs {
			sizeGB := int(bd.Size / 1024 / 1024 / 1024)
			if bd.Type == dt.Type && sizeGB >= dt.MinSizeGB {
				found = true
				actualSizeGB = sizeGB
				break
			}
		}

		statusName := fmt.Sprintf("Disk: %s (%s)", dt.Name, dt.Type)
		results = append(results, checkResult{
			Name:     statusName,
			Expected: fmt.Sprintf(">= %d GB", dt.MinSizeGB),
			Actual:   fmt.Sprintf("%d GB", actualSizeGB),
			Passed:   found,
		})
	}

	return results
}

func generateTXTReport(dir string, tmpl models.SystemTemplate, data models.SystemInfo) error {
	results := runHealthCheck(tmpl, data)

	var sb strings.Builder
	sb.WriteString("==========================================\n")
	sb.WriteString(fmt.Sprintf("HEALTH CHECK REPORT: %s\n", tmpl.Name))
	sb.WriteString(fmt.Sprintf("Date: %s\n", time.Now().Format(time.RFC1123)))
	sb.WriteString("==========================================\n\n")

	for _, r := range results {
		status := "[ PASS ]"
		if !r.Passed {
			status = "[ FAIL ]"
		}
		sb.WriteString(fmt.Sprintf("%-20s %-10s (Exp: %-10s, Act: %-10s)\n",
			r.Name, status, r.Expected, r.Actual))
	}

	return os.WriteFile(filepath.Join(dir, "report.txt"), []byte(sb.String()), 0644)
}

func generateYAMLReport(dir string, tmpl models.SystemTemplate, data models.SystemInfo) error {
	results := runHealthCheck(tmpl, data)

	globalPass := true
	for _, r := range results {
		if !r.Passed {
			globalPass = false
			break
		}
	}

	summary := "PASS"
	if !globalPass {
		summary = "FAIL"
	}

	report := yamlReport{
		TemplateName: tmpl.Name,
		Timestamp:    time.Now().Format(time.RFC3339),
		Summary:      summary,
		Checks:       results,
		RawData:      data,
	}

	bytes, err := yaml.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return os.WriteFile(filepath.Join(dir, "report.yaml"), bytes, 0644)
}

func generateHTMLReport(dir string, tmpl models.SystemTemplate, data models.SystemInfo) error {
	results := runHealthCheck(tmpl, data)

	htmlHeader := `
	<html>
	<head>
		<meta charset="UTF-8">
		<style>
			body { font-family: sans-serif; margin: 40px; background: #f4f4f9; }
			table { width: 100%; border-collapse: collapse; background: white; }
			th, td { padding: 12px; border: 1px solid #ddd; text-align: left; }
			th { background-color: #4CAF50; color: white; }
			.pass { color: green; font-weight: bold; }
			.fail { color: red; font-weight: bold; }
			.header { margin-bottom: 20px; }
		</style>
	</head>
	<body>
		<div class="header">
			<h1>Health Check: ` + tmpl.Name + `</h1>
			<p>Generated on: ` + time.Now().Format(time.RFC1123) + `</p>
		</div>
		<table>
			<tr>
				<th>Parameter</th>
				<th>Expected</th>
				<th>Actual</th>
				<th>Status</th>
			</tr>`

	var rows string
	for _, r := range results {
		statusClass := "pass"
		statusText := "OK"
		if !r.Passed {
			statusClass = "fail"
			statusText = "FAIL"
		}
		rows += fmt.Sprintf(`
			<tr>
				<td>%s</td>
				<td>%s</td>
				<td>%s</td>
				<td class="%s">%s</td>
			</tr>`, r.Name, r.Expected, r.Actual, statusClass, statusText)
	}

	footer := `</table></body></html>`

	return os.WriteFile(filepath.Join(dir, "report.html"), []byte(htmlHeader+rows+footer), 0644)
}

/*
	========================
	NEW: diagnostics logs
	========================
*/

type multiError struct {
	errs []error
}

func (m *multiError) add(err error) {
	if err != nil {
		m.errs = append(m.errs, err)
	}
}

func (m *multiError) Err() error {
	if len(m.errs) == 0 {
		return nil
	}
	var sb strings.Builder
	sb.WriteString("diagnostics log collection failed:\n")
	for i, e := range m.errs {
		sb.WriteString(fmt.Sprintf("  %d) %v\n", i+1, e))
	}
	return fmt.Errorf("%s", sb.String())
}

func generateDiagnosticsLogs(reportDir string) error {
	logsDir := filepath.Join(reportDir, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs dir: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	var merr multiError

	merr.add(writeCommandOutput(ctx, logsDir, "dmesg.txt", "dmesg"))
	merr.add(writeCommandOutput(ctx, logsDir, "lspci-k.txt", "lspci", "-k"))
	merr.add(writeCommandOutput(ctx, logsDir, "lsusb.txt", "lsusb"))
	merr.add(writeCommandOutput(ctx, logsDir, "lsblk.txt", "lsblk"))

	merr.add(writeFileContents(logsDir, "sys_vendor.txt", "/sys/class/dmi/id/sys_vendor"))
	merr.add(writeFileContents(logsDir, "product_name.txt", "/sys/class/dmi/id/product_name"))
	merr.add(writeFileContents(logsDir, "product_serial.txt", "/sys/class/dmi/id/product_serial"))

	return merr.Err()
}

func writeCommandOutput(ctx context.Context, dir, outFile, name string, args ...string) error {
	fullPath := filepath.Join(dir, outFile)

	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	var sb strings.Builder
	sb.WriteString("==========================================\n")
	sb.WriteString("COMMAND OUTPUT\n")
	sb.WriteString(fmt.Sprintf("Timestamp: %s\n", start.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("Command: %s\n", strings.Join(append([]string{name}, args...), " ")))
	sb.WriteString(fmt.Sprintf("Duration: %s\n", duration))
	sb.WriteString("==========================================\n\n")

	if stdout.Len() > 0 {
		sb.WriteString("----- STDOUT -----\n")
		sb.Write(stdout.Bytes())
		if !strings.HasSuffix(stdout.String(), "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	if stderr.Len() > 0 {
		sb.WriteString("----- STDERR -----\n")
		sb.Write(stderr.Bytes())
		if !strings.HasSuffix(stderr.String(), "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	if err != nil {
		sb.WriteString("----- ERROR -----\n")
		sb.WriteString(err.Error())
		sb.WriteString("\n")
	}

	if werr := os.WriteFile(fullPath, []byte(sb.String()), 0644); werr != nil {
		return fmt.Errorf("failed to write %s: %w", fullPath, werr)
	}

	if err != nil {
		return fmt.Errorf("command failed (%s): %w", outFile, err)
	}
	return nil
}

func writeFileContents(dir, outFile, srcPath string) error {
	fullPath := filepath.Join(dir, outFile)

	b, err := os.ReadFile(srcPath)
	now := time.Now()

	var sb strings.Builder
	sb.WriteString("==========================================\n")
	sb.WriteString("FILE CONTENTS\n")
	sb.WriteString(fmt.Sprintf("Timestamp: %s\n", now.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("Source: %s\n", srcPath))
	sb.WriteString("==========================================\n\n")

	if err != nil {
		sb.WriteString("----- ERROR -----\n")
		sb.WriteString(err.Error())
		sb.WriteString("\n")
		if werr := os.WriteFile(fullPath, []byte(sb.String()), 0644); werr != nil {
			return fmt.Errorf("failed to write %s: %w", fullPath, werr)
		}
		return fmt.Errorf("failed to read %s: %w", srcPath, err)
	}

	sb.Write(b)
	if len(b) > 0 && b[len(b)-1] != '\n' {
		sb.WriteString("\n")
	}

	if werr := os.WriteFile(fullPath, []byte(sb.String()), 0644); werr != nil {
		return fmt.Errorf("failed to write %s: %w", fullPath, werr)
	}
	return nil
}
