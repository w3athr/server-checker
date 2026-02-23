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
	staticData models.SystemInfo
	physData   map[string]physDiskData
	once       sync.Once
)

// сбор статических данных один раз, динамические при каждом вызове
func CollectAll() (models.SystemInfo, error) {

	once.Do(func() { // sync.Once - выполнение только один раз
		h, _ := host.Info()
		// системная информация (ос, ядро)
		staticData.Host = models.HostInfo{
			OS:     readOSRelease(),
			Kernel: h.KernelVersion,
		}
		// CPU информация (модель, ядра, потоки)
		staticData.CPU = collectCPUStatic()
		// информация о планках ОЗУ не меняется
		staticData.RAM.Sticks = getPhysicalRAM()
		// информация о дисках (модель, тип, износ)
		physData = getPhysicalDriveData()
		staticData.BlockDevs = collectBlockDevices(physData)
	})

	// копирование статических данных для дальнейшего обновления динамических
	currentInfo := staticData
	// динамические данные CPU (нагрузка)
	currentInfo.CPU.LoadPercent = collectCPULoad()
	// динамические данные RAM (занято, свободно)
	currentInfo.RAM = updateRAMDynamic(currentInfo.RAM)
	// динамические данные дисков (занято, свободно)
	currentInfo.Disks = collectDisks()
	// динамические данные сети (статусы, ip)
	currentInfo.Network = collectNetwork()
	// время работы системы
	u, _ := host.Uptime()
	currentInfo.Host.Uptime = u

	return currentInfo, nil
}

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
