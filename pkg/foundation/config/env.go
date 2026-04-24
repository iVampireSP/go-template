package config

import (
	"os"
	"regexp"
	"strings"
)

var (
	// ${VAR:-default} 带默认值
	envWithDefault = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*):-([^}]*)\}`)
	// ${VAR} 必填
	envRequired = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)
)

// LoadDotEnv 加载 .env 文件到环境变量
// 可供外部工具（如迁移脚本）单独调用
func LoadDotEnv() {
	for _, file := range []string{".env", ".env.local"} {
		if content, err := os.ReadFile(file); err == nil {
			parseEnvFile(string(content))
		}
	}
}

// parseEnvFile 解析 .env 文件内容
func parseEnvFile(content string) {
	lines := strings.Split(content, "\n")
	for index := 0; index < len(lines); index++ {
		line := strings.TrimSpace(strings.TrimRight(lines[index], "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if idx := strings.Index(line, "="); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			rawValue := strings.TrimSpace(line[idx+1:])
			value, consumed := parseEnvValue(lines, index, rawValue)
			index = consumed

			// 只设置未定义的环境变量
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}
}

func parseEnvValue(lines []string, startIndex int, rawValue string) (string, int) {
	if rawValue == "" {
		return "", startIndex
	}
	if rawValue[0] != '"' && rawValue[0] != '\'' {
		return rawValue, startIndex
	}

	quote := rawValue[0]
	if value, complete := parseSingleLineQuotedEnvValue(rawValue, quote); complete {
		return value, startIndex
	}

	var builder strings.Builder
	builder.WriteString(rawValue[1:])

	for index := startIndex + 1; index < len(lines); index++ {
		nextLine := strings.TrimRight(lines[index], "\r")
		if builder.Len() > 0 {
			builder.WriteByte('\n')
		}
		if end, found := findQuotedEnvValueEnd(nextLine, quote, 0); found {
			builder.WriteString(nextLine[:end])
			return builder.String(), index
		}
		builder.WriteString(nextLine)
	}

	// 引号不完整时回退原值，保持兼容
	return rawValue, startIndex
}

func parseSingleLineQuotedEnvValue(rawValue string, quote byte) (string, bool) {
	if end, found := findQuotedEnvValueEnd(rawValue, quote, 1); found {
		return rawValue[1:end], true
	}
	if findNextUnescapedQuote(rawValue, quote, 1) != -1 {
		return rawValue, true
	}
	return "", false
}

func findQuotedEnvValueEnd(value string, quote byte, start int) (int, bool) {
	for index := start; index < len(value); {
		end := findNextUnescapedQuote(value, quote, index)
		if end == -1 {
			return -1, false
		}
		suffix := strings.TrimSpace(value[end+1:])
		if suffix == "" || strings.HasPrefix(suffix, "#") {
			return end, true
		}
		index = end + 1
	}
	return -1, false
}

func findNextUnescapedQuote(value string, quote byte, start int) int {
	for index := start; index < len(value); index++ {
		if value[index] != quote {
			continue
		}

		backslashCount := 0
		for cursor := index - 1; cursor >= 0 && value[cursor] == '\\'; cursor-- {
			backslashCount++
		}
		if backslashCount%2 == 0 {
			return index
		}
	}

	return -1
}

// substituteEnv 替换字符串中的环境变量
// 支持 ${VAR} 和 ${VAR:-default} 语法
func substituteEnv(content string) string {
	// 先处理带默认值的 ${VAR:-default}
	result := envWithDefault.ReplaceAllStringFunc(content, func(match string) string {
		groups := envWithDefault.FindStringSubmatch(match)
		if len(groups) == 3 {
			if value := os.Getenv(groups[1]); value != "" {
				return escapeYAMLValue(value)
			}
			return groups[2] // 返回默认值
		}
		return match
	})

	// 再处理必填的 ${VAR}
	result = envRequired.ReplaceAllStringFunc(result, func(match string) string {
		groups := envRequired.FindStringSubmatch(match)
		if len(groups) == 2 {
			return escapeYAMLValue(os.Getenv(groups[1]))
		}
		return match
	})

	return result
}

// escapeYAMLValue 处理需要转义的 YAML 值
// 对于包含换行符或特殊字符的值，使用双引号包裹并转义
func escapeYAMLValue(value string) string {
	if value == "" {
		return value
	}

	// 检查是否需要转义（包含换行符、冒号、特殊字符等）
	needsEscape := strings.ContainsAny(value, "\n\r\t:{}[]&*#?|><!%@`\"'\\")

	if !needsEscape {
		return value
	}

	// 使用双引号包裹，并转义内部的特殊字符
	var sb strings.Builder
	sb.WriteByte('"')
	for _, r := range value {
		switch r {
		case '"':
			sb.WriteString(`\"`)
		case '\\':
			sb.WriteString(`\\`)
		case '\n':
			sb.WriteString(`\n`)
		case '\r':
			sb.WriteString(`\r`)
		case '\t':
			sb.WriteString(`\t`)
		default:
			sb.WriteRune(r)
		}
	}
	sb.WriteByte('"')
	return sb.String()
}
