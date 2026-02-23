package collector

import (
	"bufio"
	"os"
	"server-checker/internal/models"
	"strings"
	"sync"

	"github.com/shirou/gopsutil/v3/host"
)

var (
	// Используем одно имя переменной для кэша, чтобы не путаться
	cache      models.SystemInfo
	cacheMutex sync.RWMutex
	once       sync.Once
)

// CollectAll теперь разделяет статику и динамику
func CollectAll() (models.SystemInfo, error) {

	once.Do(func() {
		// --- СТАТИЧЕСКИЕ ДАННЫЕ (собираем 1 раз) ---
		h, _ := host.Info()

		cache.Host = models.HostInfo{
			OS:     readOSRelease(),
			Kernel: h.KernelVersion,
		}

		// Информация о железе
		cache.CPU = collectCPUStatic()
		cache.RAM.Sticks = getPhysicalRAM()

		// Диски (модели и SMART SATA/NVMe собираем один раз при старте)
		pData := getPhysicalDriveData()
		cache.BlockDevs = collectBlockDevices(pData)

		// Сеть (статическая часть: модели PCI, MAC, Макс. скорость)
		cache.Network = collectNetworkStatic()
	})

	// Блокируем кэш для обновления динамических данных
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// --- ДИНАМИЧЕСКИЕ ДАННЫЕ (каждую секунду) ---

	// CPU нагрузка
	cache.CPU.LoadPercent = collectCPULoad()

	// RAM занято/свободно
	cache.RAM = updateRAMDynamic(cache.RAM)

	// Свободное место на разделах
	cache.Disks = collectDisks()

	// Статус сети (Up/Down, текущая скорость, IP)
	cache.Network = updateNetworkDynamic(cache.Network)

	// Аптайм
	u, _ := host.Uptime()
	cache.Host.Uptime = u

	return cache, nil
}

// Парсинг Pretty Name версии дистрибутива
func readOSRelease() string {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var prettyName string
	var name string
	var version string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "PRETTY_NAME=") {
			prettyName = trimValue(line)
		}
		if strings.HasPrefix(line, "NAME=") {
			name = trimValue(line)
		}
		if strings.HasPrefix(line, "VERSION=") {
			version = trimValue(line)
		}
	}

	if prettyName != "" {
		return prettyName
	}
	if name != "" && version != "" {
		return name + " " + version
	}
	return name
}

func trimValue(line string) string {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return ""
	}
	val := strings.Trim(parts[1], `"`)
	return val
}
