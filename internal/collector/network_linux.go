//go:build linux

package collector

import (
	"os"
	"os/exec"
	"path/filepath"
	"server-checker/internal/models"
	"strconv"
	"strings"

	"github.com/jaypipes/ghw"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func collectNetwork() []models.NetworkInfo {
	var netList []models.NetworkInfo

	pci, _ := ghw.PCI()

	interfaces, _ := os.ReadDir("/sys/class/net")
	for _, iface := range interfaces {
		name := iface.Name()
		if name == "lo" {
			continue
		}

		info := models.NetworkInfo{
			Interface: name,
			Status:    "Down",
			Speed:     0,
		}

		if b, err := os.ReadFile(filepath.Join("/sys/class/net", name, "operstate")); err == nil {
			info.Status = cases.Title(language.English).String(strings.TrimSpace(string(b)))
		}

		if b, err := os.ReadFile(filepath.Join("/sys/class/net", name, "address")); err == nil {
			info.MAC = cases.Title(language.English).String(strings.TrimSpace(string(b)))
		}

		if b, err := os.ReadFile(filepath.Join("/sys/class/net", name, "speed")); err == nil {
			val, _ := strconv.Atoi(strings.TrimSpace(string(b)))
			info.Speed = val
		}

		if link, err := os.Readlink(filepath.Join("/sys/class/net", name, "device")); err == nil {
			parts := strings.Split(link, "/")
			pciAddr := parts[len(parts)-1]

			if pci != nil {
				device := pci.GetDevice(pciAddr)
				if device != nil {
					info.Model = device.Product.Name
				}
			}
		}
		if info.Model == "" {
			info.Model = "Ethernet Controller"
		}

		netList = append(netList, info)
	}
	return netList
}

func BlinkInterface(name string) {
	// Просто запускаем команду в фоне, не дожидаясь ответа
	exec.Command("ethtool", "--identify", name, "5").Start()
}
