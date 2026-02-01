package collector

import (
	"server-checker/internal/models"

	"github.com/shirou/gopsutil/v3/mem"
)

func updateRAMDynamic(info models.RAMInfo) models.RAMInfo {
	v, _ := mem.VirtualMemory()

	info.Total = v.Total
	info.Used = v.Used
	info.Free = v.Free
	info.UsedPercent = v.UsedPercent

	return info
}
