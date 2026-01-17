package common

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func GetClientIP(c *gin.Context) string {
	ClientIP := c.ClientIP()
	//fmt.Println("ClientIP:", ClientIP)
	RemoteIPStr := c.RemoteIP()
	//fmt.Println("RemoteIP:", RemoteIPStr)
	ip := c.Request.Header.Get("X-Forwarded-For")
	if strings.Contains(ip, "127.0.0.1") || ip == "" {
		ip = c.Request.Header.Get("X-real-ip")
	}
	if ip == "" {
		ip = "127.0.0.1"
	}
	if RemoteIPStr != "" && RemoteIPStr != "127.0.0.1" {
		ip = RemoteIPStr
	}
	if ClientIP != "127.0.0.1" {
		ip = ClientIP
	}
	return ip
}
