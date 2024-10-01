package page

const DefaultPageSize = 20

func Offset(page int) int {
	if page <= 0 {
		return 0
	}
	return (page - 1) * DefaultPageSize
}

func OffsetCustom(page int, pageSize int) int {
	return (page - 1) * pageSize
}
