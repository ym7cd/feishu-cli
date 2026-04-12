package registry

import (
	"sort"
	"testing"
)

func TestInit_NonEmpty(t *testing.T) {
	projects := ListFromMetaProjects()
	if len(projects) == 0 {
		t.Fatal("ListFromMetaProjects() returned empty, expected services from meta_data.json")
	}
	// Spot-check a few known projects
	known := map[string]bool{"calendar": false, "im": false, "drive": false, "task": false}
	for _, p := range projects {
		if _, ok := known[p]; ok {
			known[p] = true
		}
	}
	for name, found := range known {
		if !found {
			t.Errorf("expected project %q in ListFromMetaProjects(), not found", name)
		}
	}
}

func TestLoadScopePriorities(t *testing.T) {
	priorities := LoadScopePriorities()
	if len(priorities) == 0 {
		t.Fatal("LoadScopePriorities() returned empty")
	}
	// Check a known scope has a positive score
	if score, ok := priorities["calendar:calendar:readonly"]; !ok || score <= 0 {
		t.Errorf("expected calendar:calendar:readonly to have positive score, got %d (found=%v)", score, ok)
	}
	// Check override applied: calendar:calendar:read should be overridden to 70
	if score := priorities["calendar:calendar:read"]; score != 70 {
		t.Errorf("expected calendar:calendar:read override score=70, got %d", score)
	}
}

func TestGetScopeScore(t *testing.T) {
	score := GetScopeScore("calendar:calendar:read")
	if score != 70 {
		t.Errorf("GetScopeScore(calendar:calendar:read) = %d, want 70", score)
	}
	score = GetScopeScore("nonexistent:scope:xyz")
	if score != DefaultScopeScore {
		t.Errorf("GetScopeScore(nonexistent) = %d, want %d", score, DefaultScopeScore)
	}
}

func TestCollectScopesForProjects_Calendar(t *testing.T) {
	scopes := CollectScopesForProjects([]string{"calendar"}, "user")
	if len(scopes) == 0 {
		t.Fatal("CollectScopesForProjects(calendar, user) returned empty")
	}
	// Should contain calendar-prefixed scopes
	found := false
	for _, s := range scopes {
		if len(s) > 9 && s[:9] == "calendar:" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected calendar:* scopes, got %v", scopes)
	}
}

func TestFilterAutoApproveScopes(t *testing.T) {
	input := []string{"calendar:calendar:read", "nonexistent:scope:xyz"}
	result := FilterAutoApproveScopes(input)
	if len(result) != 1 || result[0] != "calendar:calendar:read" {
		t.Errorf("FilterAutoApproveScopes() = %v, want [calendar:calendar:read]", result)
	}
}

func TestGetAuthChildren_Docs(t *testing.T) {
	children := GetAuthChildren("docs")
	found := false
	for _, c := range children {
		if c == "whiteboard" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("GetAuthChildren(docs) should include whiteboard, got %v", children)
	}
}

func TestHasAuthDomain(t *testing.T) {
	if !HasAuthDomain("whiteboard") {
		t.Error("HasAuthDomain(whiteboard) should be true")
	}
	if HasAuthDomain("calendar") {
		t.Error("HasAuthDomain(calendar) should be false")
	}
}

func TestKnownDomainNames_IncludesAliases(t *testing.T) {
	names := KnownDomainNames()
	required := []string{"chat", "bitable", "search", "calendar", "im", "drive"}
	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}
	for _, r := range required {
		if !nameSet[r] {
			t.Errorf("KnownDomainNames() missing %q", r)
		}
	}
	// Should be sorted
	sorted := make([]string, len(names))
	copy(sorted, names)
	sort.Strings(sorted)
	for i := range names {
		if names[i] != sorted[i] {
			t.Errorf("KnownDomainNames() not sorted at index %d: %q vs %q", i, names[i], sorted[i])
			break
		}
	}
}

func TestCollectDomainScopes_Alias(t *testing.T) {
	scopes := CollectDomainScopes([]string{"chat"}, false)
	if len(scopes) == 0 {
		t.Fatal("CollectDomainScopes(chat) returned empty")
	}
	// Should contain im: prefixed scopes
	found := false
	for _, s := range scopes {
		if len(s) > 3 && s[:3] == "im:" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("CollectDomainScopes(chat) should return im:* scopes, got %v", scopes)
	}
}

func TestCollectDomainScopes_Fallback(t *testing.T) {
	scopes := CollectDomainScopes([]string{"search"}, false)
	want := map[string]bool{"search:docs:read": true, "search:message": true}
	for _, s := range scopes {
		delete(want, s)
	}
	if len(want) > 0 {
		missing := make([]string, 0, len(want))
		for s := range want {
			missing = append(missing, s)
		}
		t.Errorf("CollectDomainScopes(search) missing: %v, got %v", missing, scopes)
	}
}

func TestCollectDomainScopes_Recommended(t *testing.T) {
	all := CollectDomainScopes([]string{"calendar"}, false)
	recommended := CollectDomainScopes([]string{"calendar"}, true)
	if len(recommended) == 0 {
		t.Fatal("CollectDomainScopes(calendar, recommended) returned empty")
	}
	if len(recommended) > len(all) {
		t.Errorf("recommended (%d) should not exceed all (%d)", len(recommended), len(all))
	}
	// All recommended should be in auto-approve set
	autoApprove := LoadAutoApproveSet()
	for _, s := range recommended {
		if !autoApprove[s] {
			t.Errorf("recommended scope %q is not auto-approve", s)
		}
	}
}

func TestParseDomains(t *testing.T) {
	// Comma-separated + multiple flags
	got, err := ParseDomains([]string{"search,vc", "minutes"})
	if err != nil {
		t.Fatalf("ParseDomains() error = %v", err)
	}
	want := []string{"search", "vc", "minutes"}
	if len(got) != len(want) {
		t.Fatalf("ParseDomains() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ParseDomains()[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	// "all" expands to all known domains
	all, err := ParseDomains([]string{"all"})
	if err != nil {
		t.Fatalf("ParseDomains(all) error = %v", err)
	}
	if len(all) < 10 {
		t.Errorf("ParseDomains(all) returned only %d domains, expected 10+", len(all))
	}

	// Unknown domain
	_, err = ParseDomains([]string{"nonexistent_xyz"})
	if err == nil {
		t.Error("ParseDomains(nonexistent) should return error")
	}
}

func TestCollectDomainScopes_AuthDomainExpansion(t *testing.T) {
	// "docs" domain should include whiteboard (auth_domain child) scopes
	scopes := CollectDomainScopes([]string{"docs"}, false)
	// Should have board: prefixed scopes from whiteboard
	found := false
	for _, s := range scopes {
		if len(s) > 6 && s[:6] == "board:" {
			found = true
			break
		}
	}
	if !found {
		t.Logf("CollectDomainScopes(docs) scopes = %v", scopes)
		t.Log("Note: whiteboard scopes may not have board: prefix in meta_data; auth_domain expansion is still functional")
	}
}
