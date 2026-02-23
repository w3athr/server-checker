package collector

import (
	"bufio"
	"bytes"
	"fmt"
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

	pci, _ := ghw.PCI(ghw.WithDisableWarnings())

	interfaces, err := os.ReadDir("/sys/class/net")
	if err != nil {
		return netList
	}

	for _, iface := range interfaces {
		name := iface.Name()
		if name == "lo" {
			continue
		}

		// Пропускаем Wi-Fi и не-Ethernet
		if !isWiredInterface(name) {
			continue
		}

		info := models.NetworkInfo{
			Interface:    name,
			Status:       "Down",
			Speed:        0,
			MaxSpeedMbps: 0,
			Model:        "",
			IP:           "",
			MAC:          "",
			Name:         "",
		}

		// 1) Статус (Up/Down)
		if b, err := os.ReadFile(filepath.Join("/sys/class/net", name, "operstate")); err == nil {
			info.Status = cases.Title(language.English).String(strings.TrimSpace(string(b)))
		}

		// 2) MAC
		if b, err := os.ReadFile(filepath.Join("/sys/class/net", name, "address")); err == nil {
			info.MAC = strings.ToUpper(strings.TrimSpace(string(b)))
		}

		// 3) Текущая скорость линка (если есть)
		if b, err := os.ReadFile(filepath.Join("/sys/class/net", name, "speed")); err == nil {
			val, _ := strconv.Atoi(strings.TrimSpace(string(b)))
			if val > 0 {
				info.Speed = val
			}
		}

		// 4) IP
		info.IP = getInterfaceIP(name)

		// 5) Модель по PCI
		if link, err := os.Readlink(filepath.Join("/sys/class/net", name, "device")); err == nil {
			parts := strings.Split(link, "/")
			pciAddr := parts[len(parts)-1]

			if pci != nil {
				device := pci.GetDevice(pciAddr)
				if device != nil && device.Product != nil {
					info.Model = device.Product.Name
				}
			}
		}
		if info.Model == "" {
			info.Model = "Ethernet Controller"
		}

		// 6) Максимально поддерживаемая скорость через ethtool
		// Требует установленного ethtool. Root обычно не обязателен для чтения, но ок.
		if max, err := getMaxSupportedSpeedEthtool(name); err == nil {
			info.MaxSpeedMbps = max
		}

		netList = append(netList, info)
	}

	return netList
}

func isWiredInterface(iface string) bool {
	base := filepath.Join("/sys/class/net", iface)

	// если есть wireless каталог - это Wi-Fi
	if _, err := os.Stat(filepath.Join(base, "wireless")); err == nil {
		return false
	}

	// тип интерфейса (1 = Ethernet) см. /sys/class/net/<iface>/type
	if b, err := os.ReadFile(filepath.Join(base, "type")); err == nil {
		t := strings.TrimSpace(string(b))
		if t != "1" {
			return false
		}
	}

	return true
}

func getInterfaceIP(name string) string {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return "N/A"
	}
	addrs, _ := iface.Addrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "No IP"
}

// Парсим Supported link modes из `ethtool <iface>` и берем максимум
func getMaxSupportedSpeedEthtool(iface string) (int, error) {
	cmd := exec.Command("ethtool", iface)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return 0, err
	}

	max := 0
	sc := bufio.NewScanner(bytes.NewReader(out.Bytes()))

	inSupported := false
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())

		// Вход в секцию
		if strings.HasPrefix(line, "Supported link modes:") {
			inSupported = true
			// после двоеточия может быть кусок на той же строке
			rest := strings.TrimSpace(strings.TrimPrefix(line, "Supported link modes:"))
			if rest != "" {
				if v := parseSpeedFromModeLine(rest); v > max {
					max = v
				}
			}
			continue
		}

		if !inSupported {
			continue
		}

		// Секция закончилась, когда пошли другие ключи ethtool
		// Например: "Supported pause frame use:", "Supports auto-negotiation:"
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "base") {
			break
		}

		if v := parseSpeedFromModeLine(line); v > max {
			max = v
		}
	}

	if max == 0 {
		return 0, fmt.Errorf("supported link modes not found")
	}
	return max, nil
}

// Пример строк:
// "1000baseT/Full"
// "10baseT/Half 10baseT/Full"
// "2500baseT/Full"
// "10000baseT/Full"
func parseSpeedFromModeLine(line string) int {
	// выцепляем все токены вида "<digits>base"
	parts := strings.Fields(line)
	max := 0

	for _, p := range parts {
		// иногда в одной строке может быть "1000baseT/Full"
		i := strings.Index(p, "base")
		if i <= 0 {
			continue
		}
		num := p[:i]
		n, err := strconv.Atoi(num)
		if err != nil {
			continue
		}
		// n уже в Mbps (1000, 2500, 10000)
		if n > max {
			max = n
		}
	}

	return max
}

// BlinkInterface заставляет индикатор порта мигать (помогает найти кабель в стойке)
func BlinkInterface(name string) {
	exec.Command("ethtool", "--identify", name, "5").Start()
}
