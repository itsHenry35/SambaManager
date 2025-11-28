package types

// PaginationQuery represents pagination parameters
type PaginationQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`       // Page number (default: 1)
	PageSize int    `form:"page_size" binding:"omitempty,min=1"`  // Items per page (default: 20)
	Search   string `form:"search"`                                // Search query
}

// GetDefaults returns pagination query with default values
func (p *PaginationQuery) GetDefaults() PaginationQuery {
	result := *p
	if result.Page == 0 {
		result.Page = 1
	}
	if result.PageSize == 0 {
		result.PageSize = 20
	}
	return result
}

// GetOffset calculates the offset for database queries
func (p PaginationQuery) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}
