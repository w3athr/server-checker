package collector

import (
	"net"
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

	// Получаем информацию о PCI устройствах один раз для определения моделей карт
	pci, _ := ghw.PCI(ghw.WithDisableWarnings())

	// Читаем интерфейсы из системной директории Linux
	interfaces, err := os.ReadDir("/sys/class/net")
	if err != nil {
		return netList
	}

	for _, iface := range interfaces {
		name := iface.Name()
		if name == "lo" {
			continue
		} // Пропускаем петлю

		info := models.NetworkInfo{
			Interface: name,
			Status:    "Down",
			Speed:     0,
		}

		// 1. Статус (Up/Down)
		if b, err := os.ReadFile(filepath.Join("/sys/class/net", name, "operstate")); err == nil {
			info.Status = cases.Title(language.English).String(strings.TrimSpace(string(b)))
		}

		// 2. MAC адрес
		if b, err := os.ReadFile(filepath.Join("/sys/class/net", name, "address")); err == nil {
			info.MAC = strings.ToUpper(strings.TrimSpace(string(b)))
		}

		// 3. Скорость порта (в Mbps)
		if b, err := os.ReadFile(filepath.Join("/sys/class/net", name, "speed")); err == nil {
			val, _ := strconv.Atoi(strings.TrimSpace(string(b)))
			info.Speed = val
		}

		// 4. IP адрес (используем стандартный пакет net)
		info.IP = getInterfaceIP(name)

		// 5. Определение модели через PCI адрес
		if link, err := os.Readlink(filepath.Join("/sys/class/net", name, "device")); err == nil {
			// link обычно выглядит как ../../../0000:03:00.0
			parts := strings.Split(link, "/")
			pciAddr := parts[len(parts)-1]

			if pci != nil {
				device := pci.GetDevice(pciAddr)
				if device != nil {
					info.Model = device.Product.Name
				}
			}
		}

		// Заглушка, если модель не нашлась
		if info.Model == "" {
			info.Model = "Ethernet Controller"
		}

		netList = append(netList, info)
	}
	return netList
}

// Вспомогательная функция для получения IP
func getInterfaceIP(name string) string {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return "N/A"
	}
	addrs, _ := iface.Addrs()
	for _, addr := range addrs {
		// Берем первый IPv4 адрес
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "No IP"
}

// BlinkInterface заставляет индикатор порта мигать (помогает найти кабель в стойке)
func BlinkInterface(name string) {
	// Требует установленного ethtool и прав sudo
	exec.Command("ethtool", "--identify", name, "5").Start()
}
