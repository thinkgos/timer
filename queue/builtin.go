package queue

import (
	"time"

	"golang.org/x/exp/constraints"
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

func (v1 Int) CompareTo(v2 Comparable) int     { return CompareTo(int(v1), int(v2.(Int))) }
func (v1 Int8) CompareTo(v2 Comparable) int    { return CompareTo(int8(v1), int8(v2.(Int8))) }
func (v1 Int16) CompareTo(v2 Comparable) int   { return CompareTo(int16(v1), int16(v2.(Int16))) }
func (v1 Int32) CompareTo(v2 Comparable) int   { return CompareTo(int32(v1), int32(v2.(Int32))) }
func (v1 Int64) CompareTo(v2 Comparable) int   { return CompareTo(int64(v1), int64(v2.(Int64))) }
func (v1 Uint) CompareTo(v2 Comparable) int    { return CompareTo(uint(v1), uint(v2.(Uint))) }
func (v1 Uint8) CompareTo(v2 Comparable) int   { return CompareTo(uint8(v1), uint8(v2.(Uint8))) }
func (v1 Uint16) CompareTo(v2 Comparable) int  { return CompareTo(uint16(v1), uint16(v2.(Uint16))) }
func (v1 Uint32) CompareTo(v2 Comparable) int  { return CompareTo(uint32(v1), uint32(v2.(Uint32))) }
func (v1 Uint64) CompareTo(v2 Comparable) int  { return CompareTo(uint64(v1), uint64(v2.(Uint64))) }
func (v1 Float32) CompareTo(v2 Comparable) int { return CompareTo(float32(v1), float32(v2.(Float32))) }
func (v1 Float64) CompareTo(v2 Comparable) int { return CompareTo(float64(v1), float64(v2.(Float64))) }
func (v1 Uintptr) CompareTo(v2 Comparable) int { return CompareTo(uintptr(v1), uintptr(v2.(Uintptr))) }
func (v1 String) CompareTo(v2 Comparable) int  { return CompareTo(string(v1), string(v2.(String))) }
func (v1 Time) CompareTo(v2 Comparable) int {
	t1 := time.Time(v1)
	t2 := time.Time(v2.(Time))
	if t1.Before(t2) {
		return -1
	}
	if t1.After(t2) {
		return 1
	}
	return 0
}

func CompareTo[T constraints.Ordered](v1, v2 T) int {
	if v1 < v2 {
		return -1
	}
	if v1 > v2 {
		return 1
	}
	return 0
}
