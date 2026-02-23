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

				d.PowerCycles = info.PowerCycles
				d.PowerOnHours = info.PowerOnHours
				d.UnsafeShutdowns = info.UnsafeShutdowns
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
					info.PowerCycles = data.PowerCycles.Val[0]
					info.PowerOnHours = data.PowerOnHours.Val[0]
					info.UnsafeShutdowns = data.UnsafeShutdowns.Val[0]
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

func collectBlockDevices(phys map[string]models.DiskInfo) []models.BlockDeviceInfo {
	devs, _ := os.ReadDir("/sys/block")
	out := make([]models.BlockDeviceInfo, 0, len(devs))

	for _, d := range devs {
		name := d.Name()
		if strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "ram") {
			continue
		}

		devPath := "/dev/" + name

		// размер берем из /sys/block/<dev>/size (кол-во 512B секторов)
		sizeBytes, _ := readBlockDevSizeBytes(name)

		// модель/тип берем из phys мапы, которую ты уже собираешь
		p := phys[devPath]
		if p.Model == "" {
			p.Model = "Unknown"
		}
		if p.Type == "" {
			p.Type = "Unknown"
		}

		out = append(out, models.BlockDeviceInfo{
			Device: devPath,
			Model:  p.Model,
			Type:   p.Type,
			Size:   sizeBytes,
		})
	}

	return out
}

func readBlockDevSizeBytes(sysName string) (uint64, error) {
	// /sys/block/sda/size -> количество 512-byte секторов
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
