package collector

import (
	"fmt"
	"net"
	"github.com/jaypipes/ghw"
)

type SystemConfig struct {
	CPUModel        string
	TotalMemoryMB   uint64
	DIMMModules     []uint64
	TotalDiskGB     uint64
	NetworkAdapters []string
}

// CollectConfig собирает информацию о системе и возвращает ее в виде структуры SystemConfig
// или ошибку, если произошла ошибка при сборе информации
func CollectConfig() (*SystemConfig, error) {
	sysCfg := &SystemConfig{}

	// CPU
	cpuInfo, err := ghw.CPU()
	if err != nil {
		return nil, fmt.Errorf("cpu: %w", err)
	}
	if len(cpuInfo.Processors) > 0 {
		sysCfg.CPUModel = cpuInfo.Processors[0].Model
	}

	// Память
	memInfo, err := ghw.Memory()
	if err != nil {
		return nil, fmt.Errorf("memory: %w", err)
	}
	sysCfg.TotalMemoryMB = uint64(memInfo.TotalPhysicalBytes / 1024 / 1024)

	// Проверяем наличие модулей памяти
	if len(memInfo.Modules) > 0 {
		sysCfg.DIMMModules = make([]uint64, len(memInfo.Modules))
		for i, module := range memInfo.Modules {
			sysCfg.DIMMModules[i] = uint64(module.SizeBytes / 1024 / 1024)
		}
	}

	// Блоки (диски)
	blockInfo, err := ghw.Block()
	if err != nil {
		return nil, fmt.Errorf("block: %w", err)
	}
	var totalDiskBytes uint64
	for _, disk := range blockInfo.Disks {
		totalDiskBytes += disk.SizeBytes
	}
	sysCfg.TotalDiskGB = totalDiskBytes / 1024 / 1024 / 1024

	// Сетевые адаптеры
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to list interfaces: %w", err)
	}
	for _, iface := range ifaces {
		if iface.HardwareAddr.String() != "" {
			sysCfg.NetworkAdapters = append(sysCfg.NetworkAdapters, iface.Name)
		}
	}
	return sysCfg, nil
}

// FormatConfigLines преобразует SystemConfig в удобный для вывода срез строк
func FormatConfigLines(cfg *SystemConfig) []string {
	lines := []string{
		fmt.Sprintf("CPU Model: %s", cfg.CPUModel),
		fmt.Sprintf("Total Memory: %d MB", cfg.TotalMemoryMB),
	}
	// детали по DIMM
	for i, size := range cfg.DIMMModules {
		lines = append(lines, fmt.Sprintf(" DIMM %d: %d MB", i+1, size))
	}
	lines = append(lines, fmt.Sprintf("Total Disk: %d GB", cfg.TotalDiskGB))
	// по каждому NIC
	for _, nic := range cfg.NetworkAdapters {
		lines = append(lines, fmt.Sprintf("NIC: %s", nic))
	}
	return lines
}
