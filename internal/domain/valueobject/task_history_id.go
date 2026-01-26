package valueobject

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// TaskHistoryID 案件ID值对象
// 使用UUIDv7实现，提供时间有序性和全局唯一性
type TaskHistoryID struct {
	value uuid.UUID
}

// NewTaskHistoryID 创建新的案件ID
// 使用UUIDv7确保时间有序性，适合数据库索引
func NewTaskHistoryID() TaskHistoryID {
	return TaskHistoryID{value: uuid.Must(uuid.NewV7())}
}

// NewTaskHistoryIDFromUUID 从UUID创建案件ID
func NewTaskHistoryIDFromUUID(id uuid.UUID) TaskHistoryID {
	return TaskHistoryID{value: id}
}

// NewTaskHistoryIDFromString 从字符串创建案件ID
func NewTaskHistoryIDFromString(id string) (TaskHistoryID, error) {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return TaskHistoryID{}, fmt.Errorf("invalid TaskHistory ID format: %w", err)
	}
	return TaskHistoryID{value: parsedUUID}, nil
}

// String 返回字符串表示
func (id TaskHistoryID) String() string {
	return id.value.String()
}

// UUID 返回UUID值
func (id TaskHistoryID) UUID() uuid.UUID {
	return id.value
}

// Equals 比较两个TaskHistoryID是否相等
func (id TaskHistoryID) Equals(other TaskHistoryID) bool {
	return id.value == other.value
}

// IsZero 检查是否为零值
func (id TaskHistoryID) IsZero() bool {
	return id.value == uuid.Nil
}

// Value 实现 driver.Valuer 接口，用于数据库存储
func (id TaskHistoryID) Value() (driver.Value, error) {
	if id.IsZero() {
		return nil, nil
	}
	return id.value[:], nil // 返回UUID字符串格式，兼容PostgreSQL
}

// Scan 实现 sql.Scanner 接口，用于从数据库读取
func (id *TaskHistoryID) Scan(value interface{}) error {
	if value == nil {
		id.value = uuid.Nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		if len(v) == 16 {
			// 从16字节的二进制格式读取
			parsedUUID, err := uuid.FromBytes(v)
			if err != nil {
				return fmt.Errorf("failed to parse TaskHistory ID from bytes: %w", err)
			}
			id.value = parsedUUID
			return nil
		}
		// 尝试从字符串格式读取
		parsedUUID, err := uuid.ParseBytes(v)
		if err != nil {
			return fmt.Errorf("failed to parse TaskHistory ID from string bytes: %w", err)
		}
		id.value = parsedUUID
		return nil
	case string:
		parsedUUID, err := uuid.Parse(v)
		if err != nil {
			return fmt.Errorf("failed to parse TaskHistory ID from string: %w", err)
		}
		id.value = parsedUUID
		return nil
	default:
		return fmt.Errorf("unsupported type for TaskHistory ID: %T", value)
	}
}

// MarshalJSON 实现 json.Marshaler 接口
func (id TaskHistoryID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.value.String())
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (id *TaskHistoryID) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	parsedUUID, err := uuid.Parse(str)
	if err != nil {
		return fmt.Errorf("failed to parse TaskHistory ID from JSON: %w", err)
	}
	id.value = parsedUUID
	return nil
}

// IsEmpty 检查是否为空值对象
func (id *TaskHistoryID) IsEmpty() bool {
	return id.value == uuid.Nil
}
