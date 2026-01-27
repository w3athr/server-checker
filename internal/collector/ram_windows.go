//go:build windows

package collector

import (
	"server-checker/internal/models"
	"strings"

	"github.com/yusufpapurcu/wmi"
)

type Win32_PhysicalMemory struct {
	Manufacturer  string
	PartNumber    string
	Capacity      uint64
	Speed         uint32
	DeviceLocator string
}

func getPhysicalRAM() []models.RAMStick {
	var sticks []models.RAMStick
	var dst []Win32_PhysicalMemory

	query := "SELECT Manufacturer, PartNumber, Capacity, Speed, DeviceLocator FROM Win32_PhysicalMemory"
	err := wmi.Query(query, &dst)
	if err != nil {
		return sticks
	}

	for _, item := range dst {
		mfg := strings.TrimSpace(item.Manufacturer)
		part := strings.TrimSpace(item.PartNumber)
		sticks = append(sticks, models.RAMStick{
			Slot:     item.DeviceLocator,
			Model:    mfg + " " + part,
			Capacity: item.Capacity,
			Speed:    item.Speed,
		})
	}
	return sticks
}
