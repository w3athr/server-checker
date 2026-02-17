package collector

import (
	"os"
	"path/filepath"
	"server-checker/internal/models"
	"strings"

	"github.com/anatol/smart.go"
)

// сбор карты моделей физических дисков
func getPhysicalDriveData() map[string]models.DiskInfo {
	driveMap := make(map[string]models.DiskInfo)

	// читаем блочные устройства из /sys/block
	devices, _ := os.ReadDir("/sys/block")
	for _, d := range devices {
		name := d.Name()
		// пропускаем loop и ram устройства
		if strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "ram") {
			continue
		}

		devPath := "/dev/" + name
		info := models.DiskInfo{
			Model:  "Unknown",
			Type:   "HDD",
			Wear:   -1, // не удалось определить
			Status: "Unknown",
		}
		// читаем модель /sys
		if b, err := os.ReadFile(filepath.Join("/sys/block", name, "device/model")); err == nil {
			info.Model = strings.TrimSpace(string(b))
		}

		// определяем тип диска и износ
		if strings.HasPrefix(name, "nvme") {
			info.Type = "NVMe"
			// чтение SMART напрямую через ioctl (нужно sudo)
			if sm, err := smart.OpenNVMe(devPath); err == nil {
				if data, err := sm.ReadSMART(); err == nil {
					info.Wear = float64(data.PercentUsed)
					info.Status = "OK"
					if info.Wear > 50 {
						info.Status = "Warning"
					} else if info.Wear > 80 {
						info.Status = "Critical"
					}
				}
				sm.Close()
			}
		} else {
			// проверяем, вращается ли диск (HDD) или нет (SSD)
			if b, err := os.ReadFile(filepath.Join("/sys/block", name, "queue/rotational")); err == nil {
				if strings.TrimSpace(string(b)) == "0" {
					info.Type = "SSD"
				}
			}
		}
		driveMap[devPath] = info
	}
	return driveMap
}
