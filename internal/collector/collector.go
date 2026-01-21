package collector

import (
	"server-checker/internal/models"

	"github.com/shirou/gopsutil/v3/host"
)

func CollectAll() (models.SystemInfo, error) {
	var info models.SystemInfo

	info.CPU = collectCPU()
	info.RAM = collectRAM()

	h, _ := host.Info()
	info.Host = models.HostInfo{
		OS:     h.OS,
		Kernel: h.KernelVersion,
		Uptime: h.Uptime,
	}

	return info, nil
}
