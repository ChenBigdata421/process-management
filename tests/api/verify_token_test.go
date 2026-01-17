package api_tests

import (
	"fmt"
	"testing"
)

// TestGenerateToken 验证 JWT token 生成是否正确
func TestGenerateToken(t *testing.T) {
	token := GenerateTestToken(1, 1, "admin", "系统管理员", 1)
	fmt.Printf("生成的 Token: %s\n", token)

	if token == "" {
		t.Fatal("Token 生成失败")
	}

	if len(token) < 10 {
		t.Fatalf("Token 长度过短: %d", len(token))
	}

	fmt.Printf("✅ Token 生成成功，长度: %d\n", len(token))
}
