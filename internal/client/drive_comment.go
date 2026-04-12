package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// CreateNewCommentReq 创建评论请求（V2 API，支持富文本 reply_elements）
type CreateNewCommentReq struct {
	FileToken     string           // 目标文件 token（docx/doc）
	FileType      string           // docx / doc
	BlockID       string           // 可选：局部评论的 anchor block_id
	ReplyElements []map[string]any // reply_elements 数组
}

// CreateNewComment 创建新评论（V2 API，支持富文本 + 局部评论）
// API: POST /open-apis/drive/v1/files/{file_token}/new_comments
// body: {file_type, reply_elements, [anchor.block_id]}
// 返回原始 data 字段（含 comment_id、create_time、is_whole 等）
func CreateNewComment(req CreateNewCommentReq, userAccessToken string) (json.RawMessage, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	body := map[string]any{
		"file_type":      req.FileType,
		"reply_elements": req.ReplyElements,
	}
	if req.BlockID != "" {
		body["anchor"] = map[string]any{
			"block_id": req.BlockID,
		}
	}

	apiPath := fmt.Sprintf("/open-apis/drive/v1/files/%s/new_comments", url.PathEscape(req.FileToken))

	tokenType, opts := resolveTokenOpts(userAccessToken)
	resp, err := client.Post(Context(), apiPath, body, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("创建评论失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("创建评论失败: HTTP %d, body: %s", resp.StatusCode, string(resp.RawBody))
	}

	var apiResp struct {
		Code int             `json:"code"`
		Msg  string          `json:"msg"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(resp.RawBody, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	if apiResp.Code != 0 {
		return nil, fmt.Errorf("创建评论失败: code=%d, msg=%s", apiResp.Code, apiResp.Msg)
	}
	return apiResp.Data, nil
}
