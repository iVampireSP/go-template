package paginator

// LengthAwarePaginator 对标 Laravel Illuminate\Pagination\LengthAwarePaginator。
// 泛型参数 T 通常是 Resource 类型。
type LengthAwarePaginator[T any] struct {
	Items []T  `json:"items"`
	Meta  Meta `json:"meta"`
}

// New 创建分页结果。
//
//	items: 当前页的数据切片
//	total: 总记录数
//	perPage: 每页数量
//	page: 当前页码（从 1 开始）
//
// nil items 自动转为空切片，确保 JSON 序列化为 [] 而非 null。
func New[T any](items []T, total, perPage, page int) *LengthAwarePaginator[T] {
	if items == nil {
		items = []T{}
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	totalPages := 0
	if perPage > 0 && total > 0 {
		totalPages = (total + perPage - 1) / perPage
	}

	return &LengthAwarePaginator[T]{
		Items: items,
		Meta: Meta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}

// NewFromSlice 从完整切片创建分页结果，自动计算 offset 并切分。
// 对标 Laravel: new LengthAwarePaginator($items->slice($offset, $perPage), $items->count(), $perPage, $page)
func NewFromSlice[T any](items []T, perPage, page int) *LengthAwarePaginator[T] {
	total := len(items)
	offset := Offset(page, perPage)
	if offset > total {
		offset = total
	}
	end := offset + perPage
	if end > total {
		end = total
	}
	return New(items[offset:end], total, perPage, page)
}
