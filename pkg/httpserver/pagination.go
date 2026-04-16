package httpserver

// PaginationParams 用于 Huma 输入的分页参数
// 可嵌入到 Input 结构体中，Huma 会自动解析 query 参数
type PaginationParams struct {
	Page    int `query:"page" default:"1" minimum:"1" doc:"页码"`
	PerPage int `query:"per_page" default:"20" minimum:"1" maximum:"200" doc:"每页数量"`
}

// Pagination 分页元数据
type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// PaginatedResult 分页结果
type PaginatedResult[T any] struct {
	Items      []T         `json:"items"`      // 数据列表
	Pagination *Pagination `json:"pagination"` // 分页信息
}

// NewPaginatedResult 创建分页结果
// 自动将 nil 切片转换为空切片，确保 JSON 序列化为 [] 而不是 null
func NewPaginatedResult[T any](items []T, total, page, perPage int) *PaginatedResult[T] {
	if items == nil {
		items = []T{}
	}
	return &PaginatedResult[T]{
		Items:      items,
		Pagination: NewPagination(page, perPage, total),
	}
}

// NewPagination 创建分页元数据
func NewPagination(page, perPage, total int) *Pagination {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + perPage - 1) / perPage
	}

	return &Pagination{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}
