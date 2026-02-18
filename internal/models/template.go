package models

// SystemTemplate описывает ожидаемую конфигурацию системы (эталон).
type SystemTemplate struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	CPU         CPUTemplate       `yaml:"cpu"`
	RAM         RAMTemplate       `yaml:"ram"`
	Disks       []DiskTemplate    `yaml:"disks"`
	Network     []NetworkTemplate `yaml:"network"`
}

type CPUTemplate struct {
	MinCores   int     `yaml:"min_cores"`
	MinThreads int     `yaml:"min_threads"`
	MinSpeed   float64 `yaml:"min_speed_ghz"` // Минимальная частота
}

type RAMTemplate struct {
	MinTotalGB int `yaml:"min_total_gb"` // Минимальный объем в ГБ
}

type DiskTemplate struct {
	Name      string `yaml:"name"` // Например, "System Disk" или "/dev/sda" (опционально)
	MinSizeGB int    `yaml:"min_size_gb"`
	Type      string `yaml:"type"` // SSD, HDD, NVMe
}

type NetworkTemplate struct {
	InterfaceName string `yaml:"interface_name"` // eth0, enp3s0 (опционально)
	MinSpeed      int    `yaml:"min_speed_mbps"`
}
