package safetpl

import (
	"regexp"
)

// RenderTemplate 渲染模板并替换未赋值的变量
func RenderTemplate(template string, data map[string]string) string {
	re := regexp.MustCompile(`\{(\w+)\}`)
	return re.ReplaceAllStringFunc(template, func(s string) string {
		key := re.FindStringSubmatch(s)[1]
		value, exists := data[key]
		if !exists {
			return "[not set]"
		}
		return value
	})
}

//func main() {
//	// 模板字符串
//	tpl := "Hello, {name}! Your age is {age}. Welcome to {city}."
//
//	// 替换的数据
//	data := map[string]string{
//		"name": "Twilight",
//		"age":  "28",
//		// "city" 没有赋值
//	}
//
//	result := RenderTemplate(tpl, data)
//	fmt.Println(result)
//}
