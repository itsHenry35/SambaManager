package types

// SystemCheckResponse contains system environment check results
type SystemCheckResponse struct {
	Checks []CheckResult `json:"checks"`
}

// CheckResult represents a single check result
type CheckResult struct {
	ID     string `json:"id"`     // Unique identifier for i18n (e.g., "samba-installed", "home-directory")
	Status string `json:"status"` // "pass", "fail", "warning"
}

// SambaGlobalConfig represents Samba [global] section settings
type SambaGlobalConfig struct {
	Workgroup            string `json:"workgroup"`
	ServerString         string `json:"server_string"`
	Security             string `json:"security"`
	PassdbBackend        string `json:"passdb_backend"`
	MapToGuest           string `json:"map_to_guest"`
	AccessBasedShareEnum string `json:"access_based_share_enum"`
}

// SambaHomesConfig represents Samba [homes] section settings
type SambaHomesConfig struct {
	Comment       string `json:"comment"`
	Browseable    string `json:"browseable"`
	Writable      string `json:"writable"`
	ValidUsers    string `json:"valid_users"`
	ForceUser     string `json:"force_user"`
	ForceGroup    string `json:"force_group"`
	CreateMask    string `json:"create_mask"`
	DirectoryMask string `json:"directory_mask"`
}

// SambaConfigResponse contains both global and homes configuration
type SambaConfigResponse struct {
	Global SambaGlobalConfig `json:"global"`
	Homes  SambaHomesConfig  `json:"homes"`
}

// UpdateSambaConfigRequest contains configuration updates
type UpdateSambaConfigRequest struct {
	Global *SambaGlobalConfig `json:"global,omitempty"`
	Homes  *SambaHomesConfig  `json:"homes,omitempty"`
}

// SambaConfigFileResponse contains the raw smb.conf file content
type SambaConfigFileResponse struct {
	Content string `json:"content"` // Raw smb.conf content
	Path    string `json:"path"`    // Path to smb.conf
}

// UpdateSambaConfigFileRequest contains raw smb.conf update
type UpdateSambaConfigFileRequest struct {
	Content string `json:"content"` // New smb.conf content
}

// SambaStatusResponse contains smbstatus output
type SambaStatusResponse struct {
	RawOutput string `json:"raw_output"` // Raw smbstatus command output
}
