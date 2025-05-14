package collector

import {
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
}

type Metrics struct {
	CPUPercent float64
	MemUsedMB  uint64
	MemTotalMB uint64
}

type metricsMsg struct {
	cpu      float64
	memUsed  uint64
	memTotal uint64
	err      error
}

 // fetchMetrics замеряет CPU/RAM за 1 секунду и возвращает metricsMsg
 func fetchMetrics() tea.Msg {
     cpuPerc, err1 := cpu.Percent(1*time.Second, false)
     memStats, err2 := mem.VirtualMemory()
     if err1 != nil || err2 != nil {
         return metricsMsg{err: fmt.Errorf("metrics error: %v %v", err1, err2)}
     }
     return metricsMsg{
         cpu:     cpuPerc[0],
         memUsed: memStats.Used / (1024 * 1024),  // в МБ
         memTotal: memStats.Total / (1024 * 1024), // в МБ
     }
 }


func GetMetrics(interval time.Duration) (*Metrics, error) {
	cpuPerc, err1 := cpu.Percent(interval, false)
	memStats, err2 := mem.VirtualMemory()
	if err1 != nil || err2 != nil {
		return nil, fmt.Errorf("metrics error: %v %v", err1, err2)
	}
	return &Metrics{
		CPUPercent: cpuPerc[0],
		MemUsedMB:  memStats.Used / 1024 / 1024,
		MemTotalMB: memStats.Total / 1024 / 1024,
	}, nil
}
