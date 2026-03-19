package api_tests

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	. "github.com/onsi/gomega"
)

func decodeResponseBody(resp *http.Response) map[string]interface{} {
	var body map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&body)
	Expect(err).NotTo(HaveOccurred())
	return body
}

func expectBusinessCode(resp *http.Response, expected int) map[string]interface{} {
	body := decodeResponseBody(resp)
	// 处理JSON数字解析为float64的问题
	var actualCode int
	switch v := body["code"].(type) {
	case float64:
		actualCode = int(v)
	case int:
		actualCode = v
	case int32:
		actualCode = int(v)
	case int64:
		actualCode = int(v)
	default:
		actualCode = 0
	}
	Expect(actualCode).To(BeEquivalentTo(expected))
	return body
}

func expectBusinessCodeNotEqual(resp *http.Response, unexpected int) map[string]interface{} {
	body := decodeResponseBody(resp)
	// 处理JSON数字解析为float64的问题
	var actualCode int
	switch v := body["code"].(type) {
	case float64:
		actualCode = int(v)
	case int:
		actualCode = v
	case int32:
		actualCode = int(v)
	case int64:
		actualCode = int(v)
	default:
		actualCode = 0
	}
	Expect(actualCode).NotTo(BeEquivalentTo(unexpected))
	return body
}

// generateTestMediaID 生成测试用的有效UUID格式MediaID
func generateTestMediaID() string {
	return uuid.New().String()
}

// generateTestMediaIDs 生成多个测试用的MediaID
func generateTestMediaIDs(count int) []string {
	ids := make([]string, count)
	for i := 0; i < count; i++ {
		ids[i] = generateTestMediaID()
	}
	return ids
}

// generateTestMediaIDString 生成用于URL的MediaID字符串
func generateTestMediaIDString(index int) string {
	// 生成确定的UUID用于测试
	uuid := fmt.Sprintf("0193e8d9-a1e1-7143-9223-c9e2e0fe%04d", index)
	return uuid
}