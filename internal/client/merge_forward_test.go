package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// mfMsgJSON 生成一条 message JSON 片段，仅含测试关心的字段。
// upper 传空字符串则不写该字段（用于容器本身）。
func mfMsgJSON(id, msgType, upper string) string {
	if upper == "" {
		return fmt.Sprintf(`{"message_id":%q,"msg_type":%q,"body":{"content":"placeholder"}}`, id, msgType)
	}
	return fmt.Sprintf(`{"message_id":%q,"msg_type":%q,"upper_message_id":%q,"body":{"content":"placeholder"}}`, id, msgType, upper)
}

// mergeForwardStubRouter 按 message_id 切换响应；用 tenant_access_token 自动 stub。
// messages key 为 message_id，value 为 data.items JSON 数组字符串。
// 未注册的 message_id 返回飞书风格 errcode（非 0），上层应当作展开失败容错。
func mergeForwardStubRouter(t *testing.T, messages map[string]string, hits *atomic.Int64) http.HandlerFunc {
	t.Helper()
	return tenantRouteHandler(t, func(w http.ResponseWriter, r *http.Request) {
		const prefix = "/open-apis/im/v1/messages/"
		if !strings.HasPrefix(r.URL.Path, prefix) {
			http.Error(w, "unexpected path", http.StatusNotFound)
			return
		}
		msgID := strings.TrimPrefix(r.URL.Path, prefix)
		if hits != nil {
			hits.Add(1)
		}
		items, ok := messages[msgID]
		w.Header().Set("Content-Type", "application/json")
		if !ok {
			_, _ = fmt.Fprintf(w, `{"code":230002,"msg":"message %s not registered","data":{}}`, msgID)
			return
		}
		_, _ = fmt.Fprintf(w, `{"code":0,"msg":"success","data":{"items":%s}}`, items)
	})
}

// TestExpandMergeForward_Flat 验证单层展开：容器内全是非嵌套子消息。
func TestExpandMergeForward_Flat(t *testing.T) {
	const root = "om_root"
	topo := map[string]string{
		root: "[" + strings.Join([]string{
			mfMsgJSON(root, "merge_forward", ""),
			mfMsgJSON("om_sub_1", "text", root),
			mfMsgJSON("om_sub_2", "text", root),
			mfMsgJSON("om_sub_3", "interactive", root),
		}, ",") + "]",
	}
	_, cleanup := stubFeishuServer(t, mergeForwardStubRouter(t, topo, nil))
	defer cleanup()

	subs, err := expandMergeForward(root, "", 0, make(map[string]bool))
	if err != nil {
		t.Fatalf("展开失败: %v", err)
	}
	if len(subs) != 3 {
		t.Fatalf("子消息数量: got %d, want 3", len(subs))
	}
	wantIDs := []string{"om_sub_1", "om_sub_2", "om_sub_3"}
	for i, want := range wantIDs {
		if got := StringVal(subs[i].MessageId); got != want {
			t.Errorf("子消息[%d] id: got %q, want %q", i, got, want)
		}
		if got := StringVal(subs[i].UpperMessageId); got != root {
			t.Errorf("子消息[%d] upper: got %q, want %q", i, got, root)
		}
	}
}

// TestExpandMergeForward_Recursive 验证递归展开：A 包含 B（mf）和 C；B 内有 D、E。
// 期望顺序：B、D、E、C（深度优先：每个 sub 先 append 自身再递归 append 其子级）。
func TestExpandMergeForward_Recursive(t *testing.T) {
	const root = "om_A"
	topo := map[string]string{
		root: "[" + strings.Join([]string{
			mfMsgJSON(root, "merge_forward", ""),
			mfMsgJSON("om_B", "merge_forward", root),
			mfMsgJSON("om_C", "text", root),
		}, ",") + "]",
		"om_B": "[" + strings.Join([]string{
			mfMsgJSON("om_B", "merge_forward", ""),
			mfMsgJSON("om_D", "text", "om_B"),
			mfMsgJSON("om_E", "image", "om_B"),
		}, ",") + "]",
	}
	_, cleanup := stubFeishuServer(t, mergeForwardStubRouter(t, topo, nil))
	defer cleanup()

	subs, err := expandMergeForward(root, "", 0, make(map[string]bool))
	if err != nil {
		t.Fatalf("展开失败: %v", err)
	}
	gotIDs := make([]string, len(subs))
	for i, s := range subs {
		gotIDs[i] = StringVal(s.MessageId)
	}
	want := []string{"om_B", "om_D", "om_E", "om_C"}
	if strings.Join(gotIDs, ",") != strings.Join(want, ",") {
		t.Errorf("递归展开顺序: got %v, want %v", gotIDs, want)
	}
}

// TestExpandMergeForward_DepthLimit 构造 mergeForwardMaxDepth+2 层链式嵌套，验证触顶停止。
func TestExpandMergeForward_DepthLimit(t *testing.T) {
	topo := make(map[string]string)
	// 链：om_0 → om_1 → ... → om_{N+1}，每层都是 merge_forward
	depth := mergeForwardMaxDepth + 2
	for i := 0; i <= depth; i++ {
		container := fmt.Sprintf("om_%d", i)
		next := fmt.Sprintf("om_%d", i+1)
		if i == depth {
			topo[container] = "[" + mfMsgJSON(container, "merge_forward", "") + "]"
			break
		}
		topo[container] = "[" + strings.Join([]string{
			mfMsgJSON(container, "merge_forward", ""),
			mfMsgJSON(next, "merge_forward", container),
		}, ",") + "]"
	}
	_, cleanup := stubFeishuServer(t, mergeForwardStubRouter(t, topo, nil))
	defer cleanup()

	subs, err := expandMergeForward("om_0", "", 0, make(map[string]bool))
	if err != nil {
		t.Fatalf("展开失败: %v", err)
	}
	// 深度上限：第 0 层调 mfMaxDepth 次后停。
	// 实际展开会产生 mergeForwardMaxDepth 个子消息（每层 1 个 sub）。
	// 第 maxDepth 层时 depth==max → 直接 return nil，不再深入。
	if len(subs) != mergeForwardMaxDepth {
		t.Errorf("深度限制下子消息数: got %d, want %d", len(subs), mergeForwardMaxDepth)
	}
}

// TestExpandMergeForward_Cycle 验证 visited 防 cycle：A→B→A 形成环。
func TestExpandMergeForward_Cycle(t *testing.T) {
	topo := map[string]string{
		"om_A": "[" + strings.Join([]string{
			mfMsgJSON("om_A", "merge_forward", ""),
			mfMsgJSON("om_B", "merge_forward", "om_A"),
		}, ",") + "]",
		"om_B": "[" + strings.Join([]string{
			mfMsgJSON("om_B", "merge_forward", ""),
			mfMsgJSON("om_A", "merge_forward", "om_B"),
		}, ",") + "]",
	}
	_, cleanup := stubFeishuServer(t, mergeForwardStubRouter(t, topo, nil))
	defer cleanup()

	// 用 deadline 检测死循环
	done := make(chan struct{})
	var subs []*larkim.Message
	var err error
	go func() {
		subs, err = expandMergeForward("om_A", "", 0, make(map[string]bool))
		close(done)
	}()
	select {
	case <-done:
		// ok
	case <-time.After(3 * time.Second):
		t.Fatal("expandMergeForward 在 cycle 拓扑下出现死循环")
	}
	if err != nil {
		t.Fatalf("展开失败: %v", err)
	}
	// 期望：[B, A]
	//   - 外层 A 的 sub=B，append B；递归进 B
	//   - B 的 sub=A，append A（保留原始子消息）；但 A 已在 visited 中，跳过递归
	// 这样既保留了"真实存在的子消息"又阻止了无限递归。
	gotIDs := make([]string, len(subs))
	for i, s := range subs {
		gotIDs[i] = StringVal(s.MessageId)
	}
	want := []string{"om_B", "om_A"}
	if strings.Join(gotIDs, ",") != strings.Join(want, ",") {
		t.Errorf("cycle 拓扑展开结果: got %v, want %v", gotIDs, want)
	}
}

// TestExpandMergeForward_APIError 验证 API 错误向上传播。
func TestExpandMergeForward_APIError(t *testing.T) {
	_, cleanup := stubFeishuServer(t, mergeForwardStubRouter(t, map[string]string{}, nil))
	defer cleanup()

	_, err := expandMergeForward("om_not_found", "", 0, make(map[string]bool))
	if err == nil {
		t.Fatal("缺失消息应返回错误，但获得 nil")
	}
}

// TestGetMessage_MergeForwardAutoExpand 验证 GetMessage 收到 merge_forward 后自动展开。
// 模拟：第一次走 SDK builder 拿到 [容器]（按现有逻辑 SDK builder 走 tenant，不传 card_msg_content_type）；
// 然后 defer 触发 expandMergeForward 再调一次（带 raw_card_content）拿完整 items。
// 简化：让 stub 不区分两次调用的 query 参数，都返回完整 items 列表——这样无论第一次还是第二次都能拿到 message。
// 关键是验证 result.SubMessages 被填充。
func TestGetMessage_MergeForwardAutoExpand(t *testing.T) {
	const root = "om_container"
	topo := map[string]string{
		root: "[" + strings.Join([]string{
			mfMsgJSON(root, "merge_forward", ""),
			mfMsgJSON("om_inner_1", "text", root),
			mfMsgJSON("om_inner_2", "interactive", root),
		}, ",") + "]",
	}
	_, cleanup := stubFeishuServer(t, mergeForwardStubRouter(t, topo, nil))
	defer cleanup()

	result, err := GetMessage(root, "", "")
	if err != nil {
		t.Fatalf("GetMessage 返回错误: %v", err)
	}
	if result == nil || result.Message == nil {
		t.Fatal("返回结果为空")
	}
	if got := StringVal(result.Message.MsgType); got != "merge_forward" {
		t.Errorf("Message.MsgType: got %q, want merge_forward", got)
	}
	if len(result.SubMessages) != 2 {
		t.Fatalf("SubMessages 数量: got %d, want 2", len(result.SubMessages))
	}
	if got := StringVal(result.SubMessages[0].MessageId); got != "om_inner_1" {
		t.Errorf("SubMessages[0].MessageId: got %q, want om_inner_1", got)
	}
}

// TestExpandMergeForwardForContainers_Concurrent 验证 list 场景 3 个容器并发展开。
func TestExpandMergeForwardForContainers_Concurrent(t *testing.T) {
	topo := map[string]string{
		"om_C1": "[" + strings.Join([]string{
			mfMsgJSON("om_C1", "merge_forward", ""),
			mfMsgJSON("om_C1_s1", "text", "om_C1"),
		}, ",") + "]",
		"om_C2": "[" + strings.Join([]string{
			mfMsgJSON("om_C2", "merge_forward", ""),
			mfMsgJSON("om_C2_s1", "text", "om_C2"),
			mfMsgJSON("om_C2_s2", "image", "om_C2"),
		}, ",") + "]",
		"om_C3": "[" + strings.Join([]string{
			mfMsgJSON("om_C3", "merge_forward", ""),
			mfMsgJSON("om_C3_s1", "interactive", "om_C3"),
		}, ",") + "]",
	}
	_, cleanup := stubFeishuServer(t, mergeForwardStubRouter(t, topo, nil))
	defer cleanup()

	containers := []*larkim.Message{
		mustMessage(t, "om_C1", "merge_forward"),
		mustMessage(t, "om_C2", "merge_forward"),
		mustMessage(t, "om_text", "text"), // 应被跳过
		mustMessage(t, "om_C3", "merge_forward"),
	}
	result := expandMergeForwardForContainers(containers, "")
	if result == nil {
		t.Fatal("期望返回非 nil map")
	}
	if len(result) != 3 {
		t.Errorf("map 长度: got %d, want 3", len(result))
	}
	if subs, ok := result["om_C1"]; !ok || len(subs) != 1 {
		t.Errorf("om_C1 子消息: got %v", subs)
	}
	if subs, ok := result["om_C2"]; !ok || len(subs) != 2 {
		t.Errorf("om_C2 子消息: got %v", subs)
	}
	if subs, ok := result["om_C3"]; !ok || len(subs) != 1 {
		t.Errorf("om_C3 子消息: got %v", subs)
	}
	if _, ok := result["om_text"]; ok {
		t.Error("非 merge_forward 消息不应出现在 map 中")
	}
}

// TestExpandMergeForwardForContainers_PartialFailure 3 个容器中 1 个返回错误，其他仍返回。
func TestExpandMergeForwardForContainers_PartialFailure(t *testing.T) {
	topo := map[string]string{
		// om_bad 未注册，会返回 230002
		"om_C1": "[" + strings.Join([]string{
			mfMsgJSON("om_C1", "merge_forward", ""),
			mfMsgJSON("om_C1_s", "text", "om_C1"),
		}, ",") + "]",
		"om_C3": "[" + strings.Join([]string{
			mfMsgJSON("om_C3", "merge_forward", ""),
			mfMsgJSON("om_C3_s", "text", "om_C3"),
		}, ",") + "]",
	}
	_, cleanup := stubFeishuServer(t, mergeForwardStubRouter(t, topo, nil))
	defer cleanup()

	containers := []*larkim.Message{
		mustMessage(t, "om_C1", "merge_forward"),
		mustMessage(t, "om_bad", "merge_forward"),
		mustMessage(t, "om_C3", "merge_forward"),
	}
	result := expandMergeForwardForContainers(containers, "")
	if result == nil {
		t.Fatal("期望返回非 nil（至少 2 个成功）")
	}
	if len(result) != 2 {
		t.Errorf("map 长度: got %d, want 2（om_bad 应被跳过）", len(result))
	}
	if _, ok := result["om_bad"]; ok {
		t.Error("om_bad 不应出现在 map 中")
	}
}

// TestExpandMergeForwardForContainers_DisableEnv 验证逃生开关。
func TestExpandMergeForwardForContainers_DisableEnv(t *testing.T) {
	t.Setenv(mergeForwardDisableEnv, "1")
	// 即使有容器也应直接返回 nil，无需 stub
	containers := []*larkim.Message{
		mustMessage(t, "om_any", "merge_forward"),
	}
	if result := expandMergeForwardForContainers(containers, ""); result != nil {
		t.Errorf("逃生开关命中应返回 nil，got %v", result)
	}
}

// TestListMessages_AttachesMergeForwardSubMessages 验证 ListMessages 出口 defer 调用展开后填充 map。
func TestListMessages_AttachesMergeForwardSubMessages(t *testing.T) {
	const containerID = "oc_test"
	topo := map[string]string{
		"om_mf": "[" + strings.Join([]string{
			mfMsgJSON("om_mf", "merge_forward", ""),
			mfMsgJSON("om_sub_1", "text", "om_mf"),
			mfMsgJSON("om_sub_2", "text", "om_mf"),
		}, ",") + "]",
	}
	// 主 handler 区分 list 端点（返回 items 列表）和 messages/{id} 端点（merge_forward 展开）。
	var hits atomic.Int64
	handler := tenantRouteHandler(t, func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.URL.Path == "/open-apis/im/v1/messages" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"code":0,"msg":"success","data":{"items":[`+
				mfMsgJSON("om_text", "text", "")+`,`+
				mfMsgJSON("om_mf", "merge_forward", "")+
				`],"has_more":false,"page_token":""}}`)
			return
		}
		// 展开请求：messages/{id}
		const prefix = "/open-apis/im/v1/messages/"
		if !strings.HasPrefix(r.URL.Path, prefix) {
			http.Error(w, "unexpected", http.StatusNotFound)
			return
		}
		msgID := strings.TrimPrefix(r.URL.Path, prefix)
		w.Header().Set("Content-Type", "application/json")
		if items, ok := topo[msgID]; ok {
			_, _ = fmt.Fprintf(w, `{"code":0,"msg":"success","data":{"items":%s}}`, items)
			return
		}
		_, _ = fmt.Fprintf(w, `{"code":230002,"msg":"not found","data":{}}`)
	})
	_, cleanup := stubFeishuServer(t, handler)
	defer cleanup()

	result, err := ListMessages(containerID, ListMessagesOptions{ContainerIDType: "chat", PageSize: 10}, "")
	if err != nil {
		t.Fatalf("ListMessages 返回错误: %v", err)
	}
	if len(result.Items) != 2 {
		t.Fatalf("items 长度: got %d, want 2", len(result.Items))
	}
	if result.MergeForwardSubMessages == nil {
		t.Fatal("MergeForwardSubMessages 应非空")
	}
	subs, ok := result.MergeForwardSubMessages["om_mf"]
	if !ok || len(subs) != 2 {
		t.Errorf("om_mf 子消息: got %v (ok=%v)", subs, ok)
	}
}

// mustMessage 构造测试用的 *larkim.Message（通过 JSON 反序列化保证 SDK 字段映射正确）。
func mustMessage(t *testing.T, id, msgType string) *larkim.Message {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"message_id": id,
		"msg_type":   msgType,
	})
	var msg larkim.Message
	if err := json.Unmarshal(body, &msg); err != nil {
		t.Fatalf("构造 message 失败: %v", err)
	}
	return &msg
}

// 编译期保证 sync.Once 等未使用的引用不影响测试包。
var _ = sync.Once{}
