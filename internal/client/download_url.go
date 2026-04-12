package client

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DownloadOptions 预签名 URL 下载选项
type DownloadOptions struct {
	OutputDir string        // 输出目录；为空时使用当前工作目录
	Filename  string        // 强制指定文件名；为空时从响应头解析
	Overwrite bool          // 是否覆盖已存在文件
	Timeout   time.Duration // 整个下载的超时；<=0 时不设置（由调用方用 ctx 控制）
}

// DownloadResult 下载结果
type DownloadResult struct {
	SavedPath string
	Size      int64
	Filename  string
}

// 飞书预签名 URL 下载最大重定向次数
const downloadMaxRedirects = 5

// DownloadFromPresignedURL 从预签名 URL 下载文件到磁盘
// 特性：
//   - SSRF 防护：拒绝内网 IP / localhost / file://
//   - 重定向防护：最多 5 次，拒绝 HTTPS → HTTP 降级，每次重定向重新校验 host
//   - 文件名解析优先级：opts.Filename > Content-Disposition > Content-Type 扩展 > defaultFallbackName
//   - 流式写文件，失败时清理部分文件
func DownloadFromPresignedURL(presignedURL string, defaultFallbackName string, opts DownloadOptions) (*DownloadResult, error) {
	if err := validateDownloadURL(presignedURL); err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Timeout: opts.Timeout, // 0 表示不设超时
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= downloadMaxRedirects {
				return fmt.Errorf("下载重定向次数超过 %d", downloadMaxRedirects)
			}
			// 拒绝 HTTPS → HTTP 降级
			if len(via) > 0 && via[0].URL.Scheme == "https" && req.URL.Scheme == "http" {
				return errors.New("下载重定向不允许从 HTTPS 降级到 HTTP")
			}
			if err := validateDownloadURL(req.URL.String()); err != nil {
				return err
			}
			return nil
		},
	}

	req, err := http.NewRequest(http.MethodGet, presignedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("构造下载请求失败: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("下载请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		preview, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("下载失败: HTTP %d, body: %s", resp.StatusCode, string(preview))
	}

	// 决定文件名
	filename := opts.Filename
	if filename == "" {
		filename = resolveFilenameFromResponse(resp, defaultFallbackName)
	}
	filename = sanitizeFilename(filename)
	if filename == "" {
		filename = sanitizeFilename(defaultFallbackName)
	}
	if filename == "" {
		filename = "download"
	}

	// 输出目录
	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = "."
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("创建输出目录失败: %w", err)
	}

	outputPath := filepath.Join(outputDir, filename)

	// 覆盖检查
	if _, statErr := os.Stat(outputPath); statErr == nil && !opts.Overwrite {
		return nil, fmt.Errorf("文件已存在: %s（使用 --overwrite 覆盖）", outputPath)
	}

	// 流式写文件
	f, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("创建输出文件失败: %w", err)
	}

	size, copyErr := io.Copy(f, resp.Body)
	if closeErr := f.Close(); closeErr != nil && copyErr == nil {
		copyErr = closeErr
	}
	if copyErr != nil {
		// 清理部分文件
		_ = os.Remove(outputPath)
		return nil, fmt.Errorf("写入下载文件失败: %w", copyErr)
	}

	return &DownloadResult{
		SavedPath: outputPath,
		Size:      size,
		Filename:  filename,
	}, nil
}

// validateDownloadURL 校验下载 URL 避免 SSRF
// 拒绝非 http/https scheme、本地回环、内网 IP、链路本地等
func validateDownloadURL(rawURL string) error {
	if rawURL == "" {
		return errors.New("下载 URL 为空")
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("解析下载 URL 失败: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("下载 URL scheme 非法: %q", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return errors.New("下载 URL 缺少 host")
	}

	// 基于主机名的黑名单
	lowered := strings.ToLower(host)
	if lowered == "localhost" || strings.HasSuffix(lowered, ".localhost") {
		return errors.New("下载 URL 不允许指向 localhost")
	}

	// 尝试解析 IP（host 可能本身就是 IP 字面量）
	if ip := net.ParseIP(host); ip != nil {
		if isBlockedIP(ip) {
			return fmt.Errorf("下载 URL 指向受限 IP: %s", ip.String())
		}
	}
	return nil
}

// isBlockedIP 判断 IP 是否属于内网/本地/链路本地等受限段
func isBlockedIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() || ip.IsUnspecified() || ip.IsPrivate() {
		return true
	}
	return false
}

// resolveFilenameFromResponse 从响应头解析文件名
// 优先级：Content-Disposition filename > Content-Type 扩展推导 > defaultFallback
func resolveFilenameFromResponse(resp *http.Response, defaultFallback string) string {
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if _, params, err := mime.ParseMediaType(cd); err == nil {
			// RFC 5987 兼容：优先 filename*，再 filename
			if v, ok := params["filename*"]; ok && v != "" {
				if decoded, ok := decodeRFC5987(v); ok && decoded != "" {
					return decoded
				}
			}
			if v, ok := params["filename"]; ok && v != "" {
				return v
			}
		}
	}

	// 从 Content-Type 推导扩展名
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		mediaType, _, err := mime.ParseMediaType(ct)
		if err == nil {
			ext := extFromMediaType(mediaType)
			if ext != "" {
				return defaultFallback + ext
			}
		}
	}

	return defaultFallback + ".media"
}

// extFromMediaType 根据 MIME 推导文件扩展名，优先用常见映射，再回退到 mime.ExtensionsByType
func extFromMediaType(mediaType string) string {
	switch strings.ToLower(mediaType) {
	case "video/mp4":
		return ".mp4"
	case "video/quicktime":
		return ".mov"
	case "video/webm":
		return ".webm"
	case "audio/mp4", "audio/x-m4a":
		return ".m4a"
	case "audio/mpeg":
		return ".mp3"
	case "audio/wav", "audio/x-wav":
		return ".wav"
	case "audio/ogg":
		return ".ogg"
	case "application/pdf":
		return ".pdf"
	case "application/zip":
		return ".zip"
	case "text/plain":
		return ".txt"
	}
	if exts, err := mime.ExtensionsByType(mediaType); err == nil && len(exts) > 0 {
		return exts[0]
	}
	return ""
}

// sanitizeFilename 清洗文件名，剥除路径分隔符和危险字符
func sanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	name = filepath.Base(name)
	if name == "." || name == ".." || name == "/" || name == `\` {
		return ""
	}
	// 替换不安全字符
	replaced := strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|', '\x00':
			return '_'
		}
		return r
	}, name)
	// 长度兜底
	if len(replaced) > 200 {
		replaced = replaced[:200]
	}
	return replaced
}

// decodeRFC5987 解码 RFC 5987 格式的 filename*（UTF-8”<encoded>）
func decodeRFC5987(v string) (string, bool) {
	// 形如 UTF-8''%E4%B8%AD%E6%96%87.mp4
	_, encoded, ok := strings.Cut(v, "''")
	if !ok {
		return "", false
	}
	decoded, err := url.QueryUnescape(encoded)
	if err != nil {
		return "", false
	}
	return decoded, true
}
