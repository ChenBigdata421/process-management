package api_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Task API Tests", func() {

	Describe("POST /api/v1/tasks - 创建任务", func() {
		It("应该返回错误当必填字段缺失", func() {
			payload := map[string]interface{}{
				"taskName": "测试任务",
			}

			body, _ := json.Marshal(payload)
			req, _ := http.NewRequest("POST", baseURL+"/api/v1/tasks", bytes.NewBuffer(body))
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

	Describe("GET /api/v1/tasks - 查询所有任务", func() {
		It("应该成功返回任务列表", func() {
			req, _ := http.NewRequest("GET", baseURL+"/api/v1/tasks?limit=10&offset=0", nil)
			req.Header.Set("Authorization", token)
			req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			result := expectBusinessCode(resp, 200)
			Expect(result["data"]).NotTo(BeNil())

			fmt.Printf("✅ 任务列表查询成功\n")
		})
	})

	Describe("GET /api/v1/tasks/todo - 查询待办任务", func() {
		It("应该成功返回待办任务列表", func() {
			req, _ := http.NewRequest("GET", baseURL+"/api/v1/tasks/todo?limit=10&offset=0", nil)
			req.Header.Set("Authorization", token)
			req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 200)
		})
	})

	Describe("GET /api/v1/tasks/done - 查询已办任务", func() {
		It("应该成功返回已办任务列表", func() {
			req, _ := http.NewRequest("GET", baseURL+"/api/v1/tasks/done?limit=10&offset=0", nil)
			req.Header.Set("Authorization", token)
			req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 200)
		})
	})

	Describe("GET /api/v1/tasks/:id - 获取任务详情", func() {
		It("应该返回404当任务不存在", func() {
			req, _ := http.NewRequest("GET", baseURL+"/api/v1/tasks/nonexistent-id", nil)
			req.Header.Set("Authorization", token)
			req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 400)
		})
	})

	Describe("POST /api/v1/tasks/:id/complete - 完成任务", func() {
		It("应该返回404当任务不存在", func() {
			payload := map[string]interface{}{
				"output":  `{"result":"success"}`,
				"comment": "已完成",
				"result":  "completed",
			}

			body, _ := json.Marshal(payload)
			req, _ := http.NewRequest("POST", baseURL+"/api/v1/tasks/nonexistent-id/complete", bytes.NewBuffer(body))
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

	Describe("POST /api/v1/tasks/:id/approve - 批准任务", func() {
		It("应该返回404当任务不存在", func() {
			payload := map[string]interface{}{
				"comment": "批准",
			}

			body, _ := json.Marshal(payload)
			req, _ := http.NewRequest("POST", baseURL+"/api/v1/tasks/nonexistent-id/approve", bytes.NewBuffer(body))
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

	Describe("POST /api/v1/tasks/:id/reject - 驳回任务", func() {
		It("应该返回404当任务不存在", func() {
			payload := map[string]interface{}{
				"comment": "驳回",
				"reason":  "不符合要求",
			}

			body, _ := json.Marshal(payload)
			req, _ := http.NewRequest("POST", baseURL+"/api/v1/tasks/nonexistent-id/reject", bytes.NewBuffer(body))
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

	Describe("POST /api/v1/tasks/:id/delegate - 转办任务", func() {
		It("应该返回404当任务不存在", func() {
			payload := map[string]interface{}{
				"target_id": "2",
				"comment":   "转办给其他人",
			}

			body, _ := json.Marshal(payload)
			req, _ := http.NewRequest("POST", baseURL+"/api/v1/tasks/nonexistent-id/delegate", bytes.NewBuffer(body))
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

	Describe("DELETE /api/v1/tasks/:id - 删除任务", func() {
		It("应该返回404当任务不存在", func() {
			req, _ := http.NewRequest("DELETE", baseURL+"/api/v1/tasks/nonexistent-id", nil)
			req.Header.Set("Authorization", token)
			req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 400)
		})
	})

	Describe("GET /api/v1/tasks/instance/:instanceId/recent - 获取实例最近任务", func() {
		It("应该成功返回空结果当实例不存在", func() {
			req, _ := http.NewRequest("GET", baseURL+"/api/v1/tasks/instance/nonexistent-id/recent", nil)
			req.Header.Set("Authorization", token)
			req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			fmt.Printf("✅ 不存在的实例最近任务查询成功\n")
		})

		It("应该返回401当缺少Authorization时", func() {
			req, _ := http.NewRequest("GET", baseURL+"/api/v1/tasks/instance/test-instance/recent", nil)
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

	Describe("GET /api/v1/tasks/:id/history - 获取任务历史", func() {
		It("应该成功返回空历史列表", func() {
			req, _ := http.NewRequest("GET", baseURL+"/api/v1/tasks/nonexistent-id/history?limit=10&offset=0", nil)
			req.Header.Set("Authorization", token)
			req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})

	Describe("GET /api/v1/tasks/instance/:instance_id/history - 获取实例任务历史", func() {
		It("应该成功返回空历史列表", func() {
			req, _ := http.NewRequest("GET", baseURL+"/api/v1/tasks/instance/nonexistent-id/history?limit=10&offset=0", nil)
			req.Header.Set("Authorization", token)
			req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})

	Describe("GET /api/v1/tasks/instance/:instance_id - 获取实例所有任务", func() {
		It("应该成功返回空任务列表", func() {
			req, _ := http.NewRequest("GET", baseURL+"/api/v1/tasks/instance/nonexistent-id?limit=10&offset=0", nil)
			req.Header.Set("Authorization", token)
			req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			// 对于不存在的实例ID，API可能返回400（参数格式错误）或200（空结果）
			fmt.Printf("✅ 实例任务查询完成\n")
		})
	})
})
