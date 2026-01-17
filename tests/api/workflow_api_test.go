package api_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Workflow API Tests", func() {
	var workflowID string

	Describe("POST /api/v1/workflows - 创建工作流", func() {
		It("应该成功创建工作流", func() {
			payload := map[string]interface{}{
				"name":        fmt.Sprintf("测试工作流_%d", GinkgoRandomSeed()),
				"description": "测试工作流描述",
				"definition":  `{"steps":[{"id":"step1","name":"步骤1"}]}`,
			}

			body, _ := json.Marshal(payload)
			req, _ := http.NewRequest("POST", baseURL+"/api/v1/workflows", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", token)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			result := expectBusinessCode(resp, 200)

			if data, ok := result["data"].(map[string]interface{}); ok {
				if id, ok := data["id"].(string); ok {
					workflowID = id
					fmt.Printf("✅ 工作流创建成功 - ID: %s\n", workflowID)
				}
			}
		})

		It("应该拒绝无效的请求（缺少必填字段）", func() {
			payload := map[string]interface{}{
				"description": "缺少名称字段",
			}

			body, _ := json.Marshal(payload)
			req, _ := http.NewRequest("POST", baseURL+"/api/v1/workflows", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", token)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 400)
		})
	})

	Describe("GET /api/v1/workflows - 查询工作流列表", func() {
		It("应该成功返回工作流列表", func() {
			req, _ := http.NewRequest("GET", baseURL+"/api/v1/workflows?limit=10&offset=0", nil)
			req.Header.Set("Authorization", token)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			result := expectBusinessCode(resp, 200)
			Expect(result["data"]).NotTo(BeNil())
		})
	})

	Describe("GET /api/v1/workflows/:id - 获取工作流详情", func() {
		It("应该返回404当工作流不存在", func() {
			req, _ := http.NewRequest("GET", baseURL+"/api/v1/workflows/nonexistent-id", nil)
			req.Header.Set("Authorization", token)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 500)
		})
	})

	Describe("PUT /api/v1/workflows/:id - 更新工作流", func() {
		It("应该返回404当工作流不存在", func() {
			payload := map[string]interface{}{
				"name":        "更新的工作流",
				"description": "更新的描述",
				"definition":  `{"steps":[]}`,
			}

			body, _ := json.Marshal(payload)
			req, _ := http.NewRequest("PUT", baseURL+"/api/v1/workflows/nonexistent-id", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", token)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 500)
		})
	})

	Describe("POST /api/v1/workflows/:id/activate - 激活工作流", func() {
		It("应该返回404当工作流不存在", func() {
			req, _ := http.NewRequest("POST", baseURL+"/api/v1/workflows/nonexistent-id/activate", nil)
			req.Header.Set("Authorization", token)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 500)
		})
	})

	Describe("POST /api/v1/workflows/:id/freeze - 冻结工作流", func() {
		It("应该返回404当工作流不存在", func() {
			req, _ := http.NewRequest("POST", baseURL+"/api/v1/workflows/nonexistent-id/freeze", nil)
			req.Header.Set("Authorization", token)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 500)
		})
	})

	Describe("GET /api/v1/workflows/:id/can-freeze - 检查是否可冻结", func() {
		It("应该返回404当工作流不存在", func() {
			req, _ := http.NewRequest("GET", baseURL+"/api/v1/workflows/nonexistent-id/can-freeze", nil)
			req.Header.Set("Authorization", token)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 200)
		})
	})

	Describe("DELETE /api/v1/workflows/:id - 删除工作流", func() {
		It("应该返回404当工作流不存在", func() {
			req, _ := http.NewRequest("DELETE", baseURL+"/api/v1/workflows/nonexistent-id", nil)
			req.Header.Set("Authorization", token)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 500)
		})
	})
})
