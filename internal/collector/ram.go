package collector

import (
	"os/exec"
	"server-checker/internal/models"
	"strconv"
	"strings"

	"github.com/jaypipes/ghw"
	"github.com/shirou/gopsutil/v3/mem"
)

func collectRAM() models.RAMInfo {
	var info models.RAMInfo

	v, _ := mem.VirtualMemory()
	info.Total = v.Total
	info.Used = v.Used
	info.Free = v.Free
	info.UsedPercent = v.UsedPercent

	memory, err := ghw.Memory()
	if err == nil {
		for _, mod := range memory.Modules {
			info.Sticks = append(info.Sticks, models.RAMStick{
				Slot:     mod.Label,
				Model:    mod.Vendor,
				Capacity: uint64(mod.SizeBytes),
				Speed:    getRAMSpeedManual(),
			})
		}
	}

	return info
}

func getRAMSpeedManual() uint32 {
	// Для Windows (через PowerShell/WMIC)
	// wmic memorychip get speed
	cmd := exec.Command("wmic", "memorychip", "get", "speed")
	out, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 1 {
			speedStr := strings.TrimSpace(lines[1])
			val, _ := strconv.Atoi(speedStr)
			return uint32(val)
		}
	}

	// Для Linux (требует sudo для dmidecode)
	// Если мы не в sudo, это вернет 0
	cmd = exec.Command("sh", "-c", "dmidecode -t memory | grep 'Speed' | head -n 1 | awk '{print $2}'")
	out, err = cmd.Output()
	if err == nil {
		speedStr := strings.TrimSpace(string(out))
		val, _ := strconv.Atoi(speedStr)
		return uint32(val)
	}

	return 0 // Если не удалось определить
}
