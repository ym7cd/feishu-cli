package registry

import (
	_ "embed"
	"encoding/json"
)

//go:embed service_descriptions.json
var serviceDescJSON []byte

type serviceDescLocale struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type serviceDescEntry struct {
	En         serviceDescLocale `json:"en"`
	Zh         serviceDescLocale `json:"zh"`
	AuthDomain string            `json:"auth_domain,omitempty"`
}

var serviceDescMap map[string]serviceDescEntry

func loadServiceDescriptions() map[string]serviceDescEntry {
	if serviceDescMap != nil {
		return serviceDescMap
	}
	serviceDescMap = make(map[string]serviceDescEntry)
	_ = json.Unmarshal(serviceDescJSON, &serviceDescMap)
	return serviceDescMap
}

func getServiceLocale(name, lang string) *serviceDescLocale {
	m := loadServiceDescriptions()
	entry, ok := m[name]
	if !ok {
		return nil
	}
	if lang == "en" {
		return &entry.En
	}
	return &entry.Zh
}

// GetServiceDescription returns the localized description for a service domain.
func GetServiceDescription(name, lang string) string {
	loc := getServiceLocale(name, lang)
	if loc == nil {
		return ""
	}
	return loc.Description
}

// GetServiceTitle returns the localized title for a service domain.
func GetServiceTitle(name, lang string) string {
	loc := getServiceLocale(name, lang)
	if loc == nil {
		return ""
	}
	return loc.Title
}

// GetAuthDomain returns the auth_domain for a service, or "" if not set.
func GetAuthDomain(service string) string {
	m := loadServiceDescriptions()
	if entry, ok := m[service]; ok {
		return entry.AuthDomain
	}
	return ""
}

// HasAuthDomain reports whether the service has an auth_domain configured.
func HasAuthDomain(service string) bool {
	return GetAuthDomain(service) != ""
}

// GetAuthChildren returns all service names whose auth_domain equals parent.
func GetAuthChildren(parent string) []string {
	m := loadServiceDescriptions()
	var children []string
	for name, entry := range m {
		if entry.AuthDomain == parent {
			children = append(children, name)
		}
	}
	return children
}
