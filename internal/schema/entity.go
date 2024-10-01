package schema

import (
	"strconv"
)

type EntityId uint

//type EntityId int64

func (i EntityId) String() string {
	return strconv.FormatUint(uint64(i), 10)
	//return strconv.FormatInt(int64(i), 10)
}
func (i EntityId) Uint() uint {
	return uint(i)
}
