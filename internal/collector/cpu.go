package collector

import (
	"server-checker/internal/models"

	"github.com/shirou/gopsutil/v3/cpu"
)

func collectCPUStatic() models.CPUInfo {
	var info models.CPUInfo
	cpuStats, _ := cpu.Info()
	if len(cpuStats) > 0 {
		info.Model = cpuStats[0].ModelName
		info.Speed = cpuStats[0].Mhz / 1000 // преобразование в GHz
	}

	phys, _ := cpu.Counts(false)
	logic, _ := cpu.Counts(true)
	info.Cores = phys
	info.Threads = logic

	return info
}

func collectCPULoad() float64 {
	load, _ := cpu.Percent(0, false)
	if len(load) > 0 {
		return load[0]
	}
	return 0
}
