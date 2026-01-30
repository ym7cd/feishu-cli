package client

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
)

// 最大下载文件大小限制 (100MB)
const maxDownloadSize = 100 * 1024 * 1024

// 下载超时时间
const downloadTimeout = 5 * time.Minute

// UploadMedia uploads a file to Feishu drive
func UploadMedia(filePath string, parentType string, parentNode string, fileName string) (string, error) {
	client, err := GetClient()
	if err != nil {
		return "", err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("获取文件信息失败: %w", err)
	}
	fileSize := int(stat.Size())

	if fileName == "" {
		fileName = filepath.Base(filePath)
	}

	req := larkdrive.NewUploadAllMediaReqBuilder().
		Body(larkdrive.NewUploadAllMediaReqBodyBuilder().
			FileName(fileName).
			ParentType(parentType).
			ParentNode(parentNode).
			Size(fileSize).
			File(file).
			Build()).
		Build()

	resp, err := client.Drive.Media.UploadAll(Context(), req)
	if err != nil {
		return "", fmt.Errorf("上传素材失败: %w", err)
	}

	if !resp.Success() {
		return "", fmt.Errorf("上传素材失败: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	if resp.Data.FileToken == nil {
		return "", fmt.Errorf("上传成功但未返回文件 Token")
	}

	return *resp.Data.FileToken, nil
}

// DownloadMedia downloads a file from Feishu drive
func DownloadMedia(fileToken string, outputPath string) error {
	if err := validatePath(outputPath); err != nil {
		return err
	}

	client, err := GetClient()
	if err != nil {
		return err
	}

	req := larkdrive.NewDownloadMediaReqBuilder().
		FileToken(fileToken).
		Build()

	resp, err := client.Drive.Media.Download(ContextWithTimeout(downloadTimeout), req)
	if err != nil {
		return fmt.Errorf("下载素材失败: %w", err)
	}

	if !resp.Success() {
		return fmt.Errorf("下载素材失败: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return saveToFile(resp.File, outputPath)
}

// GetMediaTempURL gets a temporary download URL for a media file
func GetMediaTempURL(fileToken string) (string, error) {
	client, err := GetClient()
	if err != nil {
		return "", err
	}

	req := larkdrive.NewBatchGetTmpDownloadUrlMediaReqBuilder().
		FileTokens([]string{fileToken}).
		Build()

	resp, err := client.Drive.Media.BatchGetTmpDownloadUrl(Context(), req)
	if err != nil {
		return "", fmt.Errorf("获取临时下载链接失败: %w", err)
	}

	if !resp.Success() {
		return "", fmt.Errorf("获取临时下载链接失败: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	if len(resp.Data.TmpDownloadUrls) == 0 {
		return "", fmt.Errorf("未返回下载链接")
	}

	if resp.Data.TmpDownloadUrls[0].TmpDownloadUrl == nil {
		return "", fmt.Errorf("下载链接为空")
	}

	return *resp.Data.TmpDownloadUrls[0].TmpDownloadUrl, nil
}

// DownloadFromURL downloads a file from a URL with size limit
func DownloadFromURL(url string, outputPath string) error {
	if err := validatePath(outputPath); err != nil {
		return err
	}

	httpClient := &http.Client{
		Timeout: downloadTimeout,
	}

	resp, err := httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("从 URL 下载失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败: HTTP 状态码 %d", resp.StatusCode)
	}

	if resp.ContentLength > maxDownloadSize {
		return fmt.Errorf("文件超过大小限制: %d MB (限制 %d MB)",
			resp.ContentLength/(1024*1024), maxDownloadSize/(1024*1024))
	}

	return saveToFile(resp.Body, outputPath)
}

// saveToFile 将 reader 内容写入文件，限制最大大小
func saveToFile(reader io.Reader, outputPath string) error {
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer outFile.Close()

	limitedReader := io.LimitReader(reader, maxDownloadSize)
	written, err := io.Copy(outFile, limitedReader)
	if err != nil {
		outFile.Close()
		os.Remove(outputPath)
		return fmt.Errorf("写入文件失败: %w", err)
	}

	if written >= maxDownloadSize {
		outFile.Close()
		os.Remove(outputPath)
		return fmt.Errorf("文件超过大小限制 (%d MB)", maxDownloadSize/(1024*1024))
	}

	return nil
}

// validatePath 验证路径安全性，防止路径遍历攻击
func validatePath(path string) error {
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("路径不安全: 不允许包含 '..'")
	}
	return nil
}

// DriveFile 云空间文件信息
type DriveFile struct {
	Token        string `json:"token"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	ParentToken  string `json:"parent_token,omitempty"`
	URL          string `json:"url,omitempty"`
	CreatedTime  string `json:"created_time,omitempty"`
	ModifiedTime string `json:"modified_time,omitempty"`
	OwnerID      string `json:"owner_id,omitempty"`
}

// ListFiles 列出文件夹中的文件
func ListFiles(folderToken string, pageSize int, pageToken string) ([]*DriveFile, string, bool, error) {
	client, err := GetClient()
	if err != nil {
		return nil, "", false, err
	}

	reqBuilder := larkdrive.NewListFileReqBuilder()
	if folderToken != "" {
		reqBuilder.FolderToken(folderToken)
	}
	if pageSize > 0 {
		reqBuilder.PageSize(pageSize)
	}
	if pageToken != "" {
		reqBuilder.PageToken(pageToken)
	}

	resp, err := client.Drive.File.List(Context(), reqBuilder.Build())
	if err != nil {
		return nil, "", false, fmt.Errorf("获取文件列表失败: %w", err)
	}

	if !resp.Success() {
		return nil, "", false, fmt.Errorf("获取文件列表失败: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	var files []*DriveFile
	if resp.Data != nil && resp.Data.Files != nil {
		for _, f := range resp.Data.Files {
			files = append(files, &DriveFile{
				Token:        StringVal(f.Token),
				Name:         StringVal(f.Name),
				Type:         StringVal(f.Type),
				ParentToken:  StringVal(f.ParentToken),
				URL:          StringVal(f.Url),
				CreatedTime:  StringVal(f.CreatedTime),
				ModifiedTime: StringVal(f.ModifiedTime),
				OwnerID:      StringVal(f.OwnerId),
			})
		}
	}

	var nextPageToken string
	var hasMore bool
	if resp.Data != nil {
		nextPageToken = StringVal(resp.Data.NextPageToken)
		hasMore = BoolVal(resp.Data.HasMore)
	}

	return files, nextPageToken, hasMore, nil
}

// CreateFolder 创建文件夹
func CreateFolder(name string, folderToken string) (string, string, error) {
	client, err := GetClient()
	if err != nil {
		return "", "", err
	}

	req := larkdrive.NewCreateFolderFileReqBuilder().
		Body(larkdrive.NewCreateFolderFileReqBodyBuilder().
			Name(name).
			FolderToken(folderToken).
			Build()).
		Build()

	resp, err := client.Drive.File.CreateFolder(Context(), req)
	if err != nil {
		return "", "", fmt.Errorf("创建文件夹失败: %w", err)
	}

	if !resp.Success() {
		return "", "", fmt.Errorf("创建文件夹失败: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	var token, url string
	if resp.Data != nil {
		token = StringVal(resp.Data.Token)
		url = StringVal(resp.Data.Url)
	}

	return token, url, nil
}

// MoveFile 移动文件或文件夹
func MoveFile(fileToken string, targetFolderToken string, fileType string) (string, error) {
	client, err := GetClient()
	if err != nil {
		return "", err
	}

	req := larkdrive.NewMoveFileReqBuilder().
		Body(larkdrive.NewMoveFileReqBodyBuilder().
			Type(fileType).
			FolderToken(targetFolderToken).
			Build()).
		FileToken(fileToken).
		Build()

	resp, err := client.Drive.File.Move(Context(), req)
	if err != nil {
		return "", fmt.Errorf("移动文件失败: %w", err)
	}

	if !resp.Success() {
		return "", fmt.Errorf("移动文件失败: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	if resp.Data != nil && resp.Data.TaskId != nil {
		return *resp.Data.TaskId, nil
	}

	return "", nil
}

// CopyFile 复制文件
func CopyFile(fileToken string, targetFolderToken string, name string, fileType string) (string, string, error) {
	client, err := GetClient()
	if err != nil {
		return "", "", err
	}

	reqBuilder := larkdrive.NewCopyFileReqBodyBuilder().
		Type(fileType).
		FolderToken(targetFolderToken)

	if name != "" {
		reqBuilder.Name(name)
	}

	req := larkdrive.NewCopyFileReqBuilder().
		FileToken(fileToken).
		Body(reqBuilder.Build()).
		Build()

	resp, err := client.Drive.File.Copy(Context(), req)
	if err != nil {
		return "", "", fmt.Errorf("复制文件失败: %w", err)
	}

	if !resp.Success() {
		return "", "", fmt.Errorf("复制文件失败: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	var token, url string
	if resp.Data != nil && resp.Data.File != nil {
		token = StringVal(resp.Data.File.Token)
		url = StringVal(resp.Data.File.Url)
	}

	return token, url, nil
}

// DeleteFile 删除文件或文件夹
func DeleteFile(fileToken string, fileType string) (string, error) {
	client, err := GetClient()
	if err != nil {
		return "", err
	}

	req := larkdrive.NewDeleteFileReqBuilder().
		FileToken(fileToken).
		Type(fileType).
		Build()

	resp, err := client.Drive.File.Delete(Context(), req)
	if err != nil {
		return "", fmt.Errorf("删除文件失败: %w", err)
	}

	if !resp.Success() {
		return "", fmt.Errorf("删除文件失败: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	if resp.Data != nil && resp.Data.TaskId != nil {
		return *resp.Data.TaskId, nil
	}

	return "", nil
}

// ShortcutInfo 快捷方式信息
type ShortcutInfo struct {
	Token       string `json:"token"`
	TargetToken string `json:"target_token"`
	TargetType  string `json:"target_type"`
	ParentToken string `json:"parent_token,omitempty"`
}

// CreateShortcut 创建文件快捷方式
func CreateShortcut(parentToken string, targetFileToken string, targetType string) (*ShortcutInfo, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	req := larkdrive.NewCreateShortcutFileReqBuilder().
		Body(larkdrive.NewCreateShortcutFileReqBodyBuilder().
			ParentToken(parentToken).
			ReferEntity(larkdrive.NewReferEntityBuilder().
				ReferToken(targetFileToken).
				ReferType(targetType).
				Build()).
			Build()).
		Build()

	resp, err := client.Drive.File.CreateShortcut(Context(), req)
	if err != nil {
		return nil, fmt.Errorf("创建快捷方式失败: %w", err)
	}

	if !resp.Success() {
		return nil, fmt.Errorf("创建快捷方式失败: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	info := &ShortcutInfo{
		TargetToken: targetFileToken,
		TargetType:  targetType,
	}
	if resp.Data != nil && resp.Data.SuccShortcutNode != nil {
		info.Token = StringVal(resp.Data.SuccShortcutNode.Token)
		info.ParentToken = StringVal(resp.Data.SuccShortcutNode.ParentToken)
	}

	return info, nil
}

// DriveQuota 云空间容量信息
type DriveQuota struct {
	Total int64 `json:"total"` // 总容量（字节）
	Used  int64 `json:"used"`  // 已用容量（字节）
}

// GetDriveQuota 获取云空间容量信息
// 注意：当前飞书 SDK 版本不支持此 API
func GetDriveQuota() (*DriveQuota, error) {
	return nil, fmt.Errorf("获取云空间容量功能暂不支持：当前 SDK 版本未提供此 API")
}
