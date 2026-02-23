package models

type SystemTemplate struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	CPU         CPUTemplate     `yaml:"cpu"`
	RAM         RAMTemplate     `yaml:"ram"`
	Disks       DiskTemplate    `yaml:"disks"`   // Изменили структуру для массовой проверки
	Network     NetworkTemplate `yaml:"network"` // Изменили структуру
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
	MinCount  int    `yaml:"min_count"`
	MinSizeGB int    `yaml:"min_size_gb"`
	Type      string `yaml:"type"` // SSD, HDD, NVMe
}

type NetworkTemplate struct {
	MinTotalNICs int `yaml:"min_total_nics"`
	MinSpeedMbps int `yaml:"min_speed_mbps"` // Минимальная скорость основного пула портов
}
