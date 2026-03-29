package converter

import (
	"testing"
)

func TestParseHTMLTag_SelfClosing(t *testing.T) {
	tag := ParseHTMLTag(`<mention-user id="ou_xxx"/>`)
	if tag == nil {
		t.Fatal("expected non-nil tag")
	}
	if tag.Name != "mention-user" {
		t.Errorf("expected name 'mention-user', got %q", tag.Name)
	}
	if tag.Attrs["id"] != "ou_xxx" {
		t.Errorf("expected attr id='ou_xxx', got %q", tag.Attrs["id"])
	}
	if !tag.SelfClosing {
		t.Error("expected self-closing")
	}
	if tag.Content != "" {
		t.Errorf("expected empty content, got %q", tag.Content)
	}
}

func TestParseHTMLTag_WithContent(t *testing.T) {
	tag := ParseHTMLTag(`<mention-doc token="abc" type="docx">文档标题</mention-doc>`)
	if tag == nil {
		t.Fatal("expected non-nil tag")
	}
	if tag.Name != "mention-doc" {
		t.Errorf("expected name 'mention-doc', got %q", tag.Name)
	}
	if tag.Attrs["token"] != "abc" {
		t.Errorf("expected attr token='abc', got %q", tag.Attrs["token"])
	}
	if tag.Attrs["type"] != "docx" {
		t.Errorf("expected attr type='docx', got %q", tag.Attrs["type"])
	}
	if tag.SelfClosing {
		t.Error("expected non self-closing")
	}
	if tag.Content != "文档标题" {
		t.Errorf("expected content '文档标题', got %q", tag.Content)
	}
}

func TestParseHTMLTag_SingleQuoteAttrs(t *testing.T) {
	tag := ParseHTMLTag(`<image token='file_xxx' width='800'/>`)
	if tag == nil {
		t.Fatal("expected non-nil tag")
	}
	if tag.Name != "image" {
		t.Errorf("expected name 'image', got %q", tag.Name)
	}
	if tag.Attrs["token"] != "file_xxx" {
		t.Errorf("expected attr token='file_xxx', got %q", tag.Attrs["token"])
	}
	if tag.Attrs["width"] != "800" {
		t.Errorf("expected attr width='800', got %q", tag.Attrs["width"])
	}
}

func TestParseHTMLTag_ClosingTag(t *testing.T) {
	tag := ParseHTMLTag(`</mention-doc>`)
	if tag != nil {
		t.Error("expected nil for closing tag")
	}
}

func TestParseHTMLTag_Invalid(t *testing.T) {
	tests := []string{
		"",
		"plain text",
		"< not a tag",
		"<>",
	}
	for _, s := range tests {
		tag := ParseHTMLTag(s)
		if tag != nil {
			t.Errorf("expected nil for %q, got %+v", s, tag)
		}
	}
}

func TestParseHTMLTag_ImageFullAttrs(t *testing.T) {
	tag := ParseHTMLTag(`<image token="img_v3_xxx" width="800" height="600" align="center"/>`)
	if tag == nil {
		t.Fatal("expected non-nil tag")
	}
	if tag.Attrs["token"] != "img_v3_xxx" {
		t.Errorf("token = %q", tag.Attrs["token"])
	}
	if tag.Attrs["width"] != "800" {
		t.Errorf("width = %q", tag.Attrs["width"])
	}
	if tag.Attrs["height"] != "600" {
		t.Errorf("height = %q", tag.Attrs["height"])
	}
	if tag.Attrs["align"] != "center" {
		t.Errorf("align = %q", tag.Attrs["align"])
	}
}

func TestIsHTMLTag(t *testing.T) {
	if !IsHTMLTag(`<mention-user id="ou_xxx"/>`, "mention-user") {
		t.Error("should match mention-user")
	}
	if IsHTMLTag(`<mention-doc token="abc">title</mention-doc>`, "mention-user") {
		t.Error("should not match")
	}
	if !IsHTMLTag(`  <IMAGE token="abc"/>  `, "image") {
		t.Error("should match case-insensitive")
	}
}

func TestIsHTMLClosingTag(t *testing.T) {
	if !IsHTMLClosingTag(`</mention-doc>`, "mention-doc") {
		t.Error("should match closing tag")
	}
	if IsHTMLClosingTag(`<mention-doc>`, "mention-doc") {
		t.Error("opening tag should not match")
	}
}

func TestMapDocTypeToObjType(t *testing.T) {
	tests := map[string]int{
		"doc": 1, "sheet": 3, "bitable": 8, "docx": 22,
		"wiki": 16, "DOCX": 22, "unknown": 22,
	}
	for input, expected := range tests {
		got := mapDocTypeToObjType(input)
		if got != expected {
			t.Errorf("mapDocTypeToObjType(%q) = %d, want %d", input, got, expected)
		}
	}
}

func TestMapObjTypeToDocType(t *testing.T) {
	tests := map[int]string{
		1: "doc", 3: "sheet", 8: "bitable", 22: "docx",
		16: "wiki", 999: "docx",
	}
	for input, expected := range tests {
		v := input
		got := mapObjTypeToDocType(&v)
		if got != expected {
			t.Errorf("mapObjTypeToDocType(%d) = %q, want %q", input, got, expected)
		}
	}
	if got := mapObjTypeToDocType(nil); got != "docx" {
		t.Errorf("mapObjTypeToDocType(nil) = %q, want 'docx'", got)
	}
}
