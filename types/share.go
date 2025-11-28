package types

// Share represents a Samba share
type Share struct {
	Name       string   `json:"name"`                         // Optional custom name (alphanumeric only, no symbols)
	Owner      string   `json:"owner" binding:"required"`
	SharedWith []string `json:"shared_with" binding:"required"`
	ReadOnly   bool     `json:"read_only"`
	Comment    string   `json:"comment"`
	SubPath    string   `json:"sub_path"` // Optional subdirectory path relative to owner's home
}

// ShareResponse represents share information returned to client
type ShareResponse struct {
	ID         string   `json:"id"`          // Share ID (e.g., "alice-share-ProjectFiles" or "alice-share-20251128184826")
	Owner      string   `json:"owner"`       // Owner username
	Path       string   `json:"path"`        // Path to shared directory
	SharedWith []string `json:"shared_with"` // List of usernames with access
	ReadOnly   bool     `json:"read_only"`   // Whether the share is read-only
	Comment    string   `json:"comment"`     // Share description
	SubPath    string   `json:"sub_path"`    // Subdirectory path relative to owner's home
}

// CreateShareRequest represents a request to create a new share
type CreateShareRequest struct {
	Name       string   `json:"name"`                         // Optional custom name (alphanumeric only, no symbols)
	Owner      string   `json:"owner" binding:"required"`
	SharedWith []string `json:"shared_with" binding:"required"`
	ReadOnly   bool     `json:"read_only"`
	Comment    string   `json:"comment"`
	SubPath    string   `json:"sub_path"` // Optional subdirectory path
}

// UpdateShareRequest represents a request to update share information
type UpdateShareRequest struct {
	SharedWith []string `json:"shared_with" binding:"required"`
	ReadOnly   bool     `json:"read_only"`
	Comment    string   `json:"comment"`
	SubPath    string   `json:"sub_path"` // Optional subdirectory path
}

// CreateMyShareRequest represents a request for user to create their own share (no owner field needed)
type CreateMyShareRequest struct {
	Name       string   `json:"name"`                         // Optional custom name (alphanumeric only, no symbols)
	SharedWith []string `json:"shared_with" binding:"required"`
	ReadOnly   bool     `json:"read_only"`
	Comment    string   `json:"comment"`
	SubPath    string   `json:"sub_path"` // Optional subdirectory path
}
