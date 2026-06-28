package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// TestPrintFieldEnum_OptionsWithDesc 验证 options 路径：含 description 的值渲染为
// "value (desc)"，无 description 的只列值。
func TestPrintFieldEnum_OptionsWithDesc(t *testing.T) {
	var buf bytes.Buffer
	printFieldEnum(&buf, map[string]interface{}{
		"options": []interface{}{
			map[string]interface{}{"value": "a", "description": "选项A"},
			map[string]interface{}{"value": "b", "description": ""},
		},
	}, "  ")
	got := buf.String()
	if !strings.Contains(got, "enum: a (选项A), b") {
		t.Errorf("options 渲染不符（b 无描述只列值），得到: %q", got)
	}
}

// TestPrintFieldEnum_EnumFallback 验证无 options 但有 enum 时回退到纯值列表。
func TestPrintFieldEnum_EnumFallback(t *testing.T) {
	var buf bytes.Buffer
	printFieldEnum(&buf, map[string]interface{}{
		"enum": []interface{}{"x", "y"},
	}, "  ")
	if !strings.Contains(buf.String(), "enum: x, y") {
		t.Errorf("enum 回退应输出 'enum: x, y'，得到: %q", buf.String())
	}
}

// TestPrintFieldEnum_NumericValues 验证数字型 enum/options.value（如整数型 type 枚举）
// 不被静默丢弃：JSON 反序列化后是 float64，应渲染为整数串而非带 .0 尾巴。
func TestPrintFieldEnum_NumericValues(t *testing.T) {
	var buf bytes.Buffer
	printFieldEnum(&buf, map[string]interface{}{
		"enum": []interface{}{float64(0), float64(1), float64(2)},
	}, "  ")
	if !strings.Contains(buf.String(), "enum: 0, 1, 2") {
		t.Errorf("数字型 enum 应渲染为 '0, 1, 2'，得到: %q", buf.String())
	}

	var buf2 bytes.Buffer
	printFieldEnum(&buf2, map[string]interface{}{
		"options": []interface{}{
			map[string]interface{}{"value": float64(6), "description": "流程图"},
			map[string]interface{}{"value": float64(2), "description": "时序图"},
		},
	}, "  ")
	if !strings.Contains(buf2.String(), "enum: 6 (流程图), 2 (时序图)") {
		t.Errorf("数字型 options.value 应渲染为 '6 (流程图), 2 (时序图)'，得到: %q", buf2.String())
	}
}

// TestScalarToString 覆盖标量转字符串：整数 float64 去 .0、字符串原样、布尔、非标量空串。
func TestScalarToString(t *testing.T) {
	cases := []struct {
		in   interface{}
		want string
	}{
		{"abc", "abc"},
		{float64(42), "42"},
		{float64(3.14), "3.14"},
		{true, "true"},
		{[]interface{}{1, 2}, ""}, // 非标量
		{nil, ""},
	}
	for _, c := range cases {
		if got := scalarToString(c.in); got != c.want {
			t.Errorf("scalarToString(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestPrintFieldEnum_None 验证无枚举字段不输出任何内容。
func TestPrintFieldEnum_None(t *testing.T) {
	var buf bytes.Buffer
	printFieldEnum(&buf, map[string]interface{}{"type": "string"}, "  ")
	if buf.String() != "" {
		t.Errorf("无枚举应不输出，得到: %q", buf.String())
	}
}

// TestSchemaEnumRendering 端到端验证：approval.tasks.transfer 的 user_id_type 参数
// 含 options 枚举（来自飞书归一化端点），pretty 输出应渲染 enum 行 + 枚举描述。
func TestSchemaEnumRendering(t *testing.T) {
	var buf bytes.Buffer
	if err := runSchema(&buf, "approval.tasks.transfer", "pretty"); err != nil {
		t.Fatalf("runSchema err = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "enum:") {
		t.Errorf("pretty 输出应含 enum 行（来自 options），实际:\n%s", out)
	}
	if !strings.Contains(out, "以user_id来识别用户") {
		t.Errorf("pretty 输出应含枚举描述「以user_id来识别用户」，实际:\n%s", out)
	}
}
