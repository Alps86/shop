package performance_test

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

func (current *world) validateEnvironment() error {
	endpoint, err := url.Parse(current.public)
	if err != nil {
		return err
	}

	address := net.ParseIP(endpoint.Hostname())
	if address == nil {
		return fmt.Errorf("Messziel benötigt explizite Loopback-IP")
	}

	if !address.IsLoopback() {
		return fmt.Errorf("Messprofil benötigt Loopback")
	}

	return current.validateProfileEnvironment()
}

func (current *world) validateHardware() error {
	minimum, err := strconv.Atoi(current.data["MinimumCPU"])
	if err != nil {
		return err
	}

	if runtime.NumCPU() < minimum {
		return fmt.Errorf("CPU %d unter Referenzprofil %d", runtime.NumCPU(), minimum)
	}

	memory, err := memoryKiB()
	if err != nil {
		return err
	}

	minimumRAM, err := strconv.Atoi(current.data["MinimumRAMGiB"])
	if err != nil {
		return err
	}

	return requireMinimumRAM(memory, int64(minimumRAM))
}

func memoryKiB() (int64, error) {
	content, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, err
	}

	fields := strings.Fields(string(content))
	if fields[0] != "MemTotal:" {
		return 0, fmt.Errorf("MemTotal fehlt")
	}

	return strconv.ParseInt(fields[1], 10, 64)
}

func requireMinimumRAM(actual, minimum int64) error {
	if actual < minimum*1024*1024 {
		return fmt.Errorf("RAM %d KiB unter Referenzprofil %d GiB", actual, minimum)
	}

	return nil
}

func (current *world) validateProfileEnvironment() error {
	if current.data["Network"] != "loopback" {
		return fmt.Errorf("unbekanntes Netzwerkprofil %s", current.data["Network"])
	}

	if strings.TrimSpace(os.Getenv("SHOP_PERFORMANCE_IMAGES")) == "" {
		return fmt.Errorf("SHOP_PERFORMANCE_IMAGES muss tatsächliche Versionen und native/Compose-Grenze ausweisen")
	}

	return current.validateHardware()
}

func operatingSystem() (map[string]string, error) {
	kernel, err := exec.Command("uname", "-srvm").Output()
	if err != nil {
		return nil, err
	}

	distribution, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return nil, err
	}

	return map[string]string{"Kernel": strings.TrimSpace(string(kernel)), "Distribution": string(distribution)}, nil
}
