//go:build linux

package collector

import (
	"server-checker/internal/models"
	"strconv"
	"strings"

	"github.com/dselans/dmidecode"
)

func getPhysicalRAM() []models.RAMStick {
	var sticks []models.RAMStick
	dmi := dmidecode.New()

	if err := dmi.Run(); err != nil {
		return sticks
	}

	devices, _ := dmi.SearchByType(17) // Type 17 - Memory Device

	for _, dev := range devices {
		if dev["Size"] == "No Module Installed" || dev["Size"] == "" {
			continue
		}

		sticks = append(sticks, models.RAMStick{
			Slot:     dev["Locator"],
			Model:    dev["Manufacturer"] + " " + dev["Part Number"],
			Capacity: parseDMISize(dev["Size"]),
			Speed:    parseDMISpeed(dev["Speed"]),
		})
	}
	return sticks
}

// Вспомогательная функция для парсинга "8 GB" или "8192 MB" в байты
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
	case "KB":
		return uint64(val * 1024)
	default:
		return uint64(val)
	}
}

// Вспомогательная функция для парсинга "3200 MT/s" или "3200 MHz" в число
func parseDMISpeed(speedStr string) uint32 {
	parts := strings.Fields(speedStr)
	if len(parts) < 1 {
		return 0
	}
	// Убираем всё кроме цифр (на случай "3200 MT/s")
	val, _ := strconv.Atoi(parts[0])
	return uint32(val)
}
