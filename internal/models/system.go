package models

type SystemInfo struct {
	CPU     CPUInfo
	RAM     RAMInfo
	Disks   []DiskInfo
	Network []NetworkInfo
	Host    HostInfo
}

type CPUInfo struct {
	Model             string
	Motherboard       string
	MotherboardSerial string
	Cores             int     // ядра
	Threads           int     // потоки
	Speed             float64 // в ГГц
	LoadPercent       float64 // текущая загрузка в процентах
}

type RAMStick struct {
	Slot     string // пр. "DIMM 0"
	Model    string
	Capacity uint64 // объем планки
	Speed    uint32 // частота
}

type RAMInfo struct {
	Total       uint64
	Used        uint64
	Available   uint64
	UsedPercent float64
	Sticks      []RAMStick // список планок ОЗУ
}

type DiskInfo struct {
	Device      string // логическое имя устройства, пр. "/dev/sda1"
	Mountpoint  string // точка монтирования, пр. "/"
	Total       uint64
	Used        uint64
	Free        uint64
	UsedPercent float64
	Model       string  // модель физического диска
	Type        string  // тип диска HDD/SSD/NVMe
	Wear        float64 // износ в процентах
	Status      string  // состояние диска (OK, Warning, Critical)
}

type NetworkInfo struct {
	Name      string
	Interface string
	IP        string
	MAC       string
	Status    string // Up/Down
	Speed     int    // Mbps
	Model     string
}

type HostInfo struct {
	OS     string
	Kernel string
	Uptime uint64
}
