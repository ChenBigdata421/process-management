package valueobject

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// InstanceID 警情ID值对象
type InstanceID struct {
	value uuid.UUID
}

// NewInstanceID 创建新的InstanceID
// UUID v7 是基于时间戳的，适合数据库索引，时间戳 + 随机数
func NewInstanceID() InstanceID {
	return InstanceID{value: uuid.Must(uuid.NewV7())}
}

// InstanceIDFromString 从字符串创建InstanceID
func InstanceIDFromString(s string) (InstanceID, error) {
	if s == "" {
		return InstanceID{}, nil // 空值对象
	}

	parsedUUID, err := uuid.Parse(s)
	if err != nil {
		return InstanceID{}, fmt.Errorf("invalid InstanceID format: %w", err)
	}

	return InstanceID{value: parsedUUID}, nil
}

// InstanceIDFromBytes 从字节数组创建InstanceID（用于数据库扫描）
func InstanceIDFromBytes(b []byte) (InstanceID, error) {
	if len(b) == 0 {
		return InstanceID{}, nil
	}

	if len(b) != 16 {
		return InstanceID{}, fmt.Errorf("invalid InstanceID bytes length: expected 16, got %d", len(b))
	}

	parsedUUID, err := uuid.FromBytes(b)
	if err != nil {
		return InstanceID{}, fmt.Errorf("failed to parse InstanceID from bytes: %w", err)
	}

	return InstanceID{value: parsedUUID}, nil
}

// String 返回字符串表示
func (id InstanceID) String() string {
	if id.IsEmpty() {
		return ""
	}
	return id.value.String()
}

// IsEmpty 检查是否为空值对象
func (id InstanceID) IsEmpty() bool {
	return id.value == uuid.Nil
}

// Equals 比较两个InstanceID是否相等
func (id InstanceID) Equals(other InstanceID) bool {
	return id.value == other.value
}

// Value 实现driver.Valuer接口，用于数据库存储
func (id InstanceID) Value() (driver.Value, error) {
	if id.IsEmpty() {
		return nil, nil
	}
	return id.value[:], nil // 返回16字节数组用于MySQL binary(16)存储
}

// Scan 实现sql.Scanner接口，用于数据库扫描
func (id *InstanceID) Scan(value interface{}) error {
	if value == nil {
		*id = InstanceID{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			*id = InstanceID{}
			return nil
		}
		mediaID, err := InstanceIDFromBytes(v)
		if err != nil {
			return err
		}
		*id = mediaID
		return nil
	case string:
		mediaID, err := InstanceIDFromString(v)
		if err != nil {
			return err
		}
		*id = mediaID
		return nil
	default:
		return fmt.Errorf("cannot scan %T into InstanceID", value)
	}
}

// MarshalJSON 实现JSON序列化
func (id InstanceID) MarshalJSON() ([]byte, error) {
	if id.IsEmpty() {
		return json.Marshal("")
	}
	return json.Marshal(id.String())
}

// UnmarshalJSON 实现JSON反序列化
func (id *InstanceID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	mediaID, err := InstanceIDFromString(s)
	if err != nil {
		return err
	}

	*id = mediaID
	return nil
}

// ===== URI参数绑定支持 =====

// NewInstanceIDFromString 从字符串创建警情媒体关联ID
func NewInstanceIDFromString(id string) (InstanceID, error) {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return InstanceID{}, fmt.Errorf("无效的警情ID格式: %w", err)
	}
	return InstanceID{value: parsedUUID}, nil
}

// MarshalText 实现 encoding.TextMarshaler 接口
// 支持GORM查询参数序列化
func (id InstanceID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

// UnmarshalText 实现 encoding.TextUnmarshaler 接口
// 支持Gin框架的URI参数绑定和GORM查询参数序列化
func (id *InstanceID) UnmarshalText(text []byte) error {
	newID, err := NewInstanceIDFromString(string(text))
	if err != nil {
		return err
	}
	*id = newID
	return nil
}

// UnmarshalParam 实现 binding.BindUnmarshaler 接口
// 支持Gin框架的URI参数绑定（ShouldBindUri）和Query参数绑定
// 注意：Gin的ShouldBindUri需要此接口才能正确绑定自定义类型
func (id *InstanceID) UnmarshalParam(param string) error {
	newID, err := NewInstanceIDFromString(param)
	if err != nil {
		return err
	}
	*id = newID
	return nil
}
