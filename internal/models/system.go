package models

// Структура для хранения всей собранной информации
type SystemInfo struct {
	CPU       CPUInfo
	RAM       RAMInfo
	Disks     []DiskInfo
	BlockDevs []BlockDeviceInfo
	Network   []NetworkInfo
	Host      HostInfo
}

// Информация о процессоре, материнской плате и серийном номере
type CPUInfo struct {
	Model             string  // модель процессора
	Motherboard       string  // модель материнской платы
	MotherboardSerial string  // серийный номер материнской платы
	Cores             int     // количество физических ядер
	Threads           int     // количество логических потоков
	Speed             float64 // базовая частота в ГГц
	LoadPercent       float64 // текущая загрузка в процентах
	HardwareVendor    string  // производитель системы
	HardwareModel     string  // модель сервера/платформы
	HardwareSN        string  // серийный номер устройства
}

// Информация о физических планках оперативной памяти
type RAMStick struct {
	Slot     string // пр. "DIMM 0"
	Model    string // модель планки
	Capacity uint64 // объем планки
	Speed    uint32 // частота
}

// Вся информация по оперативной памяти
type RAMInfo struct {
	Total       uint64     // общий объем
	Used        uint64     // используемый объем
	Available   uint64     // доступная память
	UsedPercent float64    // процент использования
	Sticks      []RAMStick // список планок ОЗУ
}

// Информация о файловых системах и разделах
type DiskInfo struct {
	Device      string  // логическое имя устройства, пр. "/dev/sda1"
	Mountpoint  string  // точка монтирования, пр. "/"
	Total       uint64  // общий размер
	Used        uint64  // занятое пространство
	Free        uint64  // свободное пространство
	UsedPercent float64 // процент использования
}

// Информация о блочных устройствах (дисках)
type BlockDeviceInfo struct {
	Device          string  // "/dev/sda", "/dev/nvme0n1"
	Model           string  // модель
	Type            string  // HDD/SSD/NVMe
	Size            uint64  // общий размер устройства
	Wear            float64 // износ в процентах
	Status          string  // состояние диска (OK, Warning, Critical)
	PowerCycles     uint64  // количество включений
	PowerOnHours    uint64  // часов работы
	UnsafeShutdowns uint64  // небезопасных выключений
}

// Информация о сетевых интерфейсах
type NetworkInfo struct {
	Name         string // название интерфейса
	Interface    string // системное имя интерфейса, пр. eth0
	IP           string
	MAC          string
	Status       string // Up/Down
	Speed        int    // текущая скорость линка, Mbps
	MaxSpeedMbps int    // максимально, поддерживаемая, Mbps
	Model        string // модель сетевого адаптера
}

type HostInfo struct {
	OS     string // название операционной системы
	Kernel string // версия ядра
	Uptime uint64 // время работы системы
}
