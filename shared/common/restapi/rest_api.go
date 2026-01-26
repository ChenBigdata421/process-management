package restapi

import (
	"fmt"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"

	vd "github.com/bytedance/go-tagexpr/v2/validator"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"
)

type RestApi struct{}

// GetLogger 获取上下文提供的日志器
func (e *RestApi) GetLogger(c *gin.Context) *zap.Logger {
	ctx := c.Request.Context()
	// 从上下文中获取 logger
	requestLogger, ok := ctx.Value(logger.LoggerKey).(*zap.Logger)
	if !ok {
		// 如果没有找到 logger，使用默认 logger
		requestLogger = logger.Logger
	}
	return requestLogger
}

// Bind 参数校验  jiyuanjie add for 兼容sys_job
func (e *RestApi) Bind(c *gin.Context, d interface{}, bindings ...binding.Binding) error {
	var err error
	if len(bindings) == 0 {

		bindings = constructor.GetBindingForGin(d)
	}
	for i := range bindings {
		if bindings[i] == nil {
			err = c.ShouldBindUri(d)
		} else {
			err = c.ShouldBindWith(d, bindings[i])
		}
		if err != nil && err.Error() == "EOF" {
			e.GetLogger(c).Warn("request body is not present anymore. ")
			err = nil
			continue
		}
		if err != nil {
			err = fmt.Errorf("%v; %w", err, err)
			break
		}
	}
	//vd.SetErrorFactory(func(failPath, msg string) error {
	//	return fmt.Errorf(`"validation failed: %s %s"`, failPath, msg)
	//})
	if err1 := vd.Validate(d); err1 != nil {
		err = fmt.Errorf("%v; %w", err, err1)
	}
	return err
}

// Error 通常错误数据处理
func (e *RestApi) Error(c *gin.Context, code int, err error, msg string) {
	response.Error(c, code, err, msg)

	// 根据RESTful API 设计原则： 在 RESTful API 设计中，对于特定资源的操作
	//（如 GET、PUT、PATCH、DELETE），如果该资源不存在，返回 404 是一种常见且推荐的做法。
	// 但是，如果你希望 UPDATE 操作是幂等的，最适合的 HTTP 状态码是 200 OK。
	// 最佳实践：
	//     返回 200 OK 状态码。
	//     在响应体中包含操作结果的详细信息。
}

// OK 通常成功数据处理
func (e *RestApi) OK(c *gin.Context, data interface{}, msg string) {
	response.OK(c, data, msg)
}

// PageOK 分页数据处理
func (e *RestApi) PageOK(c *gin.Context, result interface{}, count int, pageIndex int, pageSize int, msg string) {
	response.PageOK(c, result, count, pageIndex, pageSize, msg)
}

// Custom 兼容函数
func (e *RestApi) Custom(c *gin.Context, data gin.H) {
	response.Custum(c, data)
}
