package collector

import (
	"os"
	"server-checker/internal/models"
	"strings"

	"github.com/dselans/dmidecode"
	"github.com/shirou/gopsutil/v3/cpu"
)

func collectCPUStatic() models.CPUInfo {
	var info models.CPUInfo

	cpuStats, _ := cpu.Info()
	if len(cpuStats) > 0 {
		info.Model = cpuStats[0].ModelName
		info.Speed = cpuStats[0].Mhz / 1000
	}

	phys, _ := cpu.Counts(false)
	logic, _ := cpu.Counts(true)
	info.Cores = phys
	info.Threads = logic

	// ===== Motherboard через dmidecode =====
	dmi := dmidecode.New()
	if err := dmi.Run(); err == nil {
		motherboard, _ := dmi.SearchByType(2)
		if len(motherboard) > 0 {
			manufacturer := strings.TrimSpace(motherboard[0]["Manufacturer"])
			product := strings.TrimSpace(motherboard[0]["Product Name"])
			serial := strings.TrimSpace(motherboard[0]["Serial Number"])

			var parts []string
			if manufacturer != "" {
				parts = append(parts, manufacturer)
			}
			if product != "" {
				parts = append(parts, product)
			}

			info.Motherboard = strings.Join(parts, " ")
			info.MotherboardSerial = serial
		}
	}

	// ===== Hardware vendor/model/SN через sysfs =====
	info.HardwareVendor = readSysValue("/sys/class/dmi/id/sys_vendor")
	info.HardwareModel = readSysValue("/sys/class/dmi/id/product_name")
	info.HardwareSN = readSysValue("/sys/class/dmi/id/product_serial")

	return info
}

func collectCPULoad() float64 {
	load, _ := cpu.Percent(0, false)
	if len(load) > 0 {
		return load[0]
	}
	return 0
}

func readSysValue(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}
