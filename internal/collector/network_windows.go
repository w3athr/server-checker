//go:build windows

package collector

import (
	"server-checker/internal/models"

	"github.com/yusufpapurcu/wmi"
)

type Win32_NetworkAdapter struct {
	Name                string
	NetConnectionStatus uint16
	Speed               uint64
	MACAddress          string
	InterfaceIndex      uint32
}

func collectNetwork() []models.NetworkInfo {
	var results []models.NetworkInfo
	var dst []Win32_NetworkAdapter

	query := "SELECT Name, NetConnectionStatus, Speed, MACAddress FROM Win32_NetworkAdapter WHERE PhysicalAdapter=True"
	err := wmi.Query(query, &dst)

	if err == nil {
		for _, v := range dst {
			status := "Down"
			if v.NetConnectionStatus == 2 {
				status = "Up"
			}

			results = append(results, models.NetworkInfo{
				Interface: v.Name,
				Model:     v.Name,
				Status:    status,
				MAC:       v.MACAddress,
				Speed:     int(v.Speed / 1024 / 1024),
			})
		}
	}
	return results
}

func BlinkInterface(name string) {
	// на windows пока оставляем пустым, так как стандартного способа
	// "мигнуть портом" через системные вызовы нет
}
