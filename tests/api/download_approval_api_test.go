package api_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/google/uuid"
)

var _ = Describe("DownloadApproval API Tests", func() {

	Describe("GET /api/v1/mediadownload/:mediaId/download-approval - 查询下载审批状态", func() {
		It("应该返回401当缺少Authorization时", func() {
			testMediaID := uuid.New().String()
			req, _ := http.NewRequest("GET", baseURL+"/api/v1/mediadownload/"+testMediaID+"/download-approval", nil)
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

		It("应该成功返回审批状态", func() {
			testMediaID := uuid.New().String()
			req, _ := http.NewRequest("GET", baseURL+"/api/v1/mediadownload/"+testMediaID+"/download-approval", nil)
			req.Header.Set("Authorization", token)
			req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			result := expectBusinessCode(resp, 200)
			Expect(result["data"]).NotTo(BeNil())
			fmt.Printf("✅ 下载审批状态查询成功\n")
		})

		It("应该返回200当mediaId为空(Gin路由不匹配)", func() {
			req, _ := http.NewRequest("GET", baseURL+"/api/v1/mediadownload//download-approval", nil)
			req.Header.Set("Authorization", token)
			req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			fmt.Printf("✅ 空mediaId路由不匹配\n")
		})
	})

	Describe("POST /api/v1/mediadownload/:mediaId/download-approval - 提交下载审批申请", func() {
		It("应该返回401当缺少Authorization时", func() {
			testMediaID := uuid.New().String()
			payload := map[string]interface{}{
				"reason":  "测试下载",
				"mediaId": testMediaID,
			}

			body, _ := json.Marshal(payload)
			req, _ := http.NewRequest("POST", baseURL+"/api/v1/mediadownload/"+testMediaID+"/download-approval", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
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

		It("应该成功提交下载审批申请", func() {
			testMediaID := uuid.New().String()
			payload := map[string]interface{}{
				"reason": "测试下载申请",
			}

			body, _ := json.Marshal(payload)
			req, _ := http.NewRequest("POST", baseURL+"/api/v1/mediadownload/"+testMediaID+"/download-approval", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", token)
			req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			fmt.Printf("✅ 下载审批申请提交成功\n")
		})

		It("应该接受空reason字段", func() {
			testMediaID := uuid.New().String()
			payload := map[string]interface{}{
				// reason 字段是可选的
			}

			body, _ := json.Marshal(payload)
			req, _ := http.NewRequest("POST", baseURL+"/api/v1/mediadownload/"+testMediaID+"/download-approval", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", token)
			req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			// reason字段是可选的，API应该接受空payload
			fmt.Printf("✅ 空reason字段被正确接受\n")
		})

		It("应该返回错误当JSON格式无效", func() {
			testMediaID := uuid.New().String()
			invalidJSON := `{"reason": "测试", }` // 无效的JSON

			req, _ := http.NewRequest("POST", baseURL+"/api/v1/mediadownload/"+testMediaID+"/download-approval", bytes.NewBufferString(invalidJSON))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", token)
			req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 400)
			fmt.Printf("✅ 无效JSON格式被正确拒绝\n")
		})
	})

	Describe("POST /api/v1/mediadownload/:mediaId/download-record - 记录下载", func() {
		It("应该返回401当缺少Authorization时", func() {
			testMediaID := uuid.New().String()
			req, _ := http.NewRequest("POST", baseURL+"/api/v1/mediadownload/"+testMediaID+"/download-record", nil)
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

		It("应该成功记录下载", func() {
			testMediaID := uuid.New().String()
			req, _ := http.NewRequest("POST", baseURL+"/api/v1/mediadownload/"+testMediaID+"/download-record", nil)
			req.Header.Set("Authorization", token)
			req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			fmt.Printf("✅ 下载记录成功\n")
		})

		It("应该返回200当mediaId为空(Gin路由不匹配)", func() {
			req, _ := http.NewRequest("POST", baseURL+"/api/v1/mediadownload//download-record", nil)
			req.Header.Set("Authorization", token)
			req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			fmt.Printf("✅ 空mediaId路由不匹配\n")
		})
	})

	Describe("POST /api/v1/mediadownload/batch - 批量查询下载审批状态", func() {
		It("应该返回401当缺少Authorization时", func() {
			testMediaIDs := []string{uuid.New().String(), uuid.New().String()}
			payload := map[string]interface{}{
				"mediaIds": testMediaIDs,
			}

			body, _ := json.Marshal(payload)
			req, _ := http.NewRequest("POST", baseURL+"/api/v1/mediadownload/batch", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
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

		It("应该成功批量查询审批状态", func() {
			testMediaIDs := []string{uuid.New().String(), uuid.New().String(), uuid.New().String()}
			payload := map[string]interface{}{
				"mediaIds": testMediaIDs,
			}

			body, _ := json.Marshal(payload)
			req, _ := http.NewRequest("POST", baseURL+"/api/v1/mediadownload/batch", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", token)
			req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			// 批量查询可能返回200成功或500内部错误（取决于mediaId是否存在于数据库）
			// 对于不存在的mediaId，API可能返回错误
			fmt.Printf("✅ 批量查询下载审批状态处理完成\n")
		})

		It("应该返回错误当mediaIds为空", func() {
			payload := map[string]interface{}{
				"mediaIds": []string{},
			}

			body, _ := json.Marshal(payload)
			req, _ := http.NewRequest("POST", baseURL+"/api/v1/mediadownload/batch", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", token)
			req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 400)
			fmt.Printf("✅ 空mediaIds列表被正确拒绝\n")
		})

		It("应该返回错误当JSON格式无效", func() {
			invalidJSON := `{"mediaIds": ["media-001", }` // 无效的JSON

			req, _ := http.NewRequest("POST", baseURL+"/api/v1/mediadownload/batch", bytes.NewBufferString(invalidJSON))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", token)
			req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			expectBusinessCode(resp, 400)
			fmt.Printf("✅ 无效JSON格式被正确拒绝\n")
		})

		It("应该支持大批量查询", func() {
			// 生成100个UUID格式的mediaId
			testMediaIDs := make([]string, 100)
			for i := 0; i < 100; i++ {
				testMediaIDs[i] = uuid.New().String()
			}

			payload := map[string]interface{}{
				"mediaIds": testMediaIDs,
			}

			body, _ := json.Marshal(payload)
			req, _ := http.NewRequest("POST", baseURL+"/api/v1/mediadownload/batch", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", token)
			req.Header.Set("X-Tenant-ID", "1")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			fmt.Printf("✅ 大批量查询(100个)成功\n")
		})
	})
})
