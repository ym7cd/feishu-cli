package auth

import (
	"reflect"
	"testing"
)

func TestParseScopeDomains(t *testing.T) {
	got, err := ParseScopeDomains([]string{"search,vc", "minutes"})
	if err != nil {
		t.Fatalf("ParseScopeDomains() error = %v", err)
	}
	want := []string{"search", "vc", "minutes"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseScopeDomains() = %v, want %v", got, want)
	}
}

func TestCollectDomainScopesAll(t *testing.T) {
	// Without recommend filter, search domain should return its scopes
	got, err := CollectDomainScopes([]string{"search"}, false)
	if err != nil {
		t.Fatalf("CollectDomainScopes() error = %v", err)
	}
	want := []string{"search:docs:read", "search:message"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CollectDomainScopes(search, all) = %v, want %v", got, want)
	}
}

func TestCollectDomainScopesRecommend(t *testing.T) {
	// search scopes are now in recommend.allow, so they should be returned.
	got, err := CollectDomainScopes([]string{"search"}, true)
	if err != nil {
		t.Fatalf("CollectDomainScopes() error = %v", err)
	}
	want := []string{"search:docs:read", "search:message"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CollectDomainScopes(search, recommended) = %v, want %v", got, want)
	}
}
