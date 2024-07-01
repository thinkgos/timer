package comparator

import (
	"cmp"
	"time"
)

type Int int
type Int8 int8
type Int16 int16
type Int32 int32
type Int64 int64
type Uint uint
type Uint8 uint8
type Uint16 uint16
type Uint32 uint32
type Uint64 uint64
type Float32 float32
type Float64 float64
type Uintptr uintptr
type String string
type Time time.Time

func (v1 Int) CompareTo(v2 Int) int         { return cmp.Compare(v1, v2) }
func (v1 Int8) CompareTo(v2 Int8) int       { return cmp.Compare(v1, v2) }
func (v1 Int16) CompareTo(v2 Int16) int     { return cmp.Compare(v1, v2) }
func (v1 Int32) CompareTo(v2 Int32) int     { return cmp.Compare(v1, v2) }
func (v1 Int64) CompareTo(v2 Int64) int     { return cmp.Compare(v1, v2) }
func (v1 Uint) CompareTo(v2 Uint) int       { return cmp.Compare(v1, v2) }
func (v1 Uint8) CompareTo(v2 Uint8) int     { return cmp.Compare(v1, v2) }
func (v1 Uint16) CompareTo(v2 Uint16) int   { return cmp.Compare(v1, v2) }
func (v1 Uint32) CompareTo(v2 Uint32) int   { return cmp.Compare(v1, v2) }
func (v1 Uint64) CompareTo(v2 Uint64) int   { return cmp.Compare(v1, v2) }
func (v1 Float32) CompareTo(v2 Float32) int { return cmp.Compare(v1, v2) }
func (v1 Float64) CompareTo(v2 Float64) int { return cmp.Compare(v1, v2) }
func (v1 Uintptr) CompareTo(v2 Uintptr) int { return cmp.Compare(v1, v2) }
func (v1 String) CompareTo(v2 String) int   { return cmp.Compare(v1, v2) }
func (v1 Time) CompareTo(v2 Time) int {
	t1 := time.Time(v1)
	t2 := time.Time(v2)
	if t1.Before(t2) {
		return -1
	}
	if t1.After(t2) {
		return 1
	}
	return 0
}
