package services

import (
	"bufio"
	"os"
	"os/exec"
	"strings"

	"github.com/itsHenry35/SambaManager/config"
	"github.com/itsHenry35/SambaManager/types"
)

// SystemService handles system environment checks and configuration
type SystemService struct{}

// NewSystemService creates a new system service
func NewSystemService() *SystemService {
	return &SystemService{}
}

// CheckEnvironment performs comprehensive system environment checks
func (s *SystemService) CheckEnvironment() (*types.SystemCheckResponse, error) {
	checks := []types.CheckResult{}

	// Check 1: Samba installed
	checks = append(checks, s.checkSambaInstalled())

	// Check 2: smbpasswd command available
	checks = append(checks, s.checkSmbpasswdAvailable())

	// Check 3: testparm available
	checks = append(checks, s.checkTestparmAvailable())

	// Check 4: smbclient available
	checks = append(checks, s.checkSmbclientAvailable())

	// Check 5: smb.conf exists and readable
	checks = append(checks, s.checkSmbConfExists())

	// Check 6: libnss-extrausers installed (Linux only)
	checks = append(checks, s.checkNssExtrausers())

	// Check 7: extrausers directory structure
	checks = append(checks, s.checkExtrausersDirectory())

	// Check 8: nsswitch.conf configuration
	checks = append(checks, s.checkNsswitchConf())

	// Check 9: Samba service running
	checks = append(checks, s.checkSambaService())

	// Check 10: Home directory exists
	checks = append(checks, s.checkHomeDirectory())

	// Check 11: Samba configuration valid
	checks = append(checks, s.checkSambaConfigValid())

	// Check 12: Passdb backend configured
	checks = append(checks, s.checkPassdbBackend())

	// Check 13: Security mode configured
	checks = append(checks, s.checkSecurityMode())

	// Check 14: Force user exists
	checks = append(checks, s.checkForceUser())

	// Check 15: Force group exists
	checks = append(checks, s.checkForceGroup())

	// Check 16: Running as root
	checks = append(checks, s.checkRunningAsRoot())

	// Check 17: Access based share enum enabled
	checks = append(checks, s.checkAccessBasedShareEnum())

	return &types.SystemCheckResponse{
		Checks: checks,
	}, nil
}

func (s *SystemService) checkSambaInstalled() types.CheckResult {
	cmd := exec.Command("which", "smbd")
	err := cmd.Run()

	if err != nil {
		return types.CheckResult{
			ID:     "samba-installed",
			Status: "fail",
		}
	}

	return types.CheckResult{
		ID:     "samba-installed",
		Status: "pass",
	}
}

func (s *SystemService) checkSmbpasswdAvailable() types.CheckResult {
	cmd := exec.Command("which", "smbpasswd")
	err := cmd.Run()

	if err != nil {
		return types.CheckResult{
			ID:     "smbpasswd-command",
			Status: "fail",
		}
	}

	return types.CheckResult{
		ID:     "smbpasswd-command",
		Status: "pass",
	}
}

func (s *SystemService) checkTestparmAvailable() types.CheckResult {
	cmd := exec.Command("which", "testparm")
	err := cmd.Run()

	if err != nil {
		return types.CheckResult{
			ID:     "testparm-command",
			Status: "warning",
		}
	}

	return types.CheckResult{
		ID:     "testparm-command",
		Status: "pass",
	}
}

func (s *SystemService) checkSmbclientAvailable() types.CheckResult {
	cmd := exec.Command("which", "smbclient")
	err := cmd.Run()

	if err != nil {
		return types.CheckResult{
			ID:     "smbclient-command",
			Status: "warning",
		}
	}

	return types.CheckResult{
		ID:     "smbclient-command",
		Status: "pass",
	}
}

func (s *SystemService) checkSmbConfExists() types.CheckResult {
	smbConfPath := config.AppConfig.Samba.ConfigPath
	if smbConfPath == "" {
		smbConfPath = "/etc/samba/smb.conf"
	}

	info, err := os.Stat(smbConfPath)
	if err != nil {
		return types.CheckResult{
			ID:     "smb-conf-file",
			Status: "fail",
		}
	}

	// Check if readable
	if info.Mode().Perm()&0400 == 0 {
		return types.CheckResult{
			ID:     "smb-conf-file",
			Status: "warning",
		}
	}

	return types.CheckResult{
		ID:     "smb-conf-file",
		Status: "pass",
	}
}

func (s *SystemService) checkNssExtrausers() types.CheckResult {
	// Check if libnss-extrausers package is installed (Debian/Ubuntu)
	cmd := exec.Command("dpkg", "-l", "libnss-extrausers")
	if err := cmd.Run(); err == nil {
		return types.CheckResult{
			ID:     "libnss-extrausers",
			Status: "pass",
		}
	}

	// Fallback: Try to find libnss_extrausers.so files
	paths := []string{
		"/lib/x86_64-linux-gnu/libnss_extrausers.so.2",
		"/usr/lib/x86_64-linux-gnu/libnss_extrausers.so.2",
		"/lib64/libnss_extrausers.so.2",
		"/usr/lib64/libnss_extrausers.so.2",
		"/lib/libnss_extrausers.so.2",
		"/usr/lib/libnss_extrausers.so.2",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return types.CheckResult{
				ID:     "libnss-extrausers",
				Status: "pass",
			}
		}
	}

	return types.CheckResult{
		ID:     "libnss-extrausers",
		Status: "fail",
	}
}

func (s *SystemService) checkExtrausersDirectory() types.CheckResult {
	basePath := "/var/lib/extrausers"
	requiredFiles := []string{"passwd", "group", "shadow"}

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return types.CheckResult{
			ID:     "extrausers-directory",
			Status: "fail",
		}
	}

	var missing []string
	for _, file := range requiredFiles {
		if _, err := os.Stat(basePath + "/" + file); os.IsNotExist(err) {
			missing = append(missing, file)
		}
	}

	if len(missing) > 0 {
		return types.CheckResult{
			ID:     "extrausers-directory",
			Status: "fail",
		}
	}

	return types.CheckResult{
		ID:     "extrausers-directory",
		Status: "pass",
	}
}

func (s *SystemService) checkNsswitchConf() types.CheckResult {
	nsswitchPath := "/etc/nsswitch.conf"
	file, err := os.Open(nsswitchPath)
	if err != nil {
		return types.CheckResult{
			ID:     "nsswitch-conf",
			Status: "fail",
		}
	}
	defer file.Close()

	var passwdLine, groupLine string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "passwd:") {
			passwdLine = line
		}
		if strings.HasPrefix(line, "group:") {
			groupLine = line
		}
	}

	hasExtrausersPasswd := strings.Contains(passwdLine, "extrausers")
	hasExtrausersGroup := strings.Contains(groupLine, "extrausers")

	if !hasExtrausersPasswd || !hasExtrausersGroup {
		return types.CheckResult{
			ID:     "nsswitch-conf",
			Status: "fail",
		}
	}

	return types.CheckResult{
		ID:     "nsswitch-conf",
		Status: "pass",
	}
}

func (s *SystemService) checkSambaService() types.CheckResult {
	// Try systemctl first (systemd)
	cmd := exec.Command("systemctl", "is-active", "smbd")
	if err := cmd.Run(); err == nil {
		return types.CheckResult{
			ID:     "samba-service",
			Status: "pass",
		}
	}

	// Try service command (SysV init)
	cmd = exec.Command("service", "smbd", "status")
	if err := cmd.Run(); err == nil {
		return types.CheckResult{
			ID:     "samba-service",
			Status: "pass",
		}
	}

	return types.CheckResult{
		ID:     "samba-service",
		Status: "warning",
	}
}

func (s *SystemService) checkHomeDirectory() types.CheckResult {
	homeDir := config.AppConfig.HomeDir
	if homeDir == "" {
		homeDir = "/home/samba"
	}

	info, err := os.Stat(homeDir)
	if err != nil {
		return types.CheckResult{
			ID:     "home-directory",
			Status: "fail",
		}
	}

	if !info.IsDir() {
		return types.CheckResult{
			ID:     "home-directory",
			Status: "fail",
		}
	}

	return types.CheckResult{
		ID:     "home-directory",
		Status: "pass",
	}
}

func (s *SystemService) checkSambaConfigValid() types.CheckResult {
	cmd := exec.Command("testparm", "-s", config.AppConfig.Samba.ConfigPath)
	if err := cmd.Run(); err != nil {
		return types.CheckResult{
			ID:     "samba-config-valid",
			Status: "fail",
		}
	}

	return types.CheckResult{
		ID:     "samba-config-valid",
		Status: "pass",
	}
}

func (s *SystemService) checkPassdbBackend() types.CheckResult {
	configService := NewConfigService()
	cfg, err := configService.GetSambaConfig()
	if err != nil {
		return types.CheckResult{
			ID:     "passdb-backend",
			Status: "fail",
		}
	}

	backend := strings.TrimSpace(strings.ToLower(cfg.Global.PassdbBackend))

	// The only correct value for this application
	if backend == "tdbsam" {
		return types.CheckResult{
			ID:     "passdb-backend",
			Status: "pass",
		}
	}

	return types.CheckResult{
		ID:     "passdb-backend",
		Status: "fail",
	}
}

func (s *SystemService) checkSecurityMode() types.CheckResult {
	configService := NewConfigService()
	cfg, err := configService.GetSambaConfig()
	if err != nil {
		return types.CheckResult{
			ID:     "security-mode",
			Status: "fail",
		}
	}

	security := strings.TrimSpace(strings.ToLower(cfg.Global.Security))

	// The only correct value for this application
	if security == "user" {
		return types.CheckResult{
			ID:     "security-mode",
			Status: "pass",
		}
	}

	return types.CheckResult{
		ID:     "security-mode",
		Status: "fail",
	}
}

func (s *SystemService) checkForceUser() types.CheckResult {
	configService := NewConfigService()
	cfg, err := configService.GetSambaConfig()
	if err != nil {
		return types.CheckResult{
			ID:     "force-user",
			Status: "fail",
		}
	}

	forceUser := strings.TrimSpace(strings.ToLower(cfg.Homes.ForceUser))

	// The only correct value for this application
	if forceUser == "root" {
		return types.CheckResult{
			ID:     "force-user",
			Status: "pass",
		}
	}

	return types.CheckResult{
		ID:     "force-user",
		Status: "fail",
	}
}

func (s *SystemService) checkForceGroup() types.CheckResult {
	configService := NewConfigService()
	cfg, err := configService.GetSambaConfig()
	if err != nil {
		return types.CheckResult{
			ID:     "force-group",
			Status: "fail",
		}
	}

	forceGroup := strings.TrimSpace(strings.ToLower(cfg.Homes.ForceGroup))

	// The only correct value for this application
	if forceGroup == "root" {
		return types.CheckResult{
			ID:     "force-group",
			Status: "pass",
		}
	}

	return types.CheckResult{
		ID:     "force-group",
		Status: "fail",
	}
}

func (s *SystemService) checkRunningAsRoot() types.CheckResult {
	// Check if running as root by checking effective user ID
	cmd := exec.Command("id", "-u")
	output, err := cmd.Output()
	if err != nil {
		return types.CheckResult{
			ID:     "running-as-root",
			Status: "fail",
		}
	}

	uid := strings.TrimSpace(string(output))
	if uid == "0" {
		return types.CheckResult{
			ID:     "running-as-root",
			Status: "pass",
		}
	}

	return types.CheckResult{
		ID:     "running-as-root",
		Status: "fail",
	}
}

func (s *SystemService) checkAccessBasedShareEnum() types.CheckResult {
	configService := NewConfigService()
	cfg, err := configService.GetSambaConfig()
	if err != nil {
		return types.CheckResult{
			ID:     "access-based-share-enum",
			Status: "fail",
		}
	}

	accessBasedShareEnum := strings.TrimSpace(strings.ToLower(cfg.Global.AccessBasedShareEnum))

	// Must be "yes"
	if accessBasedShareEnum == "yes" {
		return types.CheckResult{
			ID:     "access-based-share-enum",
			Status: "pass",
		}
	}

	return types.CheckResult{
		ID:     "access-based-share-enum",
		Status: "fail",
	}
}

// GetSambaStatus gets the current Samba status using smbstatus command
func (s *SystemService) GetSambaStatus() (*types.SambaStatusResponse, error) {
	cmd := exec.Command("smbstatus")
	output, _ := cmd.CombinedOutput()

	// smbstatus might return non-zero even when it works (e.g., no connections)
	// So we just return the output regardless of error
	return &types.SambaStatusResponse{
		RawOutput: string(output),
	}, nil
}
