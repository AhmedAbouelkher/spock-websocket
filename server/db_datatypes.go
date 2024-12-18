package main

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// MARK: - UUID

// This datatype stores the uuid in the database as a string. To store the uuid
// in the database as a binary (byte) array, please refer to datatypes.BinUUID.
type UUID uuid.UUID

// NewUUIDv1 generates a UUID version 1, panics on generation failure.
func NewUUIDv1() UUID {
	return UUID(uuid.Must(uuid.NewUUID()))
}

// NewUUIDv4 generates a UUID version 4, panics on generation failure.
func NewUUIDv4() UUID {
	return UUID(uuid.Must(uuid.NewRandom()))
}

// GormDataType gorm common data type.
func (UUID) GormDataType() string {
	return "string"
}

// GormDBDataType gorm db data type.
func (UUID) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql":
		return "LONGTEXT"
	case "postgres":
		return "UUID"
	case "sqlserver":
		return "NVARCHAR(128)"
	case "sqlite":
		return "TEXT"
	default:
		return ""
	}
}

// Scan is the scanner function for this datatype.
func (u *UUID) Scan(value interface{}) error {
	var result uuid.UUID
	if err := result.Scan(value); err != nil {
		return err
	}
	*u = UUID(result)
	return nil
}

// Value is the valuer function for this datatype.
func (u UUID) Value() (driver.Value, error) {
	return uuid.UUID(u).Value()
}

// String returns the string form of the UUID.
func (u UUID) String() string {
	return uuid.UUID(u).String()
}

// Equals returns true if string form of UUID matches other, false otherwise.
func (u UUID) Equals(other UUID) bool {
	return u.String() == other.String()
}

// Length returns the number of characters in string form of UUID.
func (u UUID) Length() int {
	return len(u.String())
}

// IsNil returns true if the UUID is a nil UUID (all zeroes), false otherwise.
func (u UUID) IsNil() bool {
	return uuid.UUID(u) == uuid.Nil
}

// IsEmpty returns true if UUID is nil UUID or of zero length, false otherwise.
func (u UUID) IsEmpty() bool {
	return u.IsNil() || u.Length() == 0
}

// IsNilPtr returns true if caller UUID ptr is nil, false otherwise.
func (u *UUID) IsNilPtr() bool {
	return u == nil
}

// IsEmptyPtr returns true if caller UUID ptr is nil or it's value is empty.
func (u *UUID) IsEmptyPtr() bool {
	return u.IsNilPtr() || u.IsEmpty()
}

func (u UUID) MarshalJSON() ([]byte, error) {
	return []byte(`"` + u.String() + `"`), nil
}

func UUIDFromString(s string) (UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return UUID{}, err
	}
	return UUID(id), nil
}

// MARK: - StringArray

type StringArray []string

func (a StringArray) GormDataType() string { return "string[]" }

func (a StringArray) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "postgres":
		return "VARCHAR(255)[]"
	default:
		return ""
	}
}

func (a *StringArray) Scan(value interface{}) (err error) {
	str, ok := value.(string)
	if !ok {
		return errors.New("failed to scan multi-string field - source is not a string")
	}
	str = strings.Trim(str, "{}")
	*a = strings.Split(str, ",")
	return nil
}

func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	fa := []string{}
	for _, v := range a {
		vv := strings.TrimSpace(v)
		vv = strings.Trim(vv, "\"")
		vv = strings.Trim(vv, "'")
		vv = strings.Trim(vv, ",")
		if vv == "" {
			continue
		}
		fa = append(fa, vv)
	}
	v := strings.Join(fa, ",")
	out := "{" + v + "}"
	return out, nil
}

func NewStringArray(s ...string) StringArray { return StringArray(s) }

// MARK: - IntArray

type IntArray []int

func (a IntArray) GormDataType() string { return "bigint[]" }

func (a IntArray) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "postgres":
		return "BIGINT[]"
	default:
		return ""
	}
}

func (a *IntArray) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return errors.New("failed to scan multi-int field - source is not a string")
	}
	str = strings.Trim(str, "{}")
	sa := strings.Split(str, ",")
	ia := make([]int, len(sa))
	for i, v := range sa {
		iv, err := strconv.Atoi(v)
		if err != nil {
			return err
		}
		ia[i] = iv
	}
	*a = ia
	return nil
}

func (a IntArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	fa := []string{}
	for _, v := range a {
		fa = append(fa, strconv.Itoa(v))
	}
	v := strings.Join(fa, ",")
	out := "{" + v + "}"
	return out, nil
}

func (a IntArray) IsEmpty() bool { return len(a) == 0 }

func (a IntArray) ToUintArr() []uint {
	ua := make([]uint, len(a))
	for i, v := range a {
		ua[i] = uint(v)
	}
	return ua
}

func NewIntArray(s ...int) IntArray { return IntArray(s) }

func NewIntArrayFromUint(s ...uint) IntArray {
	ia := make([]int, len(s))
	for i, v := range s {
		ia[i] = int(v)
	}
	return ia
}

func PGArray(values interface{}) string {
	switch v := values.(type) {
	case []int:
		if len(v) == 0 {
			return "{}"
		}
		return fmt.Sprintf("{%s}", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(v)), ","), "[]"))
	case []uint:
		if len(v) == 0 {
			return "{}"
		}
		return fmt.Sprintf("{%s}", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(v)), ","), "[]"))
	case []string:
		if len(v) == 0 {
			return "{}"
		}
		quoted := make([]string, len(v))
		for i, s := range v {
			quoted[i] = fmt.Sprintf("%q", s)
		}
		return fmt.Sprintf("{%s}", strings.Join(quoted, ","))
	default:
		return "{}"
	}
}
