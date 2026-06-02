// Package strutil 提供字符串工具函数。
package strutil

// SplitTags 将逗号分隔的标签字符串拆分为切片。
func SplitTags(raw string) []string {
	var parts []string
	current := ""
	for _, ch := range raw {
		if ch == ',' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// Trim 去除字符串首尾的空格和制表符。
func Trim(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

// NullableStr 安全解引用 *string，nil 返回空串。
func NullableStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
