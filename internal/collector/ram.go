package collector

import (
	"server-checker/internal/models"
	"strconv"
	"strings"

	"github.com/dselans/dmidecode"
	"github.com/shirou/gopsutil/v3/mem"
)

// updateRAMDynamic собирает текущую загрузку RAM
func updateRAMDynamic(info models.RAMInfo) models.RAMInfo {
	v, _ := mem.VirtualMemory()
	info.Total = v.Total
	info.Used = v.Used
	info.Available = v.Available
	info.UsedPercent = v.UsedPercent
	return info
}

// getPhysicalRAM собирает данные о физических планках (требует sudo)
func getPhysicalRAM() []models.RAMStick {
	var sticks []models.RAMStick
	dmi := dmidecode.New()

	if err := dmi.Run(); err != nil {
		return sticks // Если не sudo, вернет пустой список
	}

	// Type 17 — это Memory Device в стандарте SMBIOS
	devices, _ := dmi.SearchByType(17)

	for _, dev := range devices {
		sizeStr := dev["Size"]
		// Пропускаем пустые слоты
		if sizeStr == "No Module Installed" || sizeStr == "" || strings.Contains(sizeStr, "Unknown") {
			continue
		}

		sticks = append(sticks, models.RAMStick{
			Slot:     dev["Locator"],
			Model:    strings.TrimSpace(dev["Manufacturer"] + " " + dev["Part Number"]),
			Capacity: parseDMISize(sizeStr),
			Speed:    parseDMISpeed(dev["Speed"]),
		})
	}
	return sticks
}

func parseDMISize(sizeStr string) uint64 {
	parts := strings.Fields(sizeStr)
	if len(parts) < 2 {
		return 0
	}
	val, _ := strconv.ParseFloat(parts[0], 64)
	unit := strings.ToUpper(parts[1])

	switch unit {
	case "GB":
		return uint64(val * 1024 * 1024 * 1024)
	case "MB":
		return uint64(val * 1024 * 1024)
	default:
		return uint64(val)
	}
}

func parseDMISpeed(speedStr string) uint32 {
	parts := strings.Fields(speedStr)
	if len(parts) < 1 {
		return 0
	}
	val, _ := strconv.Atoi(parts[0])
	return uint32(val)
}
