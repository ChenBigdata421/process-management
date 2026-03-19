package api_tests

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var client *http.Client
var baseURL string
var token string

// 全局测试数据
var GlobalTestData struct {
	AdminUserId int
	AdminToken  string
	TestOrgId   int
	TestUserId  int
}

var testStartTime time.Time

func TestApi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Process Management API Suite")
}

var _ = BeforeSuite(func() {
	testStartTime = time.Now()
	fmt.Printf("🕐 测试开始时间: %s\n", testStartTime.Format("2006-01-02 15:04:05"))

	// 初始化数据库连接
	var err error
	// 使用正确的数据库配置 (来自 docker-compose.yml)
	dsn := "postgres://tenant_1_process-management:123456@localhost:5436/tenant_default_process?sslmode=disable&connect_timeout=1&TimeZone=Asia/Shanghai"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		Fail("无法连接到数据库: " + err.Error())
	}

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		Fail("无法获取数据库连接: " + err.Error())
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// API 服务器地址（默认 8003，可通过 TEST_API_BASE_URL 重写）
	baseURL = "http://localhost:8003"
	if override := os.Getenv("TEST_API_BASE_URL"); override != "" {
		baseURL = override
	}
	client = &http.Client{
		Timeout: 30 * time.Second,
	}

	// 生成测试 token（使用 JWT 工具函数直接签发，不调用登录接口）
	token = GenerateTestToken(1, 1, "admin", "系统管理员", 1)

	// 设置全局测试数据
	GlobalTestData.AdminUserId = 1
	GlobalTestData.AdminToken = token
	GlobalTestData.TestOrgId = 1
	GlobalTestData.TestUserId = 1

	fmt.Printf("✅ 测试环境初始化完成\n")
	fmt.Printf("   - 数据库: PostgreSQL\n")
	fmt.Printf("   - API地址: %s\n", baseURL)
	fmt.Printf("   - 用户ID: %d\n", GlobalTestData.AdminUserId)
})

var _ = AfterSuite(func() {
	fmt.Printf("🧹 开始清理测试数据...\n")

	if db != nil {
		// 清理测试数据（根据时间戳）
		db.Exec("DELETE FROM workflow_task_history WHERE created_at >= ?", testStartTime)
		db.Exec("DELETE FROM workflow_tasks WHERE created_at >= ?", testStartTime)
		db.Exec("DELETE FROM workflow_instances WHERE created_at >= ?", testStartTime)
		db.Exec("DELETE FROM workflows WHERE created_at >= ?", testStartTime)

		fmt.Printf("✅ 测试数据清理完成\n")
	}
})
