package valueobject

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// WorkflowID 警情ID值对象
type WorkflowID struct {
	value uuid.UUID
}

// NewWorkflowID 创建新的WorkflowID
// UUID v7 是基于时间戳的，适合数据库索引，时间戳 + 随机数
func NewWorkflowID() WorkflowID {
	return WorkflowID{value: uuid.Must(uuid.NewV7())}
}

// WorkflowIDFromString 从字符串创建WorkflowID
func WorkflowIDFromString(s string) (WorkflowID, error) {
	if s == "" {
		return WorkflowID{}, nil // 空值对象
	}

	parsedUUID, err := uuid.Parse(s)
	if err != nil {
		return WorkflowID{}, fmt.Errorf("invalid WorkflowID format: %w", err)
	}

	return WorkflowID{value: parsedUUID}, nil
}

// WorkflowIDFromBytes 从字节数组创建WorkflowID（用于数据库扫描）
func WorkflowIDFromBytes(b []byte) (WorkflowID, error) {
	if len(b) == 0 {
		return WorkflowID{}, nil
	}

	if len(b) != 16 {
		return WorkflowID{}, fmt.Errorf("invalid WorkflowID bytes length: expected 16, got %d", len(b))
	}

	parsedUUID, err := uuid.FromBytes(b)
	if err != nil {
		return WorkflowID{}, fmt.Errorf("failed to parse WorkflowID from bytes: %w", err)
	}

	return WorkflowID{value: parsedUUID}, nil
}

// String 返回字符串表示
func (id WorkflowID) String() string {
	if id.IsEmpty() {
		return ""
	}
	return id.value.String()
}

// IsEmpty 检查是否为空值对象
func (id WorkflowID) IsEmpty() bool {
	return id.value == uuid.Nil
}

// Equals 比较两个WorkflowID是否相等
func (id WorkflowID) Equals(other WorkflowID) bool {
	return id.value == other.value
}

// Value 实现driver.Valuer接口，用于数据库存储
func (id WorkflowID) Value() (driver.Value, error) {
	if id.IsEmpty() {
		return nil, nil
	}
	return id.value[:], nil // 返回16字节数组用于MySQL binary(16)存储
}

// Scan 实现sql.Scanner接口，用于数据库扫描
func (id *WorkflowID) Scan(value interface{}) error {
	if value == nil {
		*id = WorkflowID{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			*id = WorkflowID{}
			return nil
		}
		mediaID, err := WorkflowIDFromBytes(v)
		if err != nil {
			return err
		}
		*id = mediaID
		return nil
	case string:
		mediaID, err := WorkflowIDFromString(v)
		if err != nil {
			return err
		}
		*id = mediaID
		return nil
	default:
		return fmt.Errorf("cannot scan %T into WorkflowID", value)
	}
}

// MarshalJSON 实现JSON序列化
func (id WorkflowID) MarshalJSON() ([]byte, error) {
	if id.IsEmpty() {
		return json.Marshal("")
	}
	return json.Marshal(id.String())
}

// UnmarshalJSON 实现JSON反序列化
func (id *WorkflowID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	mediaID, err := WorkflowIDFromString(s)
	if err != nil {
		return err
	}

	*id = mediaID
	return nil
}

// ===== URI参数绑定支持 =====

// NewWorkflowIDFromString 从字符串创建警情媒体关联ID
func NewWorkflowIDFromString(id string) (WorkflowID, error) {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return WorkflowID{}, fmt.Errorf("无效的警情ID格式: %w", err)
	}
	return WorkflowID{value: parsedUUID}, nil
}

// MarshalText 实现 encoding.TextMarshaler 接口
// 支持GORM查询参数序列化
func (id WorkflowID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

// UnmarshalText 实现 encoding.TextUnmarshaler 接口
// 支持Gin框架的URI参数绑定和GORM查询参数序列化
func (id *WorkflowID) UnmarshalText(text []byte) error {
	newID, err := NewWorkflowIDFromString(string(text))
	if err != nil {
		return err
	}
	*id = newID
	return nil
}

// UnmarshalParam 实现 binding.BindUnmarshaler 接口
// 支持Gin框架的URI参数绑定（ShouldBindUri）和Query参数绑定
// 注意：Gin的ShouldBindUri需要此接口才能正确绑定自定义类型
func (id *WorkflowID) UnmarshalParam(param string) error {
	newID, err := NewWorkflowIDFromString(param)
	if err != nil {
		return err
	}
	*id = newID
	return nil
}
