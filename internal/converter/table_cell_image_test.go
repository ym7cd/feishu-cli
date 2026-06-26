package converter

import (
	"strings"
	"testing"
)

// cellElementsText 把某个单元格的富文本元素拼成纯字符串，便于断言占位文本。
func cellElementsText(td *TableData, i int) string {
	var sb strings.Builder
	if i >= len(td.CellElements) {
		return ""
	}
	for _, e := range td.CellElements[i] {
		if e != nil && e.TextRun != nil && e.TextRun.Content != nil {
			sb.WriteString(*e.TextRun.Content)
		}
	}
	return sb.String()
}

// TestTableCellImages_EmbedOn 验证 issue #164 修复：EmbedTableImages 开启时，
// 表格单元格内的本地图片被收集进 TableData.CellImages（供导入层真嵌入），且不再以占位文本形式留在富文本里。
func TestTableCellImages_EmbedOn(t *testing.T) {
	md := "| 名称 | 图片 | 备注 |\n" +
		"|------|------|------|\n" +
		"| 苹果 | ![苹果图](./apple.png) | 说明 |\n" +
		"| 香蕉 | ![](./banana.png) | 无alt |\n"

	conv := NewMarkdownToBlock([]byte(md), ConvertOptions{UploadImages: true, EmbedTableImages: true}, "")
	res, err := conv.ConvertWithTableData()
	if err != nil {
		t.Fatal(err)
	}
	if len(res.TableDatas) != 1 {
		t.Fatalf("want 1 table, got %d", len(res.TableDatas))
	}
	td := res.TableDatas[0]

	// 3 列 × 3 行（表头 + 2 数据行）= 9 个单元格，CellImages 长度须与之对齐
	if len(td.CellImages) != 9 {
		t.Fatalf("CellImages len = %d, want 9 (3×3，与最终单元格数对齐)", len(td.CellImages))
	}

	// 图片应落在 cell[4]（苹果行图片格）和 cell[7]（香蕉行图片格）
	if len(td.CellImages[4]) != 1 || !strings.Contains(td.CellImages[4][0], "apple.png") {
		t.Fatalf("cell[4] images = %v, want [apple.png]", td.CellImages[4])
	}
	if len(td.CellImages[7]) != 1 || !strings.Contains(td.CellImages[7][0], "banana.png") {
		t.Fatalf("cell[7] images = %v, want [banana.png]", td.CellImages[7])
	}
	// 其它单元格不应有图片
	for i, imgs := range td.CellImages {
		if i == 4 || i == 7 {
			continue
		}
		if len(imgs) != 0 {
			t.Fatalf("cell[%d] 不应有图片, got %v", i, imgs)
		}
	}

	// 图片格的富文本不应再含 [Image:/[图片: 占位（已被收集，等待真嵌入）
	if txt := cellElementsText(td, 4); strings.Contains(txt, "[Image:") || strings.Contains(txt, "[图片:") {
		t.Fatalf("cell[4] 富文本不应含图片占位, got %q", txt)
	}
	if txt := cellElementsText(td, 7); strings.Contains(txt, "[Image:") || strings.Contains(txt, "[图片:") {
		t.Fatalf("cell[7] 富文本不应含图片占位, got %q", txt)
	}
}

// TestTableCellImages_EmbedOff_FallsBackToPlaceholder 验证 EmbedTableImages 关闭时（如 doc content-update）
// 单元格本地图片降级为占位文本（方案A），不静默丢失，且 CellImages 保持 nil（不触发嵌入）。
func TestTableCellImages_EmbedOff_FallsBackToPlaceholder(t *testing.T) {
	md := "| 名称 | 图片 |\n" +
		"|------|------|\n" +
		"| 苹果 | ![](./apple.png) |\n"

	conv := NewMarkdownToBlock([]byte(md), ConvertOptions{UploadImages: true, EmbedTableImages: false}, "")
	res, err := conv.ConvertWithTableData()
	if err != nil {
		t.Fatal(err)
	}
	td := res.TableDatas[0]

	if td.CellImages != nil {
		t.Fatalf("EmbedTableImages=false 时 CellImages 应为 nil, got %v", td.CellImages)
	}

	// 单元格布局：表头 名称(0) 图片(1)；数据行 苹果(2) 图片格(3)。图片格应有 [Image: ...] 占位文本。
	if txt := cellElementsText(td, 3); !strings.Contains(txt, "[Image:") {
		t.Fatalf("EmbedTableImages=false 时本地图片格应有占位文本, got %q", txt)
	}
}
