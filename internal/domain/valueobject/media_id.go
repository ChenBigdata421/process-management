package valueobject

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// MediaID 媒体ID值对象
type MediaID struct {
	value uuid.UUID
}

// NewMediaID 创建新的MediaID
func NewMediaID() MediaID {
	return MediaID{value: uuid.Must(uuid.NewV7())}
}

// NewMediaIDFromUUID 从UUID创建MediaID
func NewMediaIDFromUUID(id uuid.UUID) MediaID {
	return MediaID{value: id}
}

// MediaIDFromString 从字符串创建MediaID
func MediaIDFromString(s string) (MediaID, error) {
	if s == "" {
		return MediaID{}, nil // 空值对象
	}

	parsedUUID, err := uuid.Parse(s)
	if err != nil {
		return MediaID{}, fmt.Errorf("invalid MediaID format: %w", err)
	}

	return MediaID{value: parsedUUID}, nil
}

// MediaIDFromBytes 从字节数组创建MediaID（用于数据库扫描）
func MediaIDFromBytes(b []byte) (MediaID, error) {
	if len(b) == 0 {
		return MediaID{}, nil
	}

	if len(b) != 16 {
		return MediaID{}, fmt.Errorf("invalid MediaID bytes length: expected 16, got %d", len(b))
	}

	parsedUUID, err := uuid.FromBytes(b)
	if err != nil {
		return MediaID{}, fmt.Errorf("failed to parse MediaID from bytes: %w", err)
	}

	return MediaID{value: parsedUUID}, nil
}

// String 返回字符串表示
func (id MediaID) String() string {
	if id.IsEmpty() {
		return ""
	}
	return id.value.String()
}

// IsEmpty 检查是否为空值对象
func (id MediaID) IsEmpty() bool {
	return id.value == uuid.Nil
}

// Equals 比较两个MediaID是否相等
func (id MediaID) Equals(other MediaID) bool {
	return id.value == other.value
}

// Value 实现driver.Valuer接口，用于数据库存储
func (id MediaID) Value() (driver.Value, error) {
	if id.IsEmpty() {
		return nil, nil
	}
	return id.String(), nil
}

// Scan 实现sql.Scanner接口，用于数据库扫描
func (id *MediaID) Scan(value interface{}) error {
	if value == nil {
		*id = MediaID{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			*id = MediaID{}
			return nil
		}
		mediaID, err := MediaIDFromBytes(v)
		if err != nil {
			return err
		}
		*id = mediaID
		return nil
	case string:
		mediaID, err := MediaIDFromString(v)
		if err != nil {
			return err
		}
		*id = mediaID
		return nil
	default:
		return fmt.Errorf("cannot scan %T into MediaID", value)
	}
}

// MarshalJSON 实现JSON序列化
func (id MediaID) MarshalJSON() ([]byte, error) {
	if id.IsEmpty() {
		return json.Marshal("")
	}
	return json.Marshal(id.String())
}

// UnmarshalJSON 实现JSON反序列化
func (id *MediaID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	mediaID, err := MediaIDFromString(s)
	if err != nil {
		return err
	}

	*id = mediaID
	return nil
}

// ===== URI参数绑定支持 =====

// UnmarshalText 实现 encoding.TextUnmarshaler 接口
// 支持Gin框架的URI参数绑定和GORM查询参数序列化
func (id *MediaID) UnmarshalText(text []byte) error {
	newID, err := MediaIDFromString(string(text))
	if err != nil {
		return err
	}
	*id = newID
	return nil
}

// MarshalText 实现 encoding.TextMarshaler 接口
// 支持GORM查询参数序列化
func (id MediaID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

// UnmarshalParam 实现 binding.BindUnmarshaler 接口
// 支持Gin框架的URI参数绑定（ShouldBindUri）和Query参数绑定
func (id *MediaID) UnmarshalParam(param string) error {
	newID, err := MediaIDFromString(param)
	if err != nil {
		return err
	}
	*id = newID
	return nil
}
