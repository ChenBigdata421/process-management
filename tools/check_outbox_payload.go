package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type OutboxEvent struct {
	ID        string          `gorm:"column:id"`
	Payload   json.RawMessage `gorm:"column:payload"`
	EventType string          `gorm:"column:event_type"`
	CreatedAt string          `gorm:"column:created_at"`
}

func main() {
	// 连接到数据库
	dsn := "postgres://root:123456@localhost:5436/processdb?sslmode=disable&connect_timeout=1&TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 查询最新的 outbox_events
	var events []OutboxEvent
	if err := db.Table("outbox_events").
		Order("created_at DESC").
		Limit(5).
		Find(&events).Error; err != nil {
		log.Fatalf("Failed to query outbox_events: %v", err)
	}

	if len(events) == 0 {
		fmt.Println("No outbox events found")
		return
	}

	fmt.Println("=== Outbox Events Payload Analysis ===")

	for i, event := range events {
		fmt.Printf("Event %d:\n", i+1)
		fmt.Printf("  ID: %s\n", event.ID)
		fmt.Printf("  EventType: %s\n", event.EventType)
		fmt.Printf("  CreatedAt: %s\n", event.CreatedAt)
		fmt.Printf("  Raw Payload (first 150 chars): %.150s\n", string(event.Payload))

		// 尝试作为 JSON 解析
		var jsonPayload interface{}
		if err := json.Unmarshal(event.Payload, &jsonPayload); err == nil {
			fmt.Printf("  ✓ Valid JSON format\n")

			// 检查 JSON 内容是否为字符串（可能是 Base64 编码的）
			if str, ok := jsonPayload.(string); ok {
				fmt.Printf("  JSON contains a string value (first 100 chars): %.100s\n", str)

				// 尝试将字符串内容作为 Base64 解码
				if decodedBytes, err := base64.StdEncoding.DecodeString(str); err == nil {
					fmt.Printf("  ✓ String content is valid Base64\n")
					fmt.Printf("  Decoded (first 150 chars): %.150s\n", string(decodedBytes))

					// 尝试解析解码后的 JSON
					var decodedPayload interface{}
					if err := json.Unmarshal(decodedBytes, &decodedPayload); err == nil {
						fmt.Printf("  ✓ Decoded payload is valid JSON\n")
						fmt.Printf("  Decoded JSON: %v\n", decodedPayload)
					}
				} else {
					fmt.Printf("  ✗ String content is not Base64: %v\n", err)
				}
			} else {
				fmt.Printf("  JSON is not a string, it's: %T\n", jsonPayload)
				fmt.Printf("  Content: %v\n", jsonPayload)
			}
		} else {
			fmt.Printf("  ✗ Not valid JSON: %v\n", err)
		}

		// 尝试直接作为 Base64 解码
		if decodedBytes, err := base64.StdEncoding.DecodeString(string(event.Payload)); err == nil {
			fmt.Printf("  ✓ Raw payload is valid Base64\n")
			fmt.Printf("  Decoded (first 150 chars): %.150s\n", string(decodedBytes))
		} else {
			fmt.Printf("  ✗ Raw payload is not Base64\n")
		}

		fmt.Println()
	}
}
