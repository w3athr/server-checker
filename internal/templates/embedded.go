package templates

import "server-checker/internal/models"

func GetEmbeddedTemplate(name string) (models.SystemTemplate, bool) {
	switch name {
	case "Рубеж-Н1":
		return models.SystemTemplate{
			Name:        "Рубеж-Н1",
			Description: "Atom x6425E, 16GB RAM, 1x SSD, 8x 1G NICs",
			CPU:         models.CPUTemplate{MinCores: 4, MinThreads: 4, MinSpeed: 1.9},
			RAM:         models.RAMTemplate{MinTotalGB: 16},
			Disks:       models.DiskTemplate{MinCount: 1, MinSizeGB: 230, Type: "SSD"},
			Network:     models.NetworkTemplate{MinTotalNICs: 8},
		}, true

	case "Рубеж-С":
		return models.SystemTemplate{
			Name:        "Рубеж-С",
			Description: "Xeon D-1537, 32GB RAM, 1x SSD, 16x 1G NICs",
			CPU:         models.CPUTemplate{MinCores: 8, MinThreads: 16, MinSpeed: 1.7},
			RAM:         models.RAMTemplate{MinTotalGB: 32},
			Disks:       models.DiskTemplate{MinCount: 1, MinSizeGB: 470, Type: "SSD"},
			Network:     models.NetworkTemplate{MinTotalNICs: 16},
		}, true

	case "Рубеж-Т":
		return models.SystemTemplate{
			Name:        "Рубеж-Т",
			Description: "2x Xeon Silver 4310, 64GB RAM, 2x SSD, 22x NICs",
			CPU:         models.CPUTemplate{MinCores: 24, MinThreads: 48, MinSpeed: 2.1},
			RAM:         models.RAMTemplate{MinTotalGB: 64},
			Disks:       models.DiskTemplate{MinCount: 2, MinSizeGB: 470, Type: "SSD"},
			Network:     models.NetworkTemplate{MinTotalNICs: 22},
		}, true

	default:
		return models.SystemTemplate{}, false
	}
}
