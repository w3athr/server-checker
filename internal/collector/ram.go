package collector

import (
	"server-checker/internal/models"

	"github.com/shirou/gopsutil/v3/mem"
)

func collectRAM() models.RAMInfo {
	var info models.RAMInfo

	// Общая статистика через gopsutil (работает везде)
	v, _ := mem.VirtualMemory()
	info.Total = v.Total
	info.Used = v.Used
	info.Free = v.Free
	info.UsedPercent = v.UsedPercent

	// Вызываем функцию, которая определена в разных файлах для разных ОС
	info.Sticks = getPhysicalRAM()

	return info
}
