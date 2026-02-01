//go:build windows

package collector

import (
	"server-checker/internal/models"

	"github.com/yusufpapurcu/wmi"
)

type MSFT_PhysicalDisk struct {
	FriendlyName string
	MediaType    uint16
	BusType      uint16
	HealthStatus uint16 // 0: Healthy, 1: Warning, 2: Unhealthy
}

func getPhysicalDriveData() map[string]models.DiskInfo {
	results := make(map[string]models.DiskInfo)
	var dst []MSFT_PhysicalDisk

	query := "SELECT FriendlyName, MediaType, BusType, HealthStatus FROM MSFT_PhysicalDisk"
	err := wmi.QueryNamespace(query, &dst, "ROOT\\Microsoft\\Windows\\Storage")
	if err == nil {
		for _, d := range dst {
			info := models.DiskInfo{
				Model: d.FriendlyName,
				Type:  "HDD",
				Wear:  -1, // не удалось определить
			}
			if d.BusType == 17 {
				info.Type = "NMVe"
			} else if d.MediaType == 4 {
				info.Type = "SSD"
			}

			switch d.HealthStatus {
			case 0:
				info.Status = "OK"
			case 1:
				info.Status = "Warning"
			case 2:
				info.Status = "Critical"
			}
			results[d.FriendlyName] = info
		}
	}
	return results
}
