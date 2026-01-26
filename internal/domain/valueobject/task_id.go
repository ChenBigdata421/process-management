package valueobject

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// TaskID 警情ID值对象
type TaskID struct {
	value uuid.UUID
}

// NewTaskID 创建新的TaskID
// UUID v7 是基于时间戳的，适合数据库索引，时间戳 + 随机数
func NewTaskID() TaskID {
	return TaskID{value: uuid.Must(uuid.NewV7())}
}

// TaskIDFromString 从字符串创建TaskID
func TaskIDFromString(s string) (TaskID, error) {
	if s == "" {
		return TaskID{}, nil // 空值对象
	}

	parsedUUID, err := uuid.Parse(s)
	if err != nil {
		return TaskID{}, fmt.Errorf("invalid TaskID format: %w", err)
	}

	return TaskID{value: parsedUUID}, nil
}

// TaskIDFromBytes 从字节数组创建TaskID（用于数据库扫描）
func TaskIDFromBytes(b []byte) (TaskID, error) {
	if len(b) == 0 {
		return TaskID{}, nil
	}

	if len(b) != 16 {
		return TaskID{}, fmt.Errorf("invalid TaskID bytes length: expected 16, got %d", len(b))
	}

	parsedUUID, err := uuid.FromBytes(b)
	if err != nil {
		return TaskID{}, fmt.Errorf("failed to parse TaskID from bytes: %w", err)
	}

	return TaskID{value: parsedUUID}, nil
}

// String 返回字符串表示
func (id TaskID) String() string {
	if id.IsEmpty() {
		return ""
	}
	return id.value.String()
}

// IsEmpty 检查是否为空值对象
func (id TaskID) IsEmpty() bool {
	return id.value == uuid.Nil
}

// Equals 比较两个TaskID是否相等
func (id TaskID) Equals(other TaskID) bool {
	return id.value == other.value
}

// Value 实现driver.Valuer接口，用于数据库存储
func (id TaskID) Value() (driver.Value, error) {
	if id.IsEmpty() {
		return nil, nil
	}
	return id.value[:], nil // 返回16字节数组用于MySQL binary(16)存储
}

// Scan 实现sql.Scanner接口，用于数据库扫描
func (id *TaskID) Scan(value interface{}) error {
	if value == nil {
		*id = TaskID{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			*id = TaskID{}
			return nil
		}
		mediaID, err := TaskIDFromBytes(v)
		if err != nil {
			return err
		}
		*id = mediaID
		return nil
	case string:
		mediaID, err := TaskIDFromString(v)
		if err != nil {
			return err
		}
		*id = mediaID
		return nil
	default:
		return fmt.Errorf("cannot scan %T into TaskID", value)
	}
}

// MarshalJSON 实现JSON序列化
func (id TaskID) MarshalJSON() ([]byte, error) {
	if id.IsEmpty() {
		return json.Marshal("")
	}
	return json.Marshal(id.String())
}

// UnmarshalJSON 实现JSON反序列化
func (id *TaskID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	mediaID, err := TaskIDFromString(s)
	if err != nil {
		return err
	}

	*id = mediaID
	return nil
}

// ===== URI参数绑定支持 =====

// NewTaskIDFromString 从字符串创建警情媒体关联ID
func NewTaskIDFromString(id string) (TaskID, error) {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return TaskID{}, fmt.Errorf("无效的警情ID格式: %w", err)
	}
	return TaskID{value: parsedUUID}, nil
}

// MarshalText 实现 encoding.TextMarshaler 接口
// 支持GORM查询参数序列化
func (id TaskID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

// UnmarshalText 实现 encoding.TextUnmarshaler 接口
// 支持Gin框架的URI参数绑定和GORM查询参数序列化
func (id *TaskID) UnmarshalText(text []byte) error {
	newID, err := NewTaskIDFromString(string(text))
	if err != nil {
		return err
	}
	*id = newID
	return nil
}

// UnmarshalParam 实现 binding.BindUnmarshaler 接口
// 支持Gin框架的URI参数绑定（ShouldBindUri）和Query参数绑定
// 注意：Gin的ShouldBindUri需要此接口才能正确绑定自定义类型
func (id *TaskID) UnmarshalParam(param string) error {
	newID, err := NewTaskIDFromString(param)
	if err != nil {
		return err
	}
	*id = newID
	return nil
}
