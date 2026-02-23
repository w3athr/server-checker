package models

// SystemTemplate описывает эталонную конфигурацию ПАК.
type SystemTemplate struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	CPU         CPUTemplate     `yaml:"cpu"`
	RAM         RAMTemplate     `yaml:"ram"`
	Disks       DiskTemplate    `yaml:"disks"`
	Network     NetworkTemplate `yaml:"network"`
}

type CPUTemplate struct {
	MinCores   int     `yaml:"min_cores"`
	MinThreads int     `yaml:"min_threads"`
	MinSpeed   float64 `yaml:"min_speed_ghz"`
}

type RAMTemplate struct {
	MinTotalGB int `yaml:"min_total_gb"`
}

type DiskTemplate struct {
	MinCount  int    `yaml:"min_count"`   // Сколько дисков должно быть
	MinSizeGB int    `yaml:"min_size_gb"` // Минимальный размер каждого
	Type      string `yaml:"type"`        // SSD, NVMe
}

type NetworkTemplate struct {
	MinTotalNICs int `yaml:"min_total_nics"` // Общее кол-во портов
}
