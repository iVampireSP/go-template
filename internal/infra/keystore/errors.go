package keystore

import "fmt"

// KeyNotFoundError 密钥未找到错误
type KeyNotFoundError struct {
	Name string
	Type string
}

func (e *KeyNotFoundError) Error() string {
	return fmt.Sprintf("%s key '%s' not found", e.Type, e.Name)
}
