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
	"text/template"
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

	sn := currentData.CPU.HardwareSN
	if sn == "" || sn == "N/A" {
		sn = "UnknownSN"
	}
	safeTmplName := strings.ReplaceAll(tmpl.Name, " ", "_")
	dirName := fmt.Sprintf("%s_%s_%s", safeTmplName, time.Now().Format("2006-01-02_15-04-05"), sn)

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
	cpuOk := data.CPU.Cores >= tmpl.CPU.MinCores && data.CPU.Speed >= tmpl.CPU.MinSpeed
	results = append(results, checkResult{
		Name:     "CPU Configuration",
		Expected: fmt.Sprintf(">= %d Cores @ %.1fGHz", tmpl.CPU.MinCores, tmpl.CPU.MinSpeed),
		Actual:   fmt.Sprintf("%d Cores @ %.1fGHz", data.CPU.Cores, data.CPU.Speed),
		Passed:   cpuOk,
	})

	// 2. RAM
	var physicalRAMBytes uint64
	if len(data.RAM.Sticks) > 0 {
		// Если смогли прочитать физические плашки (запущено через sudo), суммируем их физический объем
		for _, s := range data.RAM.Sticks {
			physicalRAMBytes += s.Capacity
		}
	} else {
		// Фолбек: если плашки не прочитались, берем то, что отдала ОС
		physicalRAMBytes = data.RAM.Total
	}
	
	// Используем нашу обновленную функцию bytesToGBDec (которая делит на 1024)
	actualRAM := bytesToGBDec(physicalRAMBytes)
	
	// Проверяем, хватает ли памяти (с допуском в 1 ГБ на всякий случай)
	ramOk := actualRAM >= float64(tmpl.RAM.MinTotalGB)-1.0 
	
	results = append(results, checkResult{
		Name:     "Total RAM Capacity",
		Expected: fmt.Sprintf(">= %d GB", tmpl.RAM.MinTotalGB),
		Actual:   fmt.Sprintf("%.1f GB", actualRAM), // Теперь здесь будет 16.0 GB
		Passed:   ramOk,
	})

	// 3. Диски (считаем количество подходящих)
	validDisks := 0
	for _, bd := range data.BlockDevs {
		// Для дисков используем маркетинговый объем (base-1000)
		marketingGB := float64(bd.Size) / 1000000000.0
		
		if diskTypeMatches(bd.Type, tmpl.Disks.Type) && marketingGB >= float64(tmpl.Disks.MinSizeGB)-5.0 {
			validDisks++
		}
	}
	results = append(results, checkResult{
		Name:     fmt.Sprintf("Storage Devices (%s)", tmpl.Disks.Type),
		Expected: fmt.Sprintf("%d drives x %d GB", tmpl.Disks.MinCount, tmpl.Disks.MinSizeGB),
		Actual:   fmt.Sprintf("%d drives detected", validDisks),
		Passed:   validDisks >= tmpl.Disks.MinCount,
	})

	// 4. Сеть (общее кол-во)
	totalNICs := len(data.Network)
	results = append(results, checkResult{
		Name:     "Total Network Ports",
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
	const htmlLayout = `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { 
			font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; 
			margin: 40px; 
			background: #f0f2f5; 
			color: #000; 
		}
		
		/* Лист отчета с закругленными углами */
		.card { 
			background: white; 
			padding: 40px; 
			max-width: 1000px; 
			margin: auto; 
			border: 1px solid #e1e4e8;
			border-radius: 12px;
			box-shadow: 0 4px 12px rgba(0,0,0,0.05);
		}

		h1 { color: #2c3e50; border-bottom: 2px solid #eee; padding-bottom: 10px; margin-top: 0; }
		
		/* Главные подразделы: Темный фон и закругления */
		h2 { 
			background: #34495e; 
			color: white; 
			padding: 10px 15px; 
			border-radius: 8px; 
			margin-top: 35px; 
			font-size: 1.1em; 
		}

		/* Второстепенные заголовки: Просто жирный черный текст */
		h3 { 
			color: #000; 
			margin-top: 25px; 
			margin-bottom: 10px; 
			font-weight: bold;
			font-size: 1.1em;
		}

		table { width: 100%; border-collapse: collapse; margin: 10px 0; background: white; }
		th, td { padding: 12px; border: 1px solid #eee; text-align: left; }
		th { background-color: #f8f9fa; font-weight: bold; color: #555; }

		/* Стилизация статусов (плашки как на скриншоте слева) */
		.status-cell { text-align: center; width: 80px; }
		
		.pass { 
			background-color: #d4edda; 
			color: #155724; 
			font-weight: bold; 
			padding: 6px 12px; 
			border-radius: 6px;
			display: inline-block;
			min-width: 50px;
		}
		
		.fail { 
			background-color: #f8d7da; 
			color: #721c24; 
			font-weight: bold; 
			padding: 6px 12px; 
			border-radius: 6px;
			display: inline-block;
			min-width: 50px;
		}

		.val { font-family: inherit; }

		.footer { text-align: center; margin-top: 50px; color: #999; font-size: 0.8em; border-top: 1px solid #eee; padding-top: 20px; }
		.spec-label { font-weight: bold; color: #666; width: 220px; background: #fafafa; }
	</style>
</head>
<body>
	<div class="card">
		<h1>Hardware Health Report: {{.Template.Name}}</h1>
		<p><strong>Validation Template:</strong> {{.Template.Description}}</p>
		<p><strong>Report Date:</strong> {{.Timestamp}}</p>

		<h2>1. Validation Summary</h2>
		<table>
			<thead>
				<tr><th>Check Parameter</th><th>Expected</th><th>Actual</th><th class="status-cell">Result</th></tr>
			</thead>
			<tbody>
				{{range .Results}}
				<tr>
					<td>{{.Name}}</td>
					<td class="val">{{.Expected}}</td>
					<td class="val">{{.Actual}}</td>
					<td class="status-cell">
						<span class="{{if .Passed}}pass{{else}}fail{{end}}">
							{{if .Passed}}OK{{else}}FAIL{{end}}
						</span>
					</td>
				</tr>
				{{end}}
			</tbody>
		</table>

		<h2>2. System Inventory</h2>
		
		<h3>Platform Information</h3>
		<table>
			<tr><td class="spec-label">Hardware Vendor</td><td>{{if .Data.CPU.HardwareVendor}}{{.Data.CPU.HardwareVendor}}{{else}}Default string{{end}}</td></tr>
			<tr><td class="spec-label">Product Model</td><td>{{if .Data.CPU.HardwareModel}}{{.Data.CPU.HardwareModel}}{{else}}Default string{{end}}</td></tr>
			<tr>
				<td class="spec-label">System Serial Number</td>
				<td class="code">
					<div style="display: flex; justify-content: space-between; align-items: center;">
						<span>{{.Data.CPU.HardwareSN}}</span>
						<span style="color: #999; font-style: italic; font-size: 0.85em; font-weight: normal;">For CRM add '00' at the beginning</span>
					</div>
				</td>
			</tr>
			<tr><td class="spec-label">Motherboard</td><td>{{.Data.CPU.Motherboard}} (S/N: {{.Data.CPU.MotherboardSerial}})</td></tr>
			<tr><td class="spec-label">Operating System</td><td>{{.Data.Host.OS}}</td></tr>
			<tr><td class="spec-label">Kernel Version</td><td>{{.Data.Host.Kernel}}</td></tr>
		</table>

		<h3>Processor (CPU)</h3>
		<table>
			<tr><td class="spec-label">Model Name</td><td>{{.Data.CPU.Model}}</td></tr>
			<tr><td class="spec-label">Architecture</td><td>{{.Data.CPU.Cores}} Cores / {{.Data.CPU.Threads}} Threads</td></tr>
			<tr><td class="spec-label">Current Frequency</td><td>{{printf "%.2f" .Data.CPU.Speed}} GHz</td></tr>
		</table>

		<h3>Memory (RAM)</h3>
		<p style="margin-left: 5px;"><strong>Total Capacity:</strong> {{physicalRAMGB .Data}} GB</p>
		<table>
			<thead>
				<tr><th>Slot / Locator</th><th>Module Model</th><th>Capacity</th><th>Speed</th></tr>
			</thead>
			<tbody>
				{{range .Data.RAM.Sticks}}
				<tr>
					<td>{{.Slot}}</td>
					<td>{{.Model}}</td>
					<td>{{bytesToGB .Capacity}} GB</td>
					<td>{{.Speed}} MT/s</td>
				</tr>
				{{else}}
				<tr><td colspan="4" style="text-align:center; color:#999;">Physical RAM details not available</td></tr>
				{{end}}
			</tbody>
		</table>

		<h3>Network Adapters</h3>
		<table>
			<thead>
				<tr><th>Interface</th><th>Hardware Model</th><th>MAC Address</th><th>Max Speed</th></tr>
			</thead>
			<tbody>
				{{range .Data.Network}}
				<tr>
					<td><strong>{{.Interface}}</strong></td>
					<td>{{.Model}}</td>
					<td class="code">{{.MAC}}</td>
					<td>{{.MaxSpeedMbps}} Mbps</td>
				</tr>
				{{end}}
			</tbody>
		</table>

		<h3>Storage Devices</h3>
		<table>
			<thead>
				<tr><th>Device</th><th>Model / Firmware</th><th>Type</th><th>Size</th><th>Wear Level</th><th>Hours</th><th>Status</th></tr>
			</thead>
			<tbody>
				{{range .Data.BlockDevs}}
				<tr>
					<td class="code">{{.Device}}</td>
					<td>{{.Model}}</td>
					<td>{{.Type}}</td>
					<td>{{marketingGB .Size}} GB</td>
					<td>{{if ge .Wear 0.0}}{{.Wear}}%{{else}}N/A{{end}}</td>
					<td>{{.PowerOnHours}}</td>
					<td style="font-weight:bold;">{{.Status}}</td>
				</tr>
				{{end}}
			</tbody>
		</table>

		<div class="footer">
			Server Checker Tool v1.2 | Hardware Validation System
		</div>
	</div>
</body>
</html>`

	// Остальная часть функции без изменений
	reportData := struct {
		Template  models.SystemTemplate
		Data      models.SystemInfo
		Results   []checkResult
		Timestamp string
	}{
		Template:  tmpl,
		Data:      data,
		Results:   runHealthCheck(tmpl, data),
		Timestamp: time.Now().Format(time.RFC1123),
	}

	t, err := template.New("report").Funcs(template.FuncMap{
		"bytesToGB": func(b uint64) string {
			return fmt.Sprintf("%.1f", float64(b)/1024/1024/1024)
		},
		"marketingGB": func(b uint64) string {
			// Вывод маркетингового размера диска без дробной части (например, 512)
			return fmt.Sprintf("%.0f", float64(b)/1000/1000/1000)
		},
		"physicalRAMGB": func(data models.SystemInfo) string {
			// Суммарный размер физической оперативной памяти
			if len(data.RAM.Sticks) > 0 {
				var total uint64
				for _, s := range data.RAM.Sticks {
					total += s.Capacity
				}
				return fmt.Sprintf("%.1f", float64(total)/1024/1024/1024)
			}
			return fmt.Sprintf("%.1f", float64(data.RAM.Total)/1024/1024/1024)
		},
	}).Parse(htmlLayout)

	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, reportData); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return os.WriteFile(filepath.Join(dir, "report.html"), buf.Bytes(), 0644)
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

	devices, _ := os.ReadDir("/sys/block")
	for _, d := range devices {
		name := d.Name()
		// Пропускаем виртуальные устройства, петли и ram-диски
		if strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "ram") || strings.HasPrefix(name, "dm-") {
			continue
		}

		devPath := "/dev/" + name
		// Выгружаем smartctl -a для каждого диска
		merr.add(writeCommandOutput(ctx, logsDir, "smart_"+name+".txt", "smartctl", "-a", devPath))
	}

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

	// Добавляем код выхода для ясности
	if err != nil {
		sb.WriteString(fmt.Sprintf("Exit Code/Error: %v\n", err))
	} else {
		sb.WriteString("Exit Code: 0 (Success)\n")
	}
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

	// Записываем файл на диск. Если запись удалась, возвращаем nil,
	// даже если сама команда (например, smartctl) вернула ошибку.
	if werr := os.WriteFile(fullPath, []byte(sb.String()), 0644); werr != nil {
		return fmt.Errorf("failed to write %s: %w", fullPath, werr)
	}

	// Мы НЕ возвращаем err команды, так как лог уже сохранен и это не критично для генерации отчета.
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
