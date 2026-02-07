package valueobject

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// DeleteApprovalID 警情ID值对象
type DeleteApprovalID struct {
	value uuid.UUID
}

// NewDeleteApprovalID 创建新的DeleteApprovalID
// UUID v7 是基于时间戳的，适合数据库索引，时间戳 + 随机数
func NewDeleteApprovalID() DeleteApprovalID {
	return DeleteApprovalID{value: uuid.Must(uuid.NewV7())}
}

// DeleteApprovalIDFromString 从字符串创建DeleteApprovalID
func DeleteApprovalIDFromString(s string) (DeleteApprovalID, error) {
	if s == "" {
		return DeleteApprovalID{}, nil // 空值对象
	}

	parsedUUID, err := uuid.Parse(s)
	if err != nil {
		return DeleteApprovalID{}, fmt.Errorf("invalid DeleteApprovalID format: %w", err)
	}

	return DeleteApprovalID{value: parsedUUID}, nil
}

// DeleteApprovalIDFromBytes 从字节数组创建DeleteApprovalID（用于数据库扫描）
func DeleteApprovalIDFromBytes(b []byte) (DeleteApprovalID, error) {
	if len(b) == 0 {
		return DeleteApprovalID{}, nil
	}

	if len(b) != 16 {
		return DeleteApprovalID{}, fmt.Errorf("invalid DeleteApprovalID bytes length: expected 16, got %d", len(b))
	}

	parsedUUID, err := uuid.FromBytes(b)
	if err != nil {
		return DeleteApprovalID{}, fmt.Errorf("failed to parse DeleteApprovalID from bytes: %w", err)
	}

	return DeleteApprovalID{value: parsedUUID}, nil
}

// String 返回字符串表示
func (id DeleteApprovalID) String() string {
	if id.IsEmpty() {
		return ""
	}
	return id.value.String()
}

// IsEmpty 检查是否为空值对象
func (id DeleteApprovalID) IsEmpty() bool {
	return id.value == uuid.Nil
}

// Equals 比较两个DeleteApprovalID是否相等
func (id DeleteApprovalID) Equals(other DeleteApprovalID) bool {
	return id.value == other.value
}

// Value 实现driver.Valuer接口，用于数据库存储
func (id DeleteApprovalID) Value() (driver.Value, error) {
	if id.IsEmpty() {
		return nil, nil
	}
	return id.String(), nil
}

// Scan 实现sql.Scanner接口，用于数据库扫描
func (id *DeleteApprovalID) Scan(value interface{}) error {
	if value == nil {
		*id = DeleteApprovalID{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			*id = DeleteApprovalID{}
			return nil
		}
		mediaID, err := DeleteApprovalIDFromBytes(v)
		if err != nil {
			return err
		}
		*id = mediaID
		return nil
	case string:
		mediaID, err := DeleteApprovalIDFromString(v)
		if err != nil {
			return err
		}
		*id = mediaID
		return nil
	default:
		return fmt.Errorf("cannot scan %T into DeleteApprovalID", value)
	}
}

// MarshalJSON 实现JSON序列化
func (id DeleteApprovalID) MarshalJSON() ([]byte, error) {
	if id.IsEmpty() {
		return json.Marshal("")
	}
	return json.Marshal(id.String())
}

// UnmarshalJSON 实现JSON反序列化
func (id *DeleteApprovalID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	mediaID, err := DeleteApprovalIDFromString(s)
	if err != nil {
		return err
	}

	*id = mediaID
	return nil
}

// ===== URI参数绑定支持 =====

// NewDeleteApprovalIDFromString 从字符串创建警情媒体关联ID
func NewDeleteApprovalIDFromString(id string) (DeleteApprovalID, error) {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return DeleteApprovalID{}, fmt.Errorf("无效的警情ID格式: %w", err)
	}
	return DeleteApprovalID{value: parsedUUID}, nil
}

// MarshalText 实现 encoding.TextMarshaler 接口
// 支持GORM查询参数序列化
func (id DeleteApprovalID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

// UnmarshalText 实现 encoding.TextUnmarshaler 接口
// 支持Gin框架的URI参数绑定和GORM查询参数序列化
func (id *DeleteApprovalID) UnmarshalText(text []byte) error {
	newID, err := NewDeleteApprovalIDFromString(string(text))
	if err != nil {
		return err
	}
	*id = newID
	return nil
}

// UnmarshalParam 实现 binding.BindUnmarshaler 接口
// 支持Gin框架的URI参数绑定（ShouldBindUri）和Query参数绑定
// 注意：Gin的ShouldBindUri需要此接口才能正确绑定自定义类型
func (id *DeleteApprovalID) UnmarshalParam(param string) error {
	newID, err := NewDeleteApprovalIDFromString(param)
	if err != nil {
		return err
	}
	*id = newID
	return nil
}
