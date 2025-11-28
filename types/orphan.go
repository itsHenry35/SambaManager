package types

// OrphanedDirectory represents a home directory without a corresponding user
type OrphanedDirectory struct {
	Name string `json:"name"` // Directory name (username)
	Path string `json:"path"` // Full path to directory
	Size int64  `json:"size"` // Total size in bytes
}
