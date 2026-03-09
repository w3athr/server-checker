package models

// SystemTemplate описывает эталонную конфигурацию ПАК.
type SystemTemplate struct {
	Name        string          `yaml:"name"`        // имя шаблона
	Description string          `yaml:"description"` // описание шаблона
	CPU         CPUTemplate     `yaml:"cpu"`
	RAM         RAMTemplate     `yaml:"ram"`
	Disks       DiskTemplate    `yaml:"disks"`
	Network     NetworkTemplate `yaml:"network"`
}

// Минимальные требования к процессору
type CPUTemplate struct {
	MinCores   int     `yaml:"min_cores"`     // мин. количество ядер
	MinThreads int     `yaml:"min_threads"`   // мин. количество потоков
	MinSpeed   float64 `yaml:"min_speed_ghz"` // мин. частота процессора
}

// Минимальные требования к ОЗУ
type RAMTemplate struct {
	MinTotalGB int `yaml:"min_total_gb"` // мин. общий объем ОЗУ
}

// Минимальные требования к дискам
type DiskTemplate struct {
	MinCount  int    `yaml:"min_count"`   // Сколько дисков должно быть
	MinSizeGB int    `yaml:"min_size_gb"` // Минимальный размер каждого
	Type      string `yaml:"type"`        // SSD, NVMe
}

// Минимальные требования к сетевым интерфейсам
type NetworkTemplate struct {
	MinTotalNICs int `yaml:"min_total_nics"` // Общее кол-во портов
}
