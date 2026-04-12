package registry

// GetStrFromMap extracts a string value from map[string]interface{}.
func GetStrFromMap(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetStrSliceFromMap extracts a []string value from map[string]interface{}.
func GetStrSliceFromMap(m map[string]interface{}, key string) []string {
	if m == nil {
		return nil
	}
	raw, ok := m[key].([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// SelectRecommendedScope selects the scope with the highest priority score.
// Scopes not in the priority table are skipped; falls back to the first scope.
func SelectRecommendedScope(scopes []interface{}) string {
	priorities := LoadScopePriorities()
	bestScore := -1
	bestScope := ""
	for _, s := range scopes {
		str, ok := s.(string)
		if !ok {
			continue
		}
		score, exists := priorities[str]
		if !exists {
			continue
		}
		if score > bestScore {
			bestScore = score
			bestScope = str
		}
	}
	if bestScope != "" {
		return bestScope
	}
	if len(scopes) > 0 {
		if s, ok := scopes[0].(string); ok {
			return s
		}
	}
	return ""
}
