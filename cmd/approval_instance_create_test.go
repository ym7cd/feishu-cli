package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func newApproverCCCmd() *cobra.Command {
	c := &cobra.Command{Use: "test"}
	c.Flags().String("node-approver", "", "")
	c.Flags().String("node-approver-file", "", "")
	c.Flags().String("node-cc", "", "")
	c.Flags().String("node-cc-file", "", "")
	return c
}

func TestLoadNodeApproverCC_AllEmpty(t *testing.T) {
	c := newApproverCCCmd()
	a, cc, err := loadNodeApproverCC(c)
	if err != nil {
		t.Fatalf("都空时不应报错，实际: %v", err)
	}
	if a != nil || cc != nil {
		t.Fatalf("都空时期望 nil/nil，实际 %s / %s", a, cc)
	}
}

func TestLoadNodeApproverCC_InlineArray(t *testing.T) {
	c := newApproverCCCmd()
	payload := `[{"node_id":"n1","value":["ou_a","ou_b"]}]`
	if err := c.Flags().Set("node-approver", payload); err != nil {
		t.Fatal(err)
	}
	a, cc, err := loadNodeApproverCC(c)
	if err != nil {
		t.Fatalf("合法 JSON 数组不应报错: %v", err)
	}
	if string(a) != payload {
		t.Errorf("approver 原文不匹配: got %s", a)
	}
	if cc != nil {
		t.Errorf("未设置 cc 时应为 nil")
	}
}

func TestLoadNodeApproverCC_FileAndBoth(t *testing.T) {
	dir := t.TempDir()
	approverPath := filepath.Join(dir, "approver.json")
	ccPayload := `[{"node_id":"n2","value":["ou_c"]}]`
	if err := os.WriteFile(approverPath, []byte(ccPayload), 0o644); err != nil {
		t.Fatal(err)
	}

	c := newApproverCCCmd()
	if err := c.Flags().Set("node-approver", `[{"node_id":"n1","value":["ou_a"]}]`); err != nil {
		t.Fatal(err)
	}
	if err := c.Flags().Set("node-cc-file", approverPath); err != nil {
		t.Fatal(err)
	}
	a, cc, err := loadNodeApproverCC(c)
	if err != nil {
		t.Fatalf("内联 + 文件组合不应报错: %v", err)
	}
	if a == nil || cc == nil {
		t.Fatalf("两者都应解析到值: a=%s cc=%s", a, cc)
	}
	if string(cc) != ccPayload {
		t.Errorf("cc 文件内容不匹配: got %s", cc)
	}
}

func TestLoadNodeApproverCC_NotArray(t *testing.T) {
	c := newApproverCCCmd()
	if err := c.Flags().Set("node-approver", `{"node_id":"n1"}`); err != nil {
		t.Fatal(err)
	}
	_, _, err := loadNodeApproverCC(c)
	if err == nil {
		t.Fatal("非数组 JSON 应报错")
	}
}

func TestLoadNodeApproverCC_InvalidJSON(t *testing.T) {
	c := newApproverCCCmd()
	if err := c.Flags().Set("node-cc", "not-json"); err != nil {
		t.Fatal(err)
	}
	_, _, err := loadNodeApproverCC(c)
	if err == nil {
		t.Fatal("非法 JSON 应报错")
	}
}
