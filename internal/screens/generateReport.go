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

	// 1. CPU
	cpuPassed := data.CPU.Cores >= tmpl.CPU.MinCores && data.CPU.Speed >= tmpl.CPU.MinSpeed
	results = append(results, checkResult{
		Name:     "CPU Config",
		Expected: fmt.Sprintf("%d Cores @ %.1fGHz", tmpl.CPU.MinCores, tmpl.CPU.MinSpeed),
		Actual:   fmt.Sprintf("%d Cores @ %.1fGHz", data.CPU.Cores, data.CPU.Speed),
		Passed:   cpuPassed,
	})

	// 2. RAM
	actualRAM := bytesToGBDec(data.RAM.Total)
	ramPassed := actualRAM >= float64(tmpl.RAM.MinTotalGB)-1.0 // Допуск 1ГБ на резервацию ядром
	results = append(results, checkResult{
		Name:     "Total RAM",
		Expected: fmt.Sprintf(">= %d GB", tmpl.RAM.MinTotalGB),
		Actual:   fmt.Sprintf("%.1f GB", actualRAM),
		Passed:   ramPassed,
	})

	// 3. Disks (Считаем сколько дисков нужного типа и размера)
	validDisks := 0
	for _, bd := range data.BlockDevs {
		if diskTypeMatches(bd.Type, tmpl.Disks.Type) && bytesToGBDec(bd.Size) >= float64(tmpl.Disks.MinSizeGB) {
			validDisks++
		}
	}
	results = append(results, checkResult{
		Name:     fmt.Sprintf("Storage (%s)", tmpl.Disks.Type),
		Expected: fmt.Sprintf("%d drives >= %d GB", tmpl.Disks.MinCount, tmpl.Disks.MinSizeGB),
		Actual:   fmt.Sprintf("%d drives detected", validDisks),
		Passed:   validDisks >= tmpl.Disks.MinCount,
	})

	// 4. Network (Считаем общее кол-во интерфейсов)
	totalNICs := len(data.Network)
	results = append(results, checkResult{
		Name:     "Network Interfaces",
		Expected: fmt.Sprintf("%d ports", tmpl.Network.MinTotalNICs),
		Actual:   fmt.Sprintf("%d ports detected", totalNICs),
		Passed:   totalNICs >= tmpl.Network.MinTotalNICs,
	})

	return results
}

func diskTypeMatches(actual, expected string) bool {
	a := strings.ToUpper(strings.TrimSpace(actual))
	e := strings.ToUpper(strings.TrimSpace(expected))

	if a == e {
		return true
	}

	// NVMe как подтип SSD
	if e == "SSD" && a == "NVME" {
		return true
	}

	return false
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

	// Основной стиль и заголовок
	html := `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Health Check Report - ` + tmpl.Name + `</title>
	<style>
		body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; margin: 40px; background: #f8f9fa; color: #333; }
		.container { background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
		h1 { color: #2c3e50; border-bottom: 2px solid #eee; padding-bottom: 10px; }
		h2 { color: #2c3e50; margin-top: 40px; background: #e9ecef; padding: 10px; border-radius: 4px; }
		h3 { color: #7f8c8d; border-left: 4px solid #3498db; padding-left: 10px; margin-top: 25px; }
		table { width: 100%; border-collapse: collapse; margin-top: 10px; margin-bottom: 20px; }
		th, td { padding: 12px; border: 1px solid #dee2e6; text-align: left; }
		th { background-color: #f1f1f1; font-weight: 600; }
		.pass { color: #27ae60; font-weight: bold; background: #ebfaef; text-align: center; }
		.fail { color: #c0392b; font-weight: bold; background: #fdf2f2; text-align: center; }
		.val { font-family: monospace; background: #f8f9fa; padding: 2px 4px; border-radius: 3px; }
		.footer { margin-top: 50px; font-size: 0.9em; color: #95a5a6; text-align: center; }
	</style>
</head>
<body>
<div class="container">
	<h1>Hardware Health Check: ` + tmpl.Name + `</h1>
	<p><strong>Generated on:</strong> ` + time.Now().Format(time.RFC1123) + `</p>
	<p><strong>Platform Description:</strong> ` + tmpl.Description + `</p>

	<h2>1. Validation Results</h2>
	<table>
		<tr>
			<th>Check Parameter</th>
			<th>Expected Value</th>
			<th>Actual Value</th>
			<th style="width: 100px;">Status</th>
		</tr>`

	// 1. Таблица результатов проверки
	for _, r := range results {
		statusClass := "pass"
		statusText := "PASS"
		if !r.Passed {
			statusClass = "fail"
			statusText := "FAIL"
			_ = statusText // просто для ясности
		}
		if !r.Passed {
			statusText = "FAIL"
		} else {
			statusText = "OK"
		}

		html += fmt.Sprintf(`
		<tr>
			<td>%s</td>
			<td class="val">%s</td>
			<td class="val">%s</td>
			<td class="%s">%s</td>
		</tr>`, r.Name, r.Expected, r.Actual, statusClass, statusText)
	}

	html += `</table>`

	// 2. Секция подробностей (Inventory)
	html += `<h2>2. Detailed System Inventory</h2>`

	// Общая инфо
	html += `<h3>System Information</h3>
	<table>
		<tr><th>Host S/N</th><td class="val">` + data.CPU.HardwareSN + `</td></tr>
		<tr><th>Motherboard</th><td class="val">` + data.CPU.Motherboard + ` (` + data.CPU.MotherboardSerial + `)</td></tr>
		<tr><th>OS / Kernel</th><td>` + data.Host.OS + ` / ` + data.Host.Kernel + `</td></tr>
	</table>`

	// Процессор
	html += `<h3>CPU Details</h3>
	<table>
		<tr><th>Model</th><td class="val">` + data.CPU.Model + `</td></tr>
		<tr><th>Cores / Threads</th><td>` + fmt.Sprintf("%d / %d", data.CPU.Cores, data.CPU.Threads) + `</td></tr>
	</table>`

	// Сеть
	html += `<h3>Network Interfaces</h3>
	<table>
		<tr>
			<th>Interface</th>
			<th>Model</th>
			<th>MAC Address</th>
			<th>Max Speed</th>
			<th>Current Status</th>
		</tr>`
	for _, n := range data.Network {
		html += fmt.Sprintf(`
		<tr>
			<td class="val"><b>%s</b></td>
			<td>%s</td>
			<td class="val">%s</td>
			<td>%d Mbps</td>
			<td>%s</td>
		</tr>`, n.Interface, n.Model, n.MAC, n.MaxSpeedMbps, n.Status)
	}
	html += `</table>`

	// Диски
	html += `<h3>Storage Devices</h3>
	<table>
		<tr>
			<th>Device</th>
			<th>Model</th>
			<th>Type</th>
			<th>Size</th>
			<th>Health Status</th>
			<th>Wear Level</th>
		</tr>`
	for _, d := range data.BlockDevs {
		wearStr := "N/A"
		if d.Wear >= 0 {
			wearStr = fmt.Sprintf("%.0f%%", d.Wear)
		}
		html += fmt.Sprintf(`
		<tr>
			<td class="val">%s</td>
			<td>%s</td>
			<td>%s</td>
			<td>%.1f GB</td>
			<td>%s</td>
			<td>%s</td>
		</tr>`, d.Device, d.Model, d.Type, bytesToGBDec(d.Size), d.Status, wearStr)
	}
	html += `</table>`

	html += `
	<div class="footer">
		Server Checker Tool v1.0 | Automating Hardware Validation
	</div>
</div>
</body>
</html>`

	return os.WriteFile(filepath.Join(dir, "report.html"), []byte(html), 0644)
}

// Диаг. логи
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
