package converter

import (
	"strings"
	"testing"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
)

// TestUnescapeMarkdownText 测试 CommonMark 反斜杠转义的去除。
// goldmark 的 Segment.Value 返回源文件原始字节，不处理转义序列，
// 需要 unescapeMarkdownText 将 "1\." 还原为 "1."、"\[1\]" 还原为 "[1]"。
func TestUnescapeMarkdownText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// 数字+点号转义（Defuddle 最常见的产物）
		{"escaped_dot_1", `1\. 反逆向`, "1. 反逆向"},
		{"escaped_dot_2", `2\. 反不合规`, "2. 反不合规"},
		{"escaped_dot_in_text", `正文 3\. 内容`, "正文 3. 内容"},

		// 方括号转义（脚注引用）
		{"escaped_brackets", `\[1\]`, "[1]"},
		{"escaped_brackets_in_text", `参见脚注 \[2\]`, "参见脚注 [2]"},

		// 下划线转义（变量名、字段名）
		{"escaped_underscore", `prompt\_len`, "prompt_len"},
		{"escaped_underscore_2", `needs\_web\_search`, "needs_web_search"},
		{"escaped_underscore_3", `edit\_file`, "edit_file"},
		{"escaped_underscore_4", `has\_attachments`, "has_attachments"},

		// 其他 ASCII 标点转义
		{"escaped_asterisk", `\*粗体\*`, "*粗体*"},
		{"escaped_hash", `\# 不是标题`, "# 不是标题"},
		{"escaped_backtick", "\\`code\\`", "`code`"},
		{"escaped_tilde", `\~删除线\~`, "~删除线~"},
		{"escaped_pipe", `\|表格分隔\|`, "|表格分隔|"},
		{"escaped_backslash", `\\`, `\`},

		// 不应处理的情况：反斜杠后跟非 ASCII 标点
		{"no_escape_letter", `\n 不处理`, `\n 不处理`},
		{"no_escape_space", `\ 后跟空格`, `\ 后跟空格`},
		{"no_escape_chinese", `\中文`, `\中文`},
		{"trailing_backslash", `末尾\`, `末尾\`},
		{"plain_text", "普通文本无转义", "普通文本无转义"},
		{"empty_string", "", ""},

		// 混合场景
		{"mixed", `### 1\. 标题 \[1\]`, "### 1. 标题 [1]"},
		{"mixed_underscores", `file\_name / tos\_key / file\_source`, "file_name / tos_key / file_source"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unescapeMarkdownText(tt.input)
			if got != tt.want {
				t.Errorf("unescapeMarkdownText(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestUnescapeMarkdownBytes 测试 []byte 版本。
func TestUnescapeMarkdownBytes(t *testing.T) {
	input := []byte(`prompt\_len`)
	got := unescapeMarkdownBytes(input)
	want := []byte("prompt_len")
	if string(got) != string(want) {
		t.Errorf("unescapeMarkdownBytes(%q) = %q, want %q", input, got, want)
	}
}

// TestConvert_EscapedHeading 端到端测试：Markdown 标题含转义字符 → Convert → 飞书块文本不含反斜杠。
// 复现场景：Defuddle 输出 "### 1\. 反逆向" 导入飞书后显示为 "1\. 反逆向"。
func TestConvert_EscapedHeading(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		want     string
	}{
		{
			"heading_with_escaped_dot",
			"### 1\\. 反逆向 / 反自动化",
			"1. 反逆向 / 反自动化",
		},
		{
			"heading_with_escaped_dot_2",
			"### 2\\. 反不合规地区使用",
			"2. 反不合规地区使用",
		},
		{
			"heading_without_escape",
			"### 普通标题",
			"普通标题",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMarkdownToBlock([]byte(tt.markdown), ConvertOptions{}, "")
			blocks, err := converter.Convert()
			if err != nil {
				t.Fatalf("Convert() 返回错误: %v", err)
			}
			if len(blocks) == 0 {
				t.Fatal("blocks 为空")
			}

			got := collectBlockText(blocks[0])
			if got != tt.want {
				t.Errorf("标题文本 = %q, 期望 %q", got, tt.want)
			}
		})
	}
}

// TestConvert_EscapedParagraph 端到端测试：正文段落含转义字符。
func TestConvert_EscapedParagraph(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		want     string
	}{
		{
			"paragraph_with_escaped_brackets",
			"参见脚注 \\[1\\]",
			"参见脚注 [1]",
		},
		{
			"paragraph_with_escaped_dot",
			"正文 3\\. 某某某",
			"正文 3. 某某某",
		},
		{
			"paragraph_with_escaped_underscore",
			"字段名 prompt\\_len 是整数类型",
			"字段名 prompt_len 是整数类型",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMarkdownToBlock([]byte(tt.markdown), ConvertOptions{}, "")
			blocks, err := converter.Convert()
			if err != nil {
				t.Fatalf("Convert() 返回错误: %v", err)
			}
			if len(blocks) == 0 {
				t.Fatal("blocks 为空")
			}

			got := collectBlockText(blocks[0])
			if got != tt.want {
				t.Errorf("段落文本 = %q, 期望 %q", got, tt.want)
			}
		})
	}
}

// TestConvert_EscapedListItem 端到端测试：列表项含转义字符。
func TestConvert_EscapedListItem(t *testing.T) {
	markdown := "- 选项 A\\: Google Voice\n- 选项 B\\: 实体 SIM"

	converter := NewMarkdownToBlock([]byte(markdown), ConvertOptions{}, "")
	blocks, err := converter.Convert()
	if err != nil {
		t.Fatalf("Convert() 返回错误: %v", err)
	}

	for _, b := range blocks {
		text := collectBlockText(b)
		if strings.Contains(text, `\:`) {
			t.Errorf("列表项文本含未转义的反斜杠: %q", text)
		}
	}
}

// TestConvert_EscapedTableCell 端到端测试：表格单元格含转义下划线。
// 复现场景：表格中 "prompt\_len" 导入飞书后显示为 "prompt\_len" 而非 "prompt_len"。
func TestConvert_EscapedTableCell(t *testing.T) {
	markdown := `| 字段 | 类型 | 说明 |
| --- | --- | --- |
| prompt\_len | int | prompt 字符长度 |
| needs\_web\_search | bool | 是否需要联网搜索 |
| edit\_file | string | 编辑文件工具 |`

	converter := NewMarkdownToBlock([]byte(markdown), ConvertOptions{}, "")
	result, err := converter.ConvertWithTableData()
	if err != nil {
		t.Fatalf("ConvertWithTableData() 返回错误: %v", err)
	}

	if len(result.TableDatas) == 0 {
		t.Fatal("没有表格数据")
	}

	td := result.TableDatas[0]

	// 检查所有单元格的纯文本内容中不含反斜杠转义
	for i, text := range td.CellContents {
		if strings.Contains(text, `\_`) {
			t.Errorf("表格单元格 [%d] 含未转义的反斜杠: %q", i, text)
		}
	}

	// 检查所有单元格的富文本元素中不含反斜杠转义
	for i, elements := range td.CellElements {
		for _, elem := range elements {
			if elem.TextRun != nil && elem.TextRun.Content != nil {
				text := *elem.TextRun.Content
				if strings.Contains(text, `\_`) {
					t.Errorf("表格富文本 [%d] 含未转义的反斜杠: %q", i, text)
				}
			}
		}
	}

	// 验证具体字段名出现在单元格纯文本中
	wantFields := []string{"prompt_len", "needs_web_search", "edit_file"}
	for _, want := range wantFields {
		found := false
		for _, text := range td.CellContents {
			if strings.Contains(text, want) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("表格中未找到正确的字段名 %q（可能仍含反斜杠转义）", want)
		}
	}
}

// collectBlockText 从 larkdocx.Block 中提取所有 TextElement 的拼接文本。
func collectBlockText(block *larkdocx.Block) string {
	var elements []*larkdocx.TextElement

	switch {
	case block.Heading1 != nil:
		elements = block.Heading1.Elements
	case block.Heading2 != nil:
		elements = block.Heading2.Elements
	case block.Heading3 != nil:
		elements = block.Heading3.Elements
	case block.Heading4 != nil:
		elements = block.Heading4.Elements
	case block.Heading5 != nil:
		elements = block.Heading5.Elements
	case block.Heading6 != nil:
		elements = block.Heading6.Elements
	case block.Text != nil:
		elements = block.Text.Elements
	case block.Bullet != nil:
		elements = block.Bullet.Elements
	case block.Ordered != nil:
		elements = block.Ordered.Elements
	default:
		return ""
	}

	var buf strings.Builder
	for _, e := range elements {
		if e != nil && e.TextRun != nil && e.TextRun.Content != nil {
			buf.WriteString(*e.TextRun.Content)
		}
	}
	return buf.String()
}
