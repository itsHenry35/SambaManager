package types

// User represents a system user
type User struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserResponse represents user information returned to client
type UserResponse struct {
	Username string `json:"username"`
	HomeDir  string `json:"home_dir"`
}

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=3"`
}

// UpdateUserRequest represents a request to update user information
type UpdateUserRequest struct {
	Password string `json:"password" binding:"required,min=3"`
}

// ChangePasswordRequest represents a request to change user password
type ChangePasswordRequest struct {
	Password string `json:"password" binding:"required,min=3"`
}

// UserRole represents user role type
type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

// ChangeOwnPasswordRequest represents a request for user to change their own password
type ChangeOwnPasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=3"`
}

// DeleteUserRequest represents options when deleting a user
type DeleteUserRequest struct {
	DeleteHomeDir bool `json:"delete_home_dir"` // Whether to delete the user's home directory
}
