package converter

import (
	"regexp"
	"strings"
)

// HTMLTag 表示解析后的 HTML 标签
type HTMLTag struct {
	Name        string            // 标签名（如 "mention-user", "grid"）
	Attrs       map[string]string // 属性（如 {"id": "ou_xxx"}）
	Content     string            // 标签内容（非自闭合标签的内容）
	SelfClosing bool              // 是否自闭合
	Raw         string            // 原始 HTML 字符串
}

// 正则：提取标签名（支持带连字符的自定义标签名，< 和标签名之间不允许有空格）
var reTagName = regexp.MustCompile(`^<([a-zA-Z][a-zA-Z0-9\-]*)`)

// 正则：提取属性 key="value" 或 key='value'
var reTagAttr = regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9\-_]*)\s*=\s*(?:"([^"]*)"|'([^']*)')`)

// ParseHTMLTag 从原始 HTML 字符串解析标签
// 支持：<tag attr="val"/> <tag attr="val">content</tag> <tag attr='val'>
// 返回 nil 如果不是有效的 HTML 标签
func ParseHTMLTag(raw string) *HTMLTag {
	trimmed := strings.TrimSpace(raw)
	if len(trimmed) < 3 || trimmed[0] != '<' {
		return nil
	}

	// 忽略闭合标签 </xxx>
	if strings.HasPrefix(trimmed, "</") {
		return nil
	}

	// 提取标签名
	nameMatch := reTagName.FindStringSubmatch(trimmed)
	if nameMatch == nil {
		return nil
	}
	tagName := strings.ToLower(nameMatch[1])

	// 提取属性
	attrs := make(map[string]string)
	attrMatches := reTagAttr.FindAllStringSubmatch(trimmed, -1)
	for _, m := range attrMatches {
		key := strings.ToLower(m[1])
		// m[2] 是双引号内的值，m[3] 是单引号内的值
		val := m[2]
		if val == "" {
			val = m[3]
		}
		attrs[key] = val
	}

	tag := &HTMLTag{
		Name:  tagName,
		Attrs: attrs,
		Raw:   raw,
	}

	// 判断自闭合：以 /> 结尾
	if strings.HasSuffix(strings.TrimSpace(trimmed), "/>") {
		tag.SelfClosing = true
		return tag
	}

	// 非自闭合标签：提取 >content</tag> 中的 content
	closingTag := "</" + tagName + ">"
	// 找到第一个 >（非自闭合的 >）
	gtIdx := strings.Index(trimmed, ">")
	if gtIdx < 0 {
		return tag
	}
	// 检查 > 前面不是 /
	if gtIdx > 0 && trimmed[gtIdx-1] == '/' {
		tag.SelfClosing = true
		return tag
	}

	afterGt := trimmed[gtIdx+1:]
	closingIdx := strings.LastIndex(strings.ToLower(afterGt), closingTag)
	if closingIdx >= 0 {
		tag.Content = afterGt[:closingIdx]
	}

	return tag
}

// IsHTMLTag 检查原始字符串是否是指定标签（开始标签）
func IsHTMLTag(raw, tagName string) bool {
	lower := strings.ToLower(strings.TrimSpace(raw))
	return strings.HasPrefix(lower, "<"+tagName)
}

// IsHTMLClosingTag 检查是否是闭合标签
func IsHTMLClosingTag(raw, tagName string) bool {
	lower := strings.ToLower(strings.TrimSpace(raw))
	return lower == "</"+tagName+">"
}

// reColumnTag 匹配 <column>...</column> 块（非贪婪，允许中间任意字符含换行）
var reColumnTag = regexp.MustCompile(`(?si)<column\s*>(.*?)</column\s*>`)

// ParseGridColumns 从 <grid> 标签的 Content 中提取各 <column>...</column> 的内容
func ParseGridColumns(content string) []string {
	matches := reColumnTag.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}
	columns := make([]string, 0, len(matches))
	for _, m := range matches {
		columns = append(columns, strings.TrimSpace(m[1]))
	}
	return columns
}

// mapDocTypeToObjType 将文档类型字符串映射为飞书 ObjType 整数
// docx→22, doc→1, sheet→3, bitable→8, mindnote→11, wiki→16, file→12, slides→15
func mapDocTypeToObjType(docType string) int {
	switch strings.ToLower(docType) {
	case "doc":
		return 1
	case "sheet":
		return 3
	case "bitable":
		return 8
	case "mindnote":
		return 11
	case "file":
		return 12
	case "slides":
		return 15
	case "wiki":
		return 16
	case "docx":
		return 22
	default:
		return 22 // 默认 docx
	}
}

// mapObjTypeToDocType 将飞书 ObjType 整数映射为文档类型字符串
func mapObjTypeToDocType(objType *int) string {
	if objType == nil {
		return "docx"
	}
	switch *objType {
	case 1:
		return "doc"
	case 3:
		return "sheet"
	case 8:
		return "bitable"
	case 11:
		return "mindnote"
	case 12:
		return "file"
	case 15:
		return "slides"
	case 16:
		return "wiki"
	case 22:
		return "docx"
	default:
		return "docx"
	}
}
