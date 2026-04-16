// Package paginator 提供 Laravel 风格的分页器。
//
// 用法：
//
//	items := fetchItems()
//	paged := paginator.New(items, total, perPage, page)
//	// paged.Items, paged.Meta 可直接序列化
package paginator

// Meta 分页元数据
type Meta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// Offset 计算偏移量
func Offset(page, perPage int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * perPage
}
