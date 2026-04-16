package paginator

// Pager 用于 service 层 DB 查询的分页计算器。
// 替代已弃用的 pagination.Paginator。
type Pager struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// PageParams 分页参数（不带 HTTP 标签，可用于 internal 层）
type PageParams struct {
	Page     int
	PageSize int
}

// Pager 返回分页器
func (p PageParams) Pager() *Pager {
	return NewPager(p.Page, p.PageSize)
}

// NewPager 创建分页器。
// 自动修正无效参数：page < 1 → 1, pageSize 限制在 [1, 100]。
func NewPager(page, pageSize int) *Pager {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return &Pager{
		Page:     page,
		PageSize: pageSize,
	}
}

// Offset 计算偏移量
func (p *Pager) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Limit 返回限制数量
func (p *Pager) Limit() int {
	return p.PageSize
}

// SetTotal 设置总记录数并自动计算 TotalPages。
func (p *Pager) SetTotal(total int) {
	p.Total = total
	if p.PageSize > 0 && total > 0 {
		p.TotalPages = (total + p.PageSize - 1) / p.PageSize
	} else {
		p.TotalPages = 0
	}
}
