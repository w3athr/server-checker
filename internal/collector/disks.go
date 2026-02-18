package collector

import (
	"os"
	"path/filepath"
	"server-checker/internal/models"
	"strings"

	"github.com/anatol/smart.go"
	"github.com/shirou/gopsutil/v3/disk"
)

func collectDisks() []models.DiskInfo {
	var finalDisks []models.DiskInfo
	partitions, _ := disk.Partitions(false)
	physData := getPhysicalDriveData()

	for _, p := range partitions {
		// Пропускаем системные и виртуальные ФС
		if !strings.HasPrefix(p.Device, "/dev/") {
			continue
		}

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

		// Сопоставляем раздел с физическим диском
		for devPath, info := range physData {
			if strings.HasPrefix(p.Device, devPath) {
				d.Model = info.Model
				d.Type = info.Type
				d.Wear = info.Wear
				d.Status = info.Status
				break
			}
		}
		finalDisks = append(finalDisks, d)
	}
	return finalDisks
}

func getPhysicalDriveData() map[string]models.DiskInfo {
	driveMap := make(map[string]models.DiskInfo)
	devices, _ := os.ReadDir("/sys/block")

	for _, d := range devices {
		name := d.Name()
		if strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "ram") {
			continue
		}

		devPath := "/dev/" + name
		info := models.DiskInfo{Model: "Unknown", Type: "HDD", Wear: -1, Status: "Unknown"}

		// 1. Модель из sysfs
		if b, err := os.ReadFile(filepath.Join("/sys/block", name, "device/model")); err == nil {
			info.Model = strings.TrimSpace(string(b))
		}

		// 2. Тип (SSD/HDD)
		if b, err := os.ReadFile(filepath.Join("/sys/block", name, "queue/rotational")); err == nil {
			if strings.TrimSpace(string(b)) == "0" {
				info.Type = "SSD"
			}
		}

		// 3. NVMe Wear (Специфично для NVMe)
		if strings.HasPrefix(name, "nvme") {
			info.Type = "NVMe"
			if sm, err := smart.OpenNVMe(devPath); err == nil {
				if data, err := sm.ReadSMART(); err == nil {
					info.Wear = float64(data.PercentUsed)
					info.Status = "OK"
					if info.Wear > 80 {
						info.Status = "Critical"
					}
				}
				sm.Close()
			}
		}
		driveMap[devPath] = info
	}
	return driveMap
}
