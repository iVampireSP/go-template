package schema

import (
	"strconv"
)

type EntityId uint

func (i EntityId) String() string {
	return strconv.FormatUint(uint64(i), 10)
}

func (i EntityId) Uint() uint {
	return uint(i)
}

func (i EntityId) Uint64() uint64 {
	return uint64(i)
}

func (i EntityId) Int64() int64 {
	return int64(i)
}
