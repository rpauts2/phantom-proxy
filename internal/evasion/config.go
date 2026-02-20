package evasion

// Config holds evasion parameters for external implant/payload tools
// PhantomProxy не реализует техники обхода — передаёт параметры в Sliver, Havoc, Metasploit и т.д.
type Config struct {
	// Sleep obfuscation: обфускация времени sleep для обхода sandbox
	SleepObfuscation   bool   `yaml:"sleep_obfuscation" mapstructure:"sleep_obfuscation"`
	SleepObfuscationMethod string `yaml:"sleep_method" mapstructure:"sleep_method"` // ekko, return_address, etc.

	// Sandbox/VM detection: изменить поведение при детекте
	SandboxEvasion     bool   `yaml:"sandbox_evasion" mapstructure:"sandbox_evasion"`
	SandboxEvasionMode string `yaml:"sandbox_mode" mapstructure:"sandbox_mode"` // exit, delay, mimic

	// Инструкции для генератора имплантов (Sliver, Havoc):
	// - amsi_bypass: включить в конфиг импланта
	// - etw_patch: включить в конфиг
	// - process_injection: метод (CreateRemoteThread, NtMapViewOfSection, etc.)
	AMSIBypass   bool   `yaml:"amsi_bypass" mapstructure:"amsi_bypass"`
	ETWPatch     bool   `yaml:"etw_patch" mapstructure:"etw_patch"`
	ProcessInjection string `yaml:"process_injection" mapstructure:"process_injection"` // none, CreateRemoteThread, NtMapViewOfSection

	// Syscall usage вместо WinAPI (для EDR evasion)
	SyscallUsage bool   `yaml:"syscall_usage" mapstructure:"syscall_usage"`
}

// DefaultConfig returns safe defaults
func DefaultConfig() *Config {
	return &Config{
		SleepObfuscation:   false,
		SandboxEvasion:     false,
		AMSIBypass:         false,
		ETWPatch:           false,
		ProcessInjection:   "none",
		SyscallUsage:       false,
	}
}

// ToSliverImplantConfig converts to Sliver implant build flags
func (c *Config) ToSliverImplantConfig() map[string]interface{} {
	return map[string]interface{}{
		"skip_sandbox":    c.SandboxEvasion,
		"evasion":         c.SleepObfuscation || c.AMSIBypass || c.ETWPatch,
	}
}
