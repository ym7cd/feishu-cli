package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/riba2534/feishu-cli/internal/auth"
)

type loginScopeSummary struct {
	Requested []string
	Granted   []string
	Missing   []string
}

func buildLoginScopeSummary(requestedScope, grantedScope string) *loginScopeSummary {
	requested := auth.UniqueScopeList(requestedScope)
	if len(requested) == 0 {
		return &loginScopeSummary{
			Requested: []string{},
			Granted:   auth.UniqueScopeList(grantedScope),
			Missing:   []string{},
		}
	}
	granted, missing := auth.PartitionScopes(grantedScope, requested)
	return &loginScopeSummary{
		Requested: requested,
		Granted:   granted,
		Missing:   missing,
	}
}

func buildAuthorizationCompleteEvent(token *auth.TokenStore, summary *loginScopeSummary) map[string]any {
	event := map[string]any{
		"event":            "authorization_complete",
		"expires_at":       token.ExpiresAt.Format("2006-01-02T15:04:05+08:00"),
		"scope":            token.Scope,
		"requested_scopes": emptyIfNil(summary.Requested),
		"granted_scopes":   emptyIfNil(summary.Granted),
		"missing_scopes":   emptyIfNil(summary.Missing),
	}
	if !token.RefreshExpiresAt.IsZero() {
		event["refresh_expires_at"] = token.RefreshExpiresAt.Format("2006-01-02T15:04:05+08:00")
	}
	if len(summary.Requested) > 0 {
		event["requested_scope"] = strings.Join(summary.Requested, " ")
	}
	if len(summary.Missing) > 0 {
		event["warning"] = fmt.Sprintf("以下 scope 未授予: %s", strings.Join(summary.Missing, " "))
		event["hint"] = fmt.Sprintf("确认飞书开放平台已开通后，执行 feishu-cli auth login --scope %q 重新授权", strings.Join(summary.Missing, " "))
	}
	return event
}

func printTokenSuccess(token *auth.TokenStore, summary *loginScopeSummary) {
	path, _ := auth.TokenPath()
	fmt.Fprintln(os.Stderr, "\n✓ 授权成功！")
	fmt.Fprintf(os.Stderr, "  Token 已保存到 %s\n", path)
	fmt.Fprintf(os.Stderr, "  Access Token 有效期至: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05"))
	if !token.RefreshExpiresAt.IsZero() {
		fmt.Fprintf(os.Stderr, "  Refresh Token 有效期至: %s\n", token.RefreshExpiresAt.Format("2006-01-02 15:04:05"))
	}
	if token.Scope != "" {
		fmt.Fprintf(os.Stderr, "  实际授予: %s\n", token.Scope)
	}
	if summary != nil && len(summary.Requested) > 0 {
		fmt.Fprintf(os.Stderr, "  请求范围: %s\n", strings.Join(summary.Requested, " "))
	}
	if summary != nil && len(summary.Missing) > 0 {
		fmt.Fprintf(os.Stderr, "  未授予:   %s\n", strings.Join(summary.Missing, " "))
		fmt.Fprintf(os.Stderr, "  提示:     确认飞书开放平台已开通后，执行 feishu-cli auth login --scope %q 重新授权\n", strings.Join(summary.Missing, " "))
	}
}

func emptyIfNil(items []string) []string {
	if items == nil {
		return []string{}
	}
	return items
}
