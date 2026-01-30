package client

import (
	"fmt"

	larkwiki "github.com/larksuite/oapi-sdk-go/v3/service/wiki/v2"
)

// WikiNode 知识库节点信息
type WikiNode struct {
	SpaceID         string `json:"space_id"`
	NodeToken       string `json:"node_token"`
	ObjToken        string `json:"obj_token"`
	ObjType         string `json:"obj_type"`
	ParentNodeToken string `json:"parent_node_token,omitempty"`
	NodeType        string `json:"node_type"`
	Title           string `json:"title"`
	HasChild        bool   `json:"has_child"`
	Creator         string `json:"creator,omitempty"`
	Owner           string `json:"owner,omitempty"`
	ObjCreateTime   string `json:"obj_create_time,omitempty"`
	ObjEditTime     string `json:"obj_edit_time,omitempty"`
}

// WikiSpace 知识空间信息
type WikiSpace struct {
	SpaceID     string `json:"space_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	SpaceType   string `json:"space_type,omitempty"`
	Visibility  string `json:"visibility,omitempty"`
}

// GetWikiNode 获取知识库节点信息
func GetWikiNode(token string) (*WikiNode, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	req := larkwiki.NewGetNodeSpaceReqBuilder().
		Token(token).
		Build()

	resp, err := client.Wiki.Space.GetNode(Context(), req)
	if err != nil {
		return nil, fmt.Errorf("获取节点信息失败: %w", err)
	}

	if !resp.Success() {
		return nil, fmt.Errorf("获取节点信息失败: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	node := resp.Data.Node
	if node == nil {
		return nil, fmt.Errorf("节点不存在")
	}

	return &WikiNode{
		NodeToken:       token,
		SpaceID:         StringVal(node.SpaceId),
		ObjToken:        StringVal(node.ObjToken),
		ObjType:         StringVal(node.ObjType),
		ParentNodeToken: StringVal(node.ParentNodeToken),
		NodeType:        StringVal(node.NodeType),
		Title:           StringVal(node.Title),
		HasChild:        BoolVal(node.HasChild),
		Creator:         StringVal(node.Creator),
		Owner:           StringVal(node.Owner),
		ObjCreateTime:   StringVal(node.ObjCreateTime),
		ObjEditTime:     StringVal(node.ObjEditTime),
	}, nil
}

// ListWikiSpaces 获取知识空间列表
func ListWikiSpaces(pageSize int, pageToken string) ([]*WikiSpace, string, bool, error) {
	client, err := GetClient()
	if err != nil {
		return nil, "", false, err
	}

	reqBuilder := larkwiki.NewListSpaceReqBuilder()
	if pageSize > 0 {
		reqBuilder.PageSize(pageSize)
	}
	if pageToken != "" {
		reqBuilder.PageToken(pageToken)
	}

	resp, err := client.Wiki.Space.List(Context(), reqBuilder.Build())
	if err != nil {
		return nil, "", false, fmt.Errorf("获取知识空间列表失败: %w", err)
	}

	if !resp.Success() {
		return nil, "", false, fmt.Errorf("获取知识空间列表失败: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	var spaces []*WikiSpace
	if resp.Data != nil && resp.Data.Items != nil {
		for _, item := range resp.Data.Items {
			spaces = append(spaces, &WikiSpace{
				SpaceID:     StringVal(item.SpaceId),
				Name:        StringVal(item.Name),
				Description: StringVal(item.Description),
				SpaceType:   StringVal(item.SpaceType),
				Visibility:  StringVal(item.Visibility),
			})
		}
	}

	var nextPageToken string
	var hasMore bool
	if resp.Data != nil {
		nextPageToken = StringVal(resp.Data.PageToken)
		hasMore = BoolVal(resp.Data.HasMore)
	}

	return spaces, nextPageToken, hasMore, nil
}

// ListWikiNodes 获取知识空间下的节点列表
func ListWikiNodes(spaceID string, parentNodeToken string, pageSize int, pageToken string) ([]*WikiNode, string, bool, error) {
	client, err := GetClient()
	if err != nil {
		return nil, "", false, err
	}

	reqBuilder := larkwiki.NewListSpaceNodeReqBuilder().
		SpaceId(spaceID)

	if parentNodeToken != "" {
		reqBuilder.ParentNodeToken(parentNodeToken)
	}
	if pageSize > 0 {
		reqBuilder.PageSize(pageSize)
	}
	if pageToken != "" {
		reqBuilder.PageToken(pageToken)
	}

	resp, err := client.Wiki.SpaceNode.List(Context(), reqBuilder.Build())
	if err != nil {
		return nil, "", false, fmt.Errorf("获取节点列表失败: %w", err)
	}

	if !resp.Success() {
		return nil, "", false, fmt.Errorf("获取节点列表失败: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	var nodes []*WikiNode
	if resp.Data != nil && resp.Data.Items != nil {
		for _, item := range resp.Data.Items {
			nodes = append(nodes, &WikiNode{
				SpaceID:         StringVal(item.SpaceId),
				NodeToken:       StringVal(item.NodeToken),
				ObjToken:        StringVal(item.ObjToken),
				ObjType:         StringVal(item.ObjType),
				ParentNodeToken: StringVal(item.ParentNodeToken),
				NodeType:        StringVal(item.NodeType),
				Title:           StringVal(item.Title),
				HasChild:        BoolVal(item.HasChild),
				Creator:         StringVal(item.Creator),
				Owner:           StringVal(item.Owner),
			})
		}
	}

	var nextPageToken string
	var hasMore bool
	if resp.Data != nil {
		nextPageToken = StringVal(resp.Data.PageToken)
		hasMore = BoolVal(resp.Data.HasMore)
	}

	return nodes, nextPageToken, hasMore, nil
}

// CreateWikiNodeResult 创建节点的结果
type CreateWikiNodeResult struct {
	SpaceID   string `json:"space_id"`
	NodeToken string `json:"node_token"`
	ObjToken  string `json:"obj_token"`
	ObjType   string `json:"obj_type"`
}

// CreateWikiNode 在知识空间中创建节点
func CreateWikiNode(spaceID, title, parentNode, nodeType string) (*CreateWikiNodeResult, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	nodeBuilder := larkwiki.NewNodeBuilder().
		Title(title)

	if nodeType == "" {
		nodeType = larkwiki.ObjTypeObjTypeDocx
	}
	nodeBuilder.ObjType(nodeType)

	if parentNode != "" {
		nodeBuilder.ParentNodeToken(parentNode)
	}

	req := larkwiki.NewCreateSpaceNodeReqBuilder().
		SpaceId(spaceID).
		Node(nodeBuilder.Build()).
		Build()

	resp, err := client.Wiki.SpaceNode.Create(Context(), req)
	if err != nil {
		return nil, fmt.Errorf("创建知识库节点失败: %w", err)
	}

	if !resp.Success() {
		return nil, fmt.Errorf("创建知识库节点失败: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	if resp.Data == nil || resp.Data.Node == nil {
		return nil, fmt.Errorf("创建节点成功但未返回节点信息")
	}

	node := resp.Data.Node
	return &CreateWikiNodeResult{
		SpaceID:   StringVal(node.SpaceId),
		NodeToken: StringVal(node.NodeToken),
		ObjToken:  StringVal(node.ObjToken),
		ObjType:   StringVal(node.ObjType),
	}, nil
}

// UpdateWikiNode 更新知识库节点标题
func UpdateWikiNode(spaceID, nodeToken, title string) error {
	client, err := GetClient()
	if err != nil {
		return err
	}

	body := larkwiki.NewUpdateTitleSpaceNodeReqBodyBuilder().
		Title(title).
		Build()

	req := larkwiki.NewUpdateTitleSpaceNodeReqBuilder().
		SpaceId(spaceID).
		NodeToken(nodeToken).
		Body(body).
		Build()

	resp, err := client.Wiki.SpaceNode.UpdateTitle(Context(), req)
	if err != nil {
		return fmt.Errorf("更新知识库节点标题失败: %w", err)
	}

	if !resp.Success() {
		return fmt.Errorf("更新知识库节点标题失败: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// MoveWikiNodeResult 移动节点的结果
type MoveWikiNodeResult struct {
	NodeToken string `json:"node_token"`
}

// MoveWikiNode 移动知识库节点
func MoveWikiNode(spaceID, nodeToken, targetSpaceID, targetParent string) (*MoveWikiNodeResult, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	bodyBuilder := larkwiki.NewMoveSpaceNodeReqBodyBuilder().
		TargetSpaceId(targetSpaceID)

	if targetParent != "" {
		bodyBuilder.TargetParentToken(targetParent)
	}

	req := larkwiki.NewMoveSpaceNodeReqBuilder().
		SpaceId(spaceID).
		NodeToken(nodeToken).
		Body(bodyBuilder.Build()).
		Build()

	resp, err := client.Wiki.SpaceNode.Move(Context(), req)
	if err != nil {
		return nil, fmt.Errorf("移动知识库节点失败: %w", err)
	}

	if !resp.Success() {
		return nil, fmt.Errorf("移动知识库节点失败: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	result := &MoveWikiNodeResult{}
	if resp.Data != nil && resp.Data.Node != nil {
		result.NodeToken = StringVal(resp.Data.Node.NodeToken)
	}

	return result, nil
}
