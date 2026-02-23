// internal/collector/disks.go
package collector

import (
	"fmt"
	"os"
	"path/filepath"
	"server-checker/internal/models"
	"strconv"
	"strings"

	"github.com/anatol/smart.go"
	"github.com/shirou/gopsutil/v3/disk"
)

type physDiskData struct {
	Model           string
	Type            string
	Wear            float64
	Status          string
	PowerCycles     uint64
	PowerOnHours    uint64
	UnsafeShutdowns uint64
}

func collectDisks() []models.DiskInfo {
	partitions, _ := disk.Partitions(false)
	final := make([]models.DiskInfo, 0, len(partitions))

	for _, p := range partitions {
		if !strings.HasPrefix(p.Device, "/dev/") {
			continue
		}

		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}

		final = append(final, models.DiskInfo{
			Device:      p.Device,
			Mountpoint:  p.Mountpoint,
			Total:       usage.Total,
			Used:        usage.Used,
			Free:        usage.Free,
			UsedPercent: usage.UsedPercent,
		})
	}

	return final
}

func getPhysicalDriveData() map[string]physDiskData {
	driveMap := make(map[string]physDiskData)
	devices, _ := os.ReadDir("/sys/block")

	for _, d := range devices {
		name := d.Name()
		if strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "ram") {
			continue
		}

		devPath := "/dev/" + name
		info := physDiskData{
			Model:  "Unknown",
			Type:   "HDD",
			Wear:   -1,
			Status: "Unknown",
		}

		// 1) Модель
		if b, err := os.ReadFile(filepath.Join("/sys/block", name, "device/model")); err == nil {
			info.Model = strings.TrimSpace(string(b))
		}

		// 2) Тип SSD/HDD по rotational
		if b, err := os.ReadFile(filepath.Join("/sys/block", name, "queue/rotational")); err == nil {
			if strings.TrimSpace(string(b)) == "0" {
				info.Type = "SSD"
			}
		}

		// 3) NVMe SMART
		if strings.HasPrefix(name, "nvme") {
			info.Type = "NVMe"
			if sm, err := smart.OpenNVMe(devPath); err == nil {
				if data, err := sm.ReadSMART(); err == nil {
					info.Wear = float64(data.PercentUsed)
					info.PowerCycles = data.PowerCycles.Val[0]
					info.PowerOnHours = data.PowerOnHours.Val[0]
					info.UnsafeShutdowns = data.UnsafeShutdowns.Val[0]

					info.Status = "OK"
					if info.Wear > 80 {
						info.Status = "Critical"
					} else if info.Wear > 50 {
						info.Status = "Warning"
					}
				}
				sm.Close()
			}
		} else if strings.HasPrefix(name, "sd") {
			// Логика для SATA (SSD/HDD)
			if sm, err := smart.OpenSata(devPath); err == nil {
				if data, err := sm.ReadSMARTData(); err == nil {
					info.Status = "OK"
					// Ищем атрибуты износа (Wear Level)
					// 231 (SSD Life Left) или 177 (Wear Range Delta)
					for _, attr := range data.Attrs {
						if attr.Id == 231 || attr.Id == 177 || attr.Id == 233 {
							// В SATA SMART 'Value' обычно показывает остаток здоровья (100 -> 0)
							// Мы переводим это в процент износа (0 -> 100)
							info.Wear = 100 - float64(attr.ValueRaw)
						}
						// Power On Hours для SATA
						if attr.Id == 9 {
							info.PowerOnHours = attr.ValueRaw
						}
						// Power Cycles для SATA
						if attr.Id == 12 {
							info.PowerCycles = attr.ValueRaw
						}
					}
					info.Status = calculateStatus(info.Wear)
				}
				sm.Close()
			}
		}

		driveMap[devPath] = info
	}

	return driveMap
}

// Вспомогательная функция для статуса
func calculateStatus(wear float64) string {
	if wear < 0 {
		return "Unknown"
	}
	if wear > 80 {
		return "Critical"
	}
	if wear > 50 {
		return "Warning"
	}
	return "OK"
}

func collectBlockDevices(phys map[string]physDiskData) []models.BlockDeviceInfo {
	devs, _ := os.ReadDir("/sys/block")
	out := make([]models.BlockDeviceInfo, 0, len(devs))

	for _, d := range devs {
		name := d.Name()

		if strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "ram") || strings.HasPrefix(name, "dm-") {
			continue
		}
		if isRemovable(name) {
			continue
		}
		if !hasRealDevice(name) {
			continue
		}

		devPath := "/dev/" + name
		sizeBytes, _ := readBlockDevSizeBytes(name)

		p := phys[devPath]

		model := p.Model
		if model == "" {
			model = "Unknown"
		}
		typ := p.Type
		if typ == "" {
			typ = "Unknown"
		}
		status := p.Status
		if status == "" {
			status = "Unknown"
		}

		out = append(out, models.BlockDeviceInfo{
			Device: devPath,
			Model:  model,
			Type:   typ,
			Size:   sizeBytes,

			Wear:            p.Wear,
			Status:          status,
			PowerCycles:     p.PowerCycles,
			PowerOnHours:    p.PowerOnHours,
			UnsafeShutdowns: p.UnsafeShutdowns,
		})
	}

	return out
}

func readBlockDevSizeBytes(sysName string) (uint64, error) {
	b, err := os.ReadFile(filepath.Join("/sys/block", sysName, "size"))
	if err != nil {
		return 0, err
	}

	secStr := strings.TrimSpace(string(b))
	sectors, err := strconv.ParseUint(secStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse /sys/block/%s/size: %w", sysName, err)
	}

	return sectors * 512, nil
}

func isRemovable(sysName string) bool {
	b, err := os.ReadFile(filepath.Join("/sys/block", sysName, "removable"))
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(b)) == "1"
}

func hasRealDevice(sysName string) bool {
	_, err := os.Stat(filepath.Join("/sys/block", sysName, "device"))
	return err == nil
}
