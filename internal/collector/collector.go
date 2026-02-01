package collector

import (
	"server-checker/internal/models"
	"sync"

	"github.com/shirou/gopsutil/v3/host"
)

var (
	staticData   models.SystemInfo
	physDiskData map[string]models.DiskInfo
	once         sync.Once
)

// сбор статических данных один раз, динамические при каждом вызове
func CollectAll() (models.SystemInfo, error) {

	once.Do(func() { // sync.Once - выполнение только один раз
		h, _ := host.Info()
		// системная информация (ос, ядро)
		staticData.Host = models.HostInfo{
			OS:     h.OS,
			Kernel: h.KernelVersion,
		}
		// CPU информация (модель, ядра, потоки)
		staticData.CPU = collectCPUStatic()
		// информация о планках ОЗУ не меняется
		staticData.RAM.Sticks = getPhysicalRAM()
		// информация о дисках (модель, тип, износ)
		physDiskData = getPhysicalDriveData()
	})

	// копирование статических данных для дальнейшего обновления динамических
	currentInfo := staticData
	// динамические данные CPU (нагрузка)
	currentInfo.CPU.LoadPercent = collectCPULoad()
	// динамические данные RAM (занято, свободно)
	currentInfo.RAM = updateRAMDynamic(currentInfo.RAM)
	// динамические данные дисков (занято, свободно)
	currentInfo.Disks = collectDisks()
	// время работы системы
	u, _ := host.Uptime()
	currentInfo.Host.Uptime = u

	return currentInfo, nil
}
