package collector

import (
	"server-checker/internal/models"
	"strings"

	"github.com/shirou/gopsutil/v3/disk"
)

func collectDisks() []models.DiskInfo {
	var finalDisks []models.DiskInfo

	// получаем список разделом
	partitions, _ := disk.Partitions(false)

	// получаем модели и износ
	physData := getPhysicalDriveData()

	for _, p := range partitions {
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}

		d := models.DiskInfo{
			Device:      p.Device,
			Mountpoint:  p.Mountpoint,
			Total:       usage.Total,
			Used:        usage.Used,
			Free:        usage.Free,
			UsedPercent: usage.UsedPercent,
		}

		// ищем какому физическому диску соответствует данный раздел
		// /dev/sda1  ->  sda
		for physPath, info := range physData {
			if strings.HasPrefix(p.Device, physPath) || p.Device == physPath {
				d.Model = info.Model
				d.Type = info.Type
				d.Wear = info.Wear
				d.Status = info.Status
				d.PowerCycles = info.PowerCycles
				d.PowerOnHours = info.PowerOnHours
				d.UnsafeShutdowns = info.UnsafeShutdowns
				break
			}
		}

		// если соответствие не найдено (Windows), подставляем общие данные
		if d.Model == "" && len(physData) > 1 {
			for _, info := range physData {
				d.Model = info.Model
				d.Type = info.Type
				d.Status = info.Status
				d.PowerCycles = info.PowerCycles
				d.PowerOnHours = info.PowerOnHours
				d.UnsafeShutdowns = info.UnsafeShutdowns
			}
		}

		finalDisks = append(finalDisks, d)
	}
	return finalDisks
}
