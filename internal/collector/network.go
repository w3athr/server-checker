package collector

import (
	"bufio"
	"bytes"
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

// collectNetworkStatic собирает данные, которые не меняются (модели, MAC, макс. скорость)
// Вызывается один раз при старте.
func collectNetworkStatic() []models.NetworkInfo {
	var netList []models.NetworkInfo

	// ghw.PCI() — тяжелая операция, сканируем один раз
	pci, _ := ghw.PCI(ghw.WithDisableWarnings())

	interfaces, err := os.ReadDir("/sys/class/net")
	if err != nil {
		return netList
	}

	for _, iface := range interfaces {
		name := iface.Name()
		if name == "lo" || !isWiredInterface(name) {
			continue
		}

		info := models.NetworkInfo{
			Interface: name,
			Model:     "Ethernet Controller",
		}

		// 1) MAC адрес
		if b, err := os.ReadFile(filepath.Join("/sys/class/net", name, "address")); err == nil {
			info.MAC = strings.ToUpper(strings.TrimSpace(string(b)))
		}

		// 2) Модель по PCI адресу устройства
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

		// 3) Максимально поддерживаемая скорость (через ethtool)
		if max, err := getMaxSupportedSpeedEthtool(name); err == nil {
			info.MaxSpeedMbps = max
		}

		netList = append(netList, info)
	}

	return netList
}

// updateNetworkDynamic обновляет меняющиеся данные (статус, текущая скорость, IP)
// Вызывается каждую секунду.
func updateNetworkDynamic(staticList []models.NetworkInfo) []models.NetworkInfo {
	// Создаем копию списка, чтобы не портить исходный кэш напрямую во время итерации
	updatedList := make([]models.NetworkInfo, len(staticList))
	copy(updatedList, staticList)

	// Инициализируем трансформер для замены strings.Title
	caser := cases.Title(language.English)

	for i := range updatedList {
		name := updatedList[i].Interface

		// 1) Статус (Up/Down) - заменяем strings.Title на caser.String
		if b, err := os.ReadFile(filepath.Join("/sys/class/net", name, "operstate")); err == nil {
			statusRaw := strings.TrimSpace(string(b))
			updatedList[i].Status = caser.String(statusRaw)
		}

		// 2) Текущая скорость линка
		if b, err := os.ReadFile(filepath.Join("/sys/class/net", name, "speed")); err == nil {
			val, _ := strconv.Atoi(strings.TrimSpace(string(b)))
			if val > 0 {
				updatedList[i].Speed = val
			} else {
				updatedList[i].Speed = 0
			}
		}

		// 3) Текущий IP
		updatedList[i].IP = getInterfaceIP(name)
	}

	return updatedList
}

func isWiredInterface(iface string) bool {
	base := filepath.Join("/sys/class/net", iface)
	if _, err := os.Stat(filepath.Join(base, "wireless")); err == nil {
		return false
	}
	if b, err := os.ReadFile(filepath.Join(base, "type")); err == nil {
		if strings.TrimSpace(string(b)) != "1" {
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

func getMaxSupportedSpeedEthtool(iface string) (int, error) {
	cmd := exec.Command("ethtool", iface)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return 0, err
	}

	max := 0
	sc := bufio.NewScanner(bytes.NewReader(out.Bytes()))
	inSupported := false
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "Supported link modes:") {
			inSupported = true
			rest := strings.TrimPrefix(line, "Supported link modes:")
			if v := parseSpeedFromModeLine(rest); v > max {
				max = v
			}
			continue
		}
		if inSupported {
			if strings.Contains(line, ":") && !strings.HasPrefix(line, "base") {
				break
			}
			if v := parseSpeedFromModeLine(line); v > max {
				max = v
			}
		}
	}
	return max, nil
}

func parseSpeedFromModeLine(line string) int {
	parts := strings.Fields(line)
	max := 0
	for _, p := range parts {
		i := strings.Index(p, "base")
		if i <= 0 {
			continue
		}
		if n, err := strconv.Atoi(p[:i]); err == nil {
			if n > max {
				max = n
			}
		}
	}
	return max
}

func BlinkInterface(name string) {
	exec.Command("ethtool", "--identify", name, "5").Start()
}
