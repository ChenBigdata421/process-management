package middleware

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/ChenBigdata421/jxt-core/sdk"
	"github.com/ChenBigdata421/jxt-core/sdk/config"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/jwtauth/user"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"jxt-evidence-system/process-management/shared/common"
	"jxt-evidence-system/process-management/shared/common/global"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
)

const (
	OperaStatusEnabel  = "1" // 状态-正常
	OperaStatusDisable = "2" // 状态-关闭
)

// LoggerToFile 日志记录到文件
func LoggerToFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.GetRequestLogger(c)
		// 开始时间
		startTime := time.Now()
		// 处理请求
		var body string
		switch c.Request.Method {
		case http.MethodPost, http.MethodPut, http.MethodGet, http.MethodDelete:
			bf := bytes.NewBuffer(nil)
			wt := bufio.NewWriter(bf)
			_, err := io.Copy(wt, c.Request.Body)
			if err != nil {
				log.Warn("copy body error", zap.Error(err))
				err = nil
			}
			// 必须调用Flush()将缓冲区数据写入buffer
			err = wt.Flush()
			if err != nil {
				log.Warn("flush writer error", zap.Error(err))
			}
			rb, _ := ioutil.ReadAll(bf)
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(rb))
			body = string(rb)
		}

		c.Next()
		url := c.Request.RequestURI
		if strings.Index(url, "logout") > -1 ||
			strings.Index(url, "login") > -1 {
			return
		}
		// 结束时间
		endTime := time.Now()
		if c.Request.Method == http.MethodOptions {
			return
		}

		rt, bl := c.Get("result")
		var result = ""
		if bl {
			rb, err := json.Marshal(rt)
			if err != nil {
				log.Warn("json Marshal result error", zap.Error(err))
			} else {
				result = string(rb)
			}
		}

		st, bl := c.Get("status")
		var statusBus = 0
		if bl {
			statusBus = st.(int)
		}

		// 获取请求主机名
		host := c.Request.Host
		host = strings.Split(host, ":")[0]

		// 请求方式
		reqMethod := c.Request.Method
		// 请求路由
		reqUri := c.Request.RequestURI
		// 状态码
		statusCode := c.Writer.Status()
		// 请求IP
		clientIP := common.GetClientIP(c)
		// 执行时间
		latencyTime := endTime.Sub(startTime)
		// 日志格式
		logData := map[string]interface{}{
			"host":        host,
			"statusCode":  statusCode,
			"latencyTime": latencyTime,
			"clientIP":    clientIP,
			"method":      reqMethod,
			"uri":         reqUri,
			"user":        user.GetUserName(c),
		}
		log.Info("request", zap.Any("request", logData))

		if c.Request.Method != "OPTIONS" && config.LoggerConfig.EnabledDB && statusCode != 404 {
			SetDBOperLog(c, clientIP, statusCode, reqUri, reqMethod, latencyTime, body, result, statusBus)
		}
	}
}

// SetDBOperLog 写入操作日志表 fixme 该方法后续即将弃用
func SetDBOperLog(c *gin.Context, clientIP string, statusCode int, reqUri string, reqMethod string, latencyTime time.Duration, body string, result string, status int) {

	log := logger.GetRequestLogger(c)
	l := make(map[string]interface{})
	l["_fullPath"] = c.FullPath()
	l["operUrl"] = reqUri
	l["operIp"] = clientIP
	l["operLocation"] = "" // pkg.GetLocation(clientIP, gaConfig.ExtConfig.AMap.Key)
	l["operName"] = user.GetUserName(c)
	l["requestMethod"] = reqMethod
	l["operParam"] = body
	l["operTime"] = time.Now()
	l["jsonResult"] = result
	l["latencyTime"] = latencyTime.String()
	l["statusCode"] = statusCode
	l["userAgent"] = c.Request.UserAgent()
	l["createBy"] = user.GetUserId(c)
	l["updateBy"] = user.GetUserId(c)
	if status == http.StatusOK {
		l["status"] = OperaStatusEnabel
	} else {
		l["status"] = OperaStatusDisable
	}
	q := sdk.Runtime.GetMemoryQueue(c.Request.Host)
	message, err := sdk.Runtime.GetStreamMessage("", global.OperateLog, l)
	if err != nil {
		log.Error("GetStreamMessage error", zap.Error(err))
		//日志报错错误，不中断请求
	} else {
		err = q.Append(message)
		if err != nil {
			log.Error("Append message error", zap.Error(err))
		}
	}
}
