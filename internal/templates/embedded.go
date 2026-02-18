package templates

import (
	"server-checker/internal/models"
)

// Возврат встроенного шаблона по имени
func GetEmbeddedTemplate(name string) (models.SystemTemplate, bool) {
	switch name {
	case "Рубеж-Н1":
		return models.SystemTemplate{
			Name:        "Рубеж-Н1",
			Description: "Standard configuration for Н1 servers",
			CPU: models.CPUTemplate{
				MinCores:   4,
				MinThreads: 8,
				MinSpeed:   2.0,
			},
			RAM: models.RAMTemplate{
				MinTotalGB: 16,
			},
			Disks: []models.DiskTemplate{
				{Name: "System", MinSizeGB: 240, Type: "SSD"},
			},
			Network: []models.NetworkTemplate{
				{MinSpeed: 1000},
			},
		}, true
	case "Рубеж-С":
		return models.SystemTemplate{
			Name:        "Рубеж-С",
			Description: "Standard configuration for С servers",
			CPU: models.CPUTemplate{
				MinCores:   8,
				MinThreads: 16,
				MinSpeed:   2.5,
			},
			RAM: models.RAMTemplate{
				MinTotalGB: 32,
			},
			Disks: []models.DiskTemplate{
				{Name: "System", MinSizeGB: 480, Type: "SSD"},
				{Name: "Data", MinSizeGB: 1000, Type: "HDD"},
			},
			Network: []models.NetworkTemplate{
				{MinSpeed: 1000},
			},
		}, true
	case "Рубеж-Т":
		return models.SystemTemplate{
			Name:        "Рубеж-Т",
			Description: "High-performance configuration for Т servers",
			CPU: models.CPUTemplate{
				MinCores:   16,
				MinThreads: 32,
				MinSpeed:   3.0,
			},
			RAM: models.RAMTemplate{
				MinTotalGB: 64,
			},
			Disks: []models.DiskTemplate{
				{Name: "System", MinSizeGB: 500, Type: "NVMe"},
				{Name: "Data", MinSizeGB: 2000, Type: "SSD"},
			},
			Network: []models.NetworkTemplate{
				{MinSpeed: 10000},
			},
		}, true
	default:
		return models.SystemTemplate{}, false
	}
}
