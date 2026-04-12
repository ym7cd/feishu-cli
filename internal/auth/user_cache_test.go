package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadAndDeleteCurrentUserCache(t *testing.T) {
	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "user_profile.json")

	original := userCachePathFunc
	userCachePathFunc = func() (string, error) { return cacheFile, nil }
	defer func() { userCachePathFunc = original }()

	cache := &CurrentUserCache{
		OpenID:           "ou_test",
		UserID:           "u_test",
		UnionID:          "on_test",
		Name:             "Tester",
		TokenFingerprint: UserTokenFingerprint("access-token"),
	}

	if err := SaveCurrentUserCache(cache); err != nil {
		t.Fatalf("SaveCurrentUserCache() error: %v", err)
	}

	info, err := os.Stat(cacheFile)
	if err != nil {
		t.Fatalf("Stat() error: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Fatalf("cache file permissions = %o, want 0600", perm)
	}

	loaded, err := LoadCurrentUserCache()
	if err != nil {
		t.Fatalf("LoadCurrentUserCache() error: %v", err)
	}
	if loaded.OpenID != cache.OpenID {
		t.Fatalf("OpenID = %q, want %q", loaded.OpenID, cache.OpenID)
	}
	if !loaded.MatchesToken("access-token") {
		t.Fatalf("MatchesToken(access-token) = false, want true")
	}
	if loaded.MatchesToken("another-token") {
		t.Fatalf("MatchesToken(another-token) = true, want false")
	}

	if err := DeleteCurrentUserCache(); err != nil {
		t.Fatalf("DeleteCurrentUserCache() error: %v", err)
	}

	loaded, err = LoadCurrentUserCache()
	if err != nil {
		t.Fatalf("LoadCurrentUserCache() after delete error: %v", err)
	}
	if loaded != nil {
		t.Fatalf("expected nil cache after delete")
	}
}

func TestSaveTokenDeletesCurrentUserCache(t *testing.T) {
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "token.json")
	cacheFile := filepath.Join(tmpDir, "user_profile.json")

	originalTokenPath := tokenPathFunc
	originalUserCachePath := userCachePathFunc
	tokenPathFunc = func() (string, error) { return tokenFile, nil }
	userCachePathFunc = func() (string, error) { return cacheFile, nil }
	defer func() {
		tokenPathFunc = originalTokenPath
		userCachePathFunc = originalUserCachePath
	}()

	if err := SaveCurrentUserCache(&CurrentUserCache{OpenID: "ou_test"}); err != nil {
		t.Fatalf("SaveCurrentUserCache() setup error: %v", err)
	}

	if err := SaveToken(&TokenStore{AccessToken: "access"}); err != nil {
		t.Fatalf("SaveToken() error: %v", err)
	}

	cache, err := LoadCurrentUserCache()
	if err != nil {
		t.Fatalf("LoadCurrentUserCache() error: %v", err)
	}
	if cache != nil {
		t.Fatalf("expected current user cache to be cleared after SaveToken")
	}
}

func TestDeleteTokenDeletesCurrentUserCache(t *testing.T) {
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "token.json")
	cacheFile := filepath.Join(tmpDir, "user_profile.json")

	originalTokenPath := tokenPathFunc
	originalUserCachePath := userCachePathFunc
	tokenPathFunc = func() (string, error) { return tokenFile, nil }
	userCachePathFunc = func() (string, error) { return cacheFile, nil }
	defer func() {
		tokenPathFunc = originalTokenPath
		userCachePathFunc = originalUserCachePath
	}()

	if err := os.WriteFile(tokenFile, []byte(`{"access_token":"access"}`), 0600); err != nil {
		t.Fatalf("write token file: %v", err)
	}
	if err := SaveCurrentUserCache(&CurrentUserCache{OpenID: "ou_test"}); err != nil {
		t.Fatalf("SaveCurrentUserCache() setup error: %v", err)
	}

	if err := DeleteToken(); err != nil {
		t.Fatalf("DeleteToken() error: %v", err)
	}

	cache, err := LoadCurrentUserCache()
	if err != nil {
		t.Fatalf("LoadCurrentUserCache() error: %v", err)
	}
	if cache != nil {
		t.Fatalf("expected current user cache to be cleared after DeleteToken")
	}
}

func TestSaveTokenIgnoresCurrentUserCacheDeleteError(t *testing.T) {
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "token.json")
	cachePath := nonRemovableCachePath(t)

	originalTokenPath := tokenPathFunc
	originalUserCachePath := userCachePathFunc
	tokenPathFunc = func() (string, error) { return tokenFile, nil }
	userCachePathFunc = func() (string, error) { return cachePath, nil }
	defer func() {
		tokenPathFunc = originalTokenPath
		userCachePathFunc = originalUserCachePath
	}()

	if err := SaveToken(&TokenStore{AccessToken: "access"}); err != nil {
		t.Fatalf("SaveToken() error = %v, want nil when cache cleanup fails", err)
	}

	if _, err := os.Stat(tokenFile); err != nil {
		t.Fatalf("token file not written: %v", err)
	}
}

func TestDeleteTokenIgnoresCurrentUserCacheDeleteError(t *testing.T) {
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "token.json")
	cachePath := nonRemovableCachePath(t)

	originalTokenPath := tokenPathFunc
	originalUserCachePath := userCachePathFunc
	tokenPathFunc = func() (string, error) { return tokenFile, nil }
	userCachePathFunc = func() (string, error) { return cachePath, nil }
	defer func() {
		tokenPathFunc = originalTokenPath
		userCachePathFunc = originalUserCachePath
	}()

	if err := os.WriteFile(tokenFile, []byte(`{"access_token":"access"}`), 0600); err != nil {
		t.Fatalf("write token file: %v", err)
	}

	if err := DeleteToken(); err != nil {
		t.Fatalf("DeleteToken() error = %v, want nil when cache cleanup fails", err)
	}

	if _, err := os.Stat(tokenFile); !os.IsNotExist(err) {
		t.Fatalf("token file should be removed, got err = %v", err)
	}
}

func nonRemovableCachePath(t *testing.T) string {
	t.Helper()

	cacheDir := filepath.Join(t.TempDir(), "user_profile.json")
	if err := os.Mkdir(cacheDir, 0700); err != nil {
		t.Fatalf("mkdir cache dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, "child"), []byte("busy"), 0600); err != nil {
		t.Fatalf("write cache dir child: %v", err)
	}

	return cacheDir
}
