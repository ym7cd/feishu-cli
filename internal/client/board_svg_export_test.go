package client

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

func TestParseExportWhiteboardSVGResponse_OK(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg"><rect width="10" height="10"/></svg>`
	enc := base64.StdEncoding.EncodeToString([]byte(svg))
	body := fmt.Sprintf(`{"code":0,"msg":"","data":{"content":%q,"mime_type":"image/svg+xml"}}`, enc)

	got, err := parseExportWhiteboardSVGResponse([]byte(body))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if got.SVG != svg {
		t.Errorf("SVG 解码不匹配\n got: %s\nwant: %s", got.SVG, svg)
	}
	if got.MimeType != "image/svg+xml" {
		t.Errorf("mime_type 不匹配: %s", got.MimeType)
	}
}

func TestParseExportWhiteboardSVGResponse_NonZeroCode(t *testing.T) {
	body := `{"code":1254030,"msg":"no permission to export","data":null}`
	_, err := parseExportWhiteboardSVGResponse([]byte(body))
	if err == nil {
		t.Fatal("code 非 0 应报错")
	}
	if !strings.Contains(err.Error(), "1254030") || !strings.Contains(err.Error(), "no permission") {
		t.Errorf("错误应含 code 和 msg，得到: %v", err)
	}
}

func TestParseExportWhiteboardSVGResponse_EmptyContent(t *testing.T) {
	body := `{"code":0,"msg":"","data":{"content":"","mime_type":""}}`
	_, err := parseExportWhiteboardSVGResponse([]byte(body))
	if err == nil {
		t.Fatal("content 为空应报错")
	}
}

func TestParseExportWhiteboardSVGResponse_InvalidBase64(t *testing.T) {
	body := `{"code":0,"msg":"","data":{"content":"!!!not-base64!!!","mime_type":""}}`
	_, err := parseExportWhiteboardSVGResponse([]byte(body))
	if err == nil {
		t.Fatal("非法 base64 应报错")
	}
	if !strings.Contains(err.Error(), "base64") {
		t.Errorf("错误应提示 base64 解码失败，得到: %v", err)
	}
}

func TestParseExportWhiteboardSVGResponse_InvalidJSON(t *testing.T) {
	_, err := parseExportWhiteboardSVGResponse([]byte("not-json"))
	if err == nil {
		t.Fatal("非法 JSON 应报错")
	}
}
