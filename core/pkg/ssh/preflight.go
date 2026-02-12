package ssh

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// PreflightReport contains the results of all pre-flight checks on a server.
type PreflightReport struct {
	OS            string   `json:"os"`
	Distribution  string   `json:"distribution"`
	KernelVersion string   `json:"kernel_version"`
	Arch          string   `json:"arch"`
	CPUCores      int      `json:"cpu_cores"`
	RAMBytes      int64    `json:"ram_bytes"`
	CgroupsV2     bool     `json:"cgroups_v2"`
	Compatible    bool     `json:"compatible"`
	Errors        []string `json:"errors,omitempty"`
}

// SupportedDistros lists the Linux distributions compatible with K3s.
var SupportedDistros = []string{"ubuntu", "debian", "rhel", "centos", "rocky", "almalinux"}

// RunPreflightCheck performs a comprehensive pre-flight audit on a remote server.
func RunPreflightCheck(client *Client) (*PreflightReport, error) {
	report := &PreflightReport{}
	var errors []string

	// 1. Detect OS
	result, err := client.ExecuteCommand("cat /etc/os-release")
	if err != nil {
		return nil, fmt.Errorf("failed to detect OS: %w", err)
	}
	parseOSRelease(result.Stdout, report)

	// 2. Check if distribution is supported
	supported := false
	distLower := strings.ToLower(report.Distribution)
	for _, d := range SupportedDistros {
		if strings.Contains(distLower, d) {
			supported = true
			break
		}
	}
	if !supported {
		errors = append(errors, fmt.Sprintf("unsupported distribution: %s", report.Distribution))
	}

	// 3. Detect kernel version
	result, err = client.ExecuteCommand("uname -r")
	if err != nil {
		errors = append(errors, fmt.Sprintf("failed to get kernel version: %v", err))
	} else {
		report.KernelVersion = strings.TrimSpace(result.Stdout)
	}

	// 4. Detect CPU architecture
	result, err = client.ExecuteCommand("uname -m")
	if err != nil {
		errors = append(errors, fmt.Sprintf("failed to detect architecture: %v", err))
	} else {
		report.Arch = strings.TrimSpace(result.Stdout)
	}

	// 5. Detect CPU cores
	result, err = client.ExecuteCommand("nproc")
	if err != nil {
		errors = append(errors, fmt.Sprintf("failed to detect CPU cores: %v", err))
	} else {
		cores, parseErr := strconv.Atoi(strings.TrimSpace(result.Stdout))
		if parseErr != nil {
			errors = append(errors, fmt.Sprintf("failed to parse CPU cores: %v", parseErr))
		} else {
			report.CPUCores = cores
		}
	}

	// 6. Detect RAM
	result, err = client.ExecuteCommand("grep MemTotal /proc/meminfo | awk '{print $2}'")
	if err != nil {
		errors = append(errors, fmt.Sprintf("failed to detect RAM: %v", err))
	} else {
		kbStr := strings.TrimSpace(result.Stdout)
		kb, parseErr := strconv.ParseInt(kbStr, 10, 64)
		if parseErr != nil {
			errors = append(errors, fmt.Sprintf("failed to parse RAM: %v", parseErr))
		} else {
			report.RAMBytes = kb * 1024 // Convert KB to bytes
		}
	}

	// 7. Check cgroups v2
	result, err = client.ExecuteCommand("stat -fc %T /sys/fs/cgroup")
	if err != nil {
		errors = append(errors, fmt.Sprintf("failed to check cgroups: %v", err))
	} else {
		cgroupType := strings.TrimSpace(result.Stdout)
		report.CgroupsV2 = cgroupType == "cgroup2fs"
	}

	// 8. Check essential kernel modules
	requiredModules := []string{"overlay", "br_netfilter"}
	for _, mod := range requiredModules {
		cmd := fmt.Sprintf("lsmod | grep -q %s && echo 'loaded' || echo 'not_loaded'", mod)
		result, err = client.ExecuteCommand(cmd)
		if err != nil || strings.TrimSpace(result.Stdout) != "loaded" {
			errors = append(errors, fmt.Sprintf("kernel module %s is not loaded", mod))
		}
	}

	report.Errors = errors
	report.Compatible = len(errors) == 0

	return report, nil
}

// ToJSON serializes the PreflightReport to a JSON string.
func (r *PreflightReport) ToJSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// parseOSRelease parses /etc/os-release content and sets report fields.
func parseOSRelease(content string, report *PreflightReport) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := strings.Trim(parts[1], "\"")

		switch key {
		case "ID":
			report.Distribution = value
		case "PRETTY_NAME":
			report.OS = value
		}
	}
}
