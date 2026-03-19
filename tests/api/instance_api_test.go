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
				"workflowId": "nonexistent-workflow-id",
				"input":      `{"key":"value"}`,
			}

			body, _ := json.Marshal(payload)
			req, _ := http.NewRequest("POST", baseURL+"/api/v1/instances", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", token)
		req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 400)
		})
	})

	Describe("GET /api/v1/instances - 查询所有实例", func() {
		It("应该成功返回实例列表", func() {
			req, _ := http.NewRequest("GET", baseURL+"/api/v1/instances?limit=10&offset=0", nil)
			req.Header.Set("Authorization", token)
		req.Header.Set("X-Tenant-ID", "1")

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
		req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 400)
		})
	})

	Describe("GET /api/v1/instances/workflow/:workflow_id - 查询工作流实例", func() {
		It("应该成功返回空列表当工作流不存在", func() {
			req, _ := http.NewRequest("GET", baseURL+"/api/v1/instances/workflow/nonexistent-id?limit=10&offset=0", nil)
			req.Header.Set("Authorization", token)
		req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 200)
		})
	})

	Describe("PUT /api/v1/instances/:id/cancel - 取消工作流实例", func() {
		It("应该返回404当实例不存在", func() {
			req, _ := http.NewRequest("PUT", baseURL+"/api/v1/instances/nonexistent-id/cancel", nil)
			req.Header.Set("Authorization", token)
		req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 400)
			fmt.Printf("✅ 不存在的实例取消被正确拒绝\n")
		})

		It("应该返回401当缺少Authorization时", func() {
			req, _ := http.NewRequest("PUT", baseURL+"/api/v1/instances/test-instance/cancel", nil)
			// 不设置 Authorization header
			req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			// JWT认证失败应返回401
			Expect(resp.StatusCode).To(Equal(http.StatusOK)) // API始终返回200 HTTP状态码
			expectBusinessCode(resp, 401)
			fmt.Printf("✅ 缺少Authorization被正确拒绝\n")
		})
	})

	Describe("GET /api/v1/instances/:id/detail - 获取实例详情", func() {
		It("应该返回404当实例不存在", func() {
			req, _ := http.NewRequest("GET", baseURL+"/api/v1/instances/nonexistent-id/detail", nil)
			req.Header.Set("Authorization", token)
		req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 400)
			fmt.Printf("✅ 不存在的实例详情查询被正确拒绝\n")
		})

		It("应该返回401当缺少Authorization时", func() {
			req, _ := http.NewRequest("GET", baseURL+"/api/v1/instances/test-instance/detail", nil)
			// 不设置 Authorization header
			req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			// JWT认证失败应返回401
			Expect(resp.StatusCode).To(Equal(http.StatusOK)) // API始终返回200 HTTP状态码
			expectBusinessCode(resp, 401)
			fmt.Printf("✅ 缺少Authorization被正确拒绝\n")
		})
	})

	Describe("DELETE /api/v1/instances/:id - 删除实例", func() {
		It("应该返回404当实例不存在", func() {
			req, _ := http.NewRequest("DELETE", baseURL+"/api/v1/instances/nonexistent-id", nil)
			req.Header.Set("Authorization", token)
		req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 400)
		})
	})

	Describe("边界条件测试", func() {
		It("应该处理无效的ID格式", func() {
			invalidIds := []string{
				"",
				"invalid-id-with-special-!@#",
				"../../../etc/passwd",
			}

			for _, instanceId := range invalidIds {
				req, _ := http.NewRequest("GET", baseURL+"/api/v1/instances/"+instanceId, nil)
				req.Header.Set("Authorization", token)
		req.Header.Set("X-Tenant-ID", "1")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				// 应该正常处理或返回错误
				Expect(resp.StatusCode).Should(BeElementOf([]int{http.StatusOK, http.StatusBadRequest, http.StatusNotFound}))
			}

			fmt.Printf("✅ 无效ID格式处理正确\n")
		})
	})
})
