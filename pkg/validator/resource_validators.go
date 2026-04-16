package validator

import (
	"regexp"

	"github.com/gookit/validate"
)

var (
	// CPU 验证正则：只支持 m (毫核) 单位，如 100m, 1000m
	cpuRegex = regexp.MustCompile(`^[1-9]\d*m$`)

	// Memory 验证正则：只支持 Mi (兆字节) 单位，如 128Mi, 1024Mi
	memoryRegex = regexp.MustCompile(`^[1-9]\d*Mi$`)
)

func init() {
	// 注册 CPU 资源验证器
	// CPU 只支持 m 单位（如 100m, 500m, 1000m）
	validate.AddValidator("resource_cpu", func(val any) bool {
		str, ok := val.(string)
		if !ok || str == "" {
			return true // 空值由 required 处理
		}
		return cpuRegex.MatchString(str)
	})

	// 注册 Memory 资源验证器
	// 内存只支持 Mi 单位（如 128Mi, 512Mi, 1024Mi）
	validate.AddValidator("resource_memory", func(val any) bool {
		str, ok := val.(string)
		if !ok || str == "" {
			return true // 空值由 required 处理
		}
		return memoryRegex.MatchString(str)
	})
}
