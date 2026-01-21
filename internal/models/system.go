package models

type SystemInfo struct {
	CPU     CPUInfo
	RAM     RAMInfo
	Disks   []DiskInfo
	Network []NetworkInfo
	Host    HostInfo
}

type CPUInfo struct {
	Model       string
	Cores       int     // ядра
	Threads     int     // потоки
	Speed       float64 // в ГГц
	LoadPercent float64 // текущая загрузка в процентах
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
	Free        uint64
	UsedPercent float64
	Sticks      []RAMStick // список планок ОЗУ
}

type DiskInfo struct {
	Device      string
	Mountpoint  string
	Total       uint64
	Used        uint64
	Free        uint64
	UsedPercent float64
	Model       string // потребует доп. вызов
}

type NetworkInfo struct {
	Name string
	IP   string
	MAC  string
	Up   bool
}

type HostInfo struct {
	OS     string
	Kernel string
	Uptime uint64
}
