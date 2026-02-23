package templates

import "server-checker/internal/models"

func GetEmbeddedTemplate(name string) (models.SystemTemplate, bool) {
	switch name {
	case "Рубеж-Н1":
		return models.SystemTemplate{
			Name:        "Рубеж-Н1",
			Description: "Atom x6425E (4C/4T), 16GB RAM, 8x 1G NICs",
			CPU: models.CPUTemplate{
				MinCores: 4, MinThreads: 4, MinSpeed: 1.9,
			},
			RAM: models.RAMTemplate{MinTotalGB: 16},
			Disks: models.DiskTemplate{
				MinCount: 1, MinSizeGB: 230, Type: "SSD",
			},
			Network: models.NetworkTemplate{
				MinTotalNICs: 8, MinSpeedMbps: 1000,
			},
		}, true

	case "Рубеж-С":
		return models.SystemTemplate{
			Name:        "Рубеж-С",
			Description: "Xeon D-1537 (8C/16T), 32GB RAM, 16x 1G NICs",
			CPU: models.CPUTemplate{
				MinCores: 8, MinThreads: 16, MinSpeed: 1.7,
			},
			RAM: models.RAMTemplate{MinTotalGB: 32},
			Disks: models.DiskTemplate{
				MinCount: 1, MinSizeGB: 470, Type: "SSD",
			},
			Network: models.NetworkTemplate{
				MinTotalNICs: 16, MinSpeedMbps: 1000,
			},
		}, true

	case "Рубеж-Т":
		return models.SystemTemplate{
			Name:        "Рубеж-Т",
			Description: "Xeon Silver 4310 (12C/24T x2), 64GB RAM, 22x NICs (2x 10G + 20x 1G)",
			CPU: models.CPUTemplate{
				MinCores: 24, MinThreads: 48, MinSpeed: 2.1,
			},
			RAM: models.RAMTemplate{MinTotalGB: 64},
			Disks: models.DiskTemplate{
				MinCount: 2, MinSizeGB: 470, Type: "SSD",
			},
			Network: models.NetworkTemplate{
				MinTotalNICs: 22, MinSpeedMbps: 1000,
			},
		}, true

	default:
		return models.SystemTemplate{}, false
	}
}
