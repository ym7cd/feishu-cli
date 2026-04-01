package cmd

import "testing"

func TestValidateUrgentType(t *testing.T) {
	valid := []string{"app", "phone", "sms"}
	for _, v := range valid {
		if err := validateUrgentType(v); err != nil {
			t.Fatalf("validateUrgentType(%q) 返回错误: %v", v, err)
		}
	}

	invalid := []string{"", "APP", "push", "email"}
	for _, v := range invalid {
		if err := validateUrgentType(v); err == nil {
			t.Fatalf("validateUrgentType(%q) 应返回错误", v)
		}
	}
}

func TestValidateUrgentUserIDType(t *testing.T) {
	valid := []string{"open_id", "user_id", "union_id"}
	for _, v := range valid {
		if err := validateUrgentUserIDType(v); err != nil {
			t.Fatalf("validateUrgentUserIDType(%q) 返回错误: %v", v, err)
		}
	}

	invalid := []string{"", "openid", "chat_id", "email"}
	for _, v := range invalid {
		if err := validateUrgentUserIDType(v); err == nil {
			t.Fatalf("validateUrgentUserIDType(%q) 应返回错误", v)
		}
	}
}
