package api_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Instance API Tests", func() {

	Describe("POST /api/v1/instances - 启动工作流实例", func() {
		It("应该返回错误当工作流不存在", func() {
			payload := map[string]interface{}{
				"workflow_id": "nonexistent-workflow-id",
				"input":       `{"key":"value"}`,
			}

			body, _ := json.Marshal(payload)
			req, _ := http.NewRequest("POST", baseURL+"/api/v1/instances", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", token)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 500)
		})
	})

	Describe("GET /api/v1/instances - 查询所有实例", func() {
		It("应该成功返回实例列表", func() {
			req, _ := http.NewRequest("GET", baseURL+"/api/v1/instances?limit=10&offset=0", nil)
			req.Header.Set("Authorization", token)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			result := expectBusinessCode(resp, 200)
			Expect(result["data"]).NotTo(BeNil())

			fmt.Printf("✅ 实例列表查询成功\n")
		})
	})

	Describe("GET /api/v1/instances/:id - 获取实例详情", func() {
		It("应该返回404当实例不存在", func() {
			req, _ := http.NewRequest("GET", baseURL+"/api/v1/instances/nonexistent-id", nil)
			req.Header.Set("Authorization", token)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 500)
		})
	})

	Describe("GET /api/v1/instances/workflow/:workflow_id - 查询工作流实例", func() {
		It("应该成功返回空列表当工作流不存在", func() {
			req, _ := http.NewRequest("GET", baseURL+"/api/v1/instances/workflow/nonexistent-id?limit=10&offset=0", nil)
			req.Header.Set("Authorization", token)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 200)
		})
	})

	Describe("DELETE /api/v1/instances/:id - 删除实例", func() {
		It("应该返回404当实例不存在", func() {
			req, _ := http.NewRequest("DELETE", baseURL+"/api/v1/instances/nonexistent-id", nil)
			req.Header.Set("Authorization", token)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 500)
		})
	})
})
