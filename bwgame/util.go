package bwgame

import (
	"cmp"
	"fmt"
	"iter"
	"sort"
	"unsafe"
)

type Exception struct {
	Message string
}

func (e *Exception) Error() string {
	return e.Message
}

func throwError(format string, args ...any) {
	panic(&Exception{Message: fmt.Sprintf(format, args...)})
}

type Pair[A, B any] struct {
	First  A
	Second B
}

type xyComponent interface {
	~int8 | ~int16 | ~int32 | ~int64 | ~int | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uint
}

type XyFp8 = XyT[Fp8]

func XyLess[T xyComponent](a, b XyT[T]) bool {
	if a.Y == b.Y {
		return a.X < b.X
	}
	return a.Y < b.Y
}

func XyGreater[T xyComponent](a, b XyT[T]) bool {
	if a.Y == b.Y {
		return a.X > b.X
	}
	return a.Y > b.Y
}

func XyLessEqual[T xyComponent](a, b XyT[T]) bool {
	if a.Y == b.Y {
		return a.X <= b.X
	}
	return a.Y <= b.Y
}

func XyGreaterEqual[T xyComponent](a, b XyT[T]) bool {
	if a.Y == b.Y {
		return a.X >= b.X
	}
	return a.Y >= b.Y
}

func XyAdd[T xyComponent](a, b XyT[T]) XyT[T] {
	return XyT[T]{X: a.X + b.X, Y: a.Y + b.Y}
}

func XySub[T xyComponent](a, b XyT[T]) XyT[T] {
	return XyT[T]{X: a.X - b.X, Y: a.Y - b.Y}
}

func XyNeg[T xyComponent](a XyT[T]) XyT[T] {
	return XyT[T]{X: -a.X, Y: -a.Y}
}

func XyMul[T xyComponent](a, b XyT[T]) XyT[T] {
	return XyT[T]{X: a.X * b.X, Y: a.Y * b.Y}
}

func XyDiv[T xyComponent](a, b XyT[T]) XyT[T] {
	return XyT[T]{X: a.X / b.X, Y: a.Y / b.Y}
}

func XyMulScalar[T xyComponent](a XyT[T], v T) XyT[T] {
	return XyT[T]{X: a.X * v, Y: a.Y * v}
}

func XyDivScalar[T xyComponent](a XyT[T], v T) XyT[T] {
	return XyT[T]{X: a.X / v, Y: a.Y / v}
}

func XyFp8Add(a, b XyFp8) XyFp8 {
	return XyFp8{X: a.X.Add(b.X), Y: a.Y.Add(b.Y)}
}

func XyFp8Sub(a, b XyFp8) XyFp8 {
	return XyFp8{X: a.X.Sub(b.X), Y: a.Y.Sub(b.Y)}
}

func XyFp8Neg(a XyFp8) XyFp8 {
	return XyFp8{X: a.X.Neg(), Y: a.Y.Neg()}
}

func XyFp8Mul(a XyFp8, v Fp8) XyFp8 {
	return XyFp8{X: a.X.Mul(v), Y: a.Y.Mul(v)}
}

func XyFp8Div(a XyFp8, v Fp8) XyFp8 {
	return XyFp8{X: a.X.Div(v), Y: a.Y.Div(v)}
}

func XyFp8MulInt(a XyFp8, v int32) XyFp8 {
	return XyFp8{X: a.X.MulInt(v), Y: a.Y.MulInt(v)}
}

func XyFp8DivInt(a XyFp8, v int32) XyFp8 {
	return XyFp8{X: a.X.DivInt(v), Y: a.Y.DivInt(v)}
}

func RectAdd[T xyComponent](a, b RectT[XyT[T]]) RectT[XyT[T]] {
	return RectT[XyT[T]]{From: XyAdd(a.From, b.From), To: XyAdd(a.To, b.To)}
}

func Fp8Integer(v int32) Fp8 {
	return Fp8{RawValue: int32(uint32(v) << 8)}
}

func Fp8Zero() Fp8 {
	return Fp8Integer(0)
}

func (f Fp8) IntegerPart() int32 {
	return f.RawValue >> 8
}

func (f Fp8) FractionalPart() int32 {
	return f.RawValue & (1<<8 - 1)
}

func (f Fp8) Floor() Fp8 {
	return Fp8Integer(f.IntegerPart())
}

func (f Fp8) Ceil() Fp8 {
	return f.Add(Fp8Integer(1)).Sub(Fp8FromRaw(1)).Floor()
}

func (f Fp8) Abs() Fp8 {
	if f.RawValue >= 0 {
		return f
	}
	return Fp8{RawValue: -f.RawValue}
}

func (f Fp8) AsUnsigned() Ufp8 {
	return Ufp8{RawValue: uint32(f.RawValue)}
}

func (f Fp8) Less(n Fp8) bool         { return f.RawValue < n.RawValue }
func (f Fp8) LessEqual(n Fp8) bool    { return f.RawValue <= n.RawValue }
func (f Fp8) Greater(n Fp8) bool      { return f.RawValue > n.RawValue }
func (f Fp8) GreaterEqual(n Fp8) bool { return f.RawValue >= n.RawValue }

func (f Fp8) Neg() Fp8           { return Fp8{RawValue: -f.RawValue} }
func (f Fp8) Add(n Fp8) Fp8      { return Fp8{RawValue: f.RawValue + n.RawValue} }
func (f Fp8) Sub(n Fp8) Fp8      { return Fp8{RawValue: f.RawValue - n.RawValue} }
func (f Fp8) MulInt(v int32) Fp8 { return Fp8{RawValue: f.RawValue * v} }
func (f Fp8) DivInt(v int32) Fp8 { return Fp8{RawValue: f.RawValue / v} }

func (f Fp8) Mul(n Fp8) Fp8 {
	return Fp8{RawValue: int32(int64(f.RawValue) * int64(n.RawValue) >> 8)}
}

func (f Fp8) Div(n Fp8) Fp8 {
	return Fp8{RawValue: int32(int64(f.RawValue) << 8 / int64(n.RawValue))}
}

func Fp8MultiplyDivide(a, b, c Fp8) Fp8 {
	return Fp8{RawValue: int32(int64(a.RawValue) * int64(b.RawValue) / int64(c.RawValue))}
}

func Fp8DivideMultiply(a, b, c Fp8) Fp8 {
	return Fp8{RawValue: a.RawValue / b.RawValue * c.RawValue}
}

func Fp8Extend(d DirectionT) Fp8 {
	return Fp8{RawValue: int32(d.RawValue)}
}

func Fp8Truncate(n Fp32x8) Fp8 {
	return Fp8{RawValue: int32(n.RawValue)}
}

type Ufp8 struct {
	RawValue uint32
}

func Ufp8FromRaw(rawValue uint32) Ufp8 {
	return Ufp8{RawValue: rawValue}
}

func Ufp8Integer(v uint32) Ufp8 {
	return Ufp8{RawValue: v << 8}
}

func Ufp8Zero() Ufp8 {
	return Ufp8Integer(0)
}

func Ufp8One() Ufp8 {
	return Ufp8Integer(1)
}

func (f Ufp8) IntegerPart() uint32 {
	return f.RawValue >> 8
}

func (f Ufp8) FractionalPart() uint32 {
	return f.RawValue & (1<<8 - 1)
}

func (f Ufp8) Floor() Ufp8 {
	return Ufp8Integer(f.IntegerPart())
}

func (f Ufp8) Ceil() Ufp8 {
	return f.Add(Ufp8Integer(1)).Sub(Ufp8FromRaw(1)).Floor()
}

func (f Ufp8) Abs() Ufp8 {
	return f
}

func (f Ufp8) AsSigned() Fp8 {
	return Fp8{RawValue: int32(f.RawValue)}
}

func (f Ufp8) Less(n Ufp8) bool         { return f.RawValue < n.RawValue }
func (f Ufp8) LessEqual(n Ufp8) bool    { return f.RawValue <= n.RawValue }
func (f Ufp8) Greater(n Ufp8) bool      { return f.RawValue > n.RawValue }
func (f Ufp8) GreaterEqual(n Ufp8) bool { return f.RawValue >= n.RawValue }

func (f Ufp8) Add(n Ufp8) Ufp8      { return Ufp8{RawValue: f.RawValue + n.RawValue} }
func (f Ufp8) Sub(n Ufp8) Ufp8      { return Ufp8{RawValue: f.RawValue - n.RawValue} }
func (f Ufp8) MulInt(v uint32) Ufp8 { return Ufp8{RawValue: f.RawValue * v} }
func (f Ufp8) DivInt(v uint32) Ufp8 { return Ufp8{RawValue: f.RawValue / v} }

func (f Ufp8) Mul(n Ufp8) Ufp8 {
	return Ufp8{RawValue: uint32(uint64(f.RawValue) * uint64(n.RawValue) >> 8)}
}

func (f Ufp8) Div(n Ufp8) Ufp8 {
	return Ufp8{RawValue: uint32(uint64(f.RawValue) << 8 / uint64(n.RawValue))}
}

func Ufp8MultiplyDivide(a, b, c Ufp8) Ufp8 {
	return Ufp8{RawValue: uint32(uint64(a.RawValue) * uint64(b.RawValue) / uint64(c.RawValue))}
}

func Ufp8DivideMultiply(a, b, c Ufp8) Ufp8 {
	return Ufp8{RawValue: a.RawValue / b.RawValue * c.RawValue}
}

func Fp1Integer(v int32) Fp1 {
	return Fp1{RawValue: int32(uint32(v) << 1)}
}

func Fp1Zero() Fp1 {
	return Fp1Integer(0)
}

func Fp1One() Fp1 {
	return Fp1Integer(1)
}

func (f Fp1) IntegerPart() int32 {
	return f.RawValue >> 1
}

func (f Fp1) FractionalPart() int32 {
	return f.RawValue & 1
}

func (f Fp1) Floor() Fp1 {
	return Fp1Integer(f.IntegerPart())
}

func (f Fp1) Ceil() Fp1 {
	return f.Add(Fp1Integer(1)).Sub(Fp1FromRaw(1)).Floor()
}

func (f Fp1) Abs() Fp1 {
	if f.RawValue >= 0 {
		return f
	}
	return Fp1{RawValue: -f.RawValue}
}

func (f Fp1) Less(n Fp1) bool         { return f.RawValue < n.RawValue }
func (f Fp1) LessEqual(n Fp1) bool    { return f.RawValue <= n.RawValue }
func (f Fp1) Greater(n Fp1) bool      { return f.RawValue > n.RawValue }
func (f Fp1) GreaterEqual(n Fp1) bool { return f.RawValue >= n.RawValue }

func (f Fp1) Neg() Fp1           { return Fp1{RawValue: -f.RawValue} }
func (f Fp1) Add(n Fp1) Fp1      { return Fp1{RawValue: f.RawValue + n.RawValue} }
func (f Fp1) Sub(n Fp1) Fp1      { return Fp1{RawValue: f.RawValue - n.RawValue} }
func (f Fp1) MulInt(v int32) Fp1 { return Fp1{RawValue: f.RawValue * v} }
func (f Fp1) DivInt(v int32) Fp1 { return Fp1{RawValue: f.RawValue / v} }

func (f Fp1) Mul(n Fp1) Fp1 {
	return Fp1{RawValue: int32(int64(f.RawValue) * int64(n.RawValue) >> 1)}
}

func (f Fp1) Div(n Fp1) Fp1 {
	return Fp1{RawValue: int32(int64(f.RawValue) << 1 / int64(n.RawValue))}
}

type Fp16 struct {
	RawValue int32
}

func Fp16FromRaw(rawValue int32) Fp16 {
	return Fp16{RawValue: rawValue}
}

func Fp16Integer(v int32) Fp16 {
	return Fp16{RawValue: int32(uint32(v) << 16)}
}

func (f Fp16) IntegerPart() int32 {
	return f.RawValue >> 16
}

func (f Fp16) FractionalPart() int32 {
	return f.RawValue & (1<<16 - 1)
}

func (f Fp16) Floor() Fp16 {
	return Fp16Integer(f.IntegerPart())
}

func (f Fp16) Ceil() Fp16 {
	return f.Add(Fp16Integer(1)).Sub(Fp16FromRaw(1)).Floor()
}

func (f Fp16) Abs() Fp16 {
	if f.RawValue >= 0 {
		return f
	}
	return Fp16{RawValue: -f.RawValue}
}

func (f Fp16) AsUnsigned() Ufp16 {
	return Ufp16{RawValue: uint32(f.RawValue)}
}

func (f Fp16) Less(n Fp16) bool         { return f.RawValue < n.RawValue }
func (f Fp16) LessEqual(n Fp16) bool    { return f.RawValue <= n.RawValue }
func (f Fp16) Greater(n Fp16) bool      { return f.RawValue > n.RawValue }
func (f Fp16) GreaterEqual(n Fp16) bool { return f.RawValue >= n.RawValue }

func (f Fp16) Neg() Fp16           { return Fp16{RawValue: -f.RawValue} }
func (f Fp16) Add(n Fp16) Fp16     { return Fp16{RawValue: f.RawValue + n.RawValue} }
func (f Fp16) Sub(n Fp16) Fp16     { return Fp16{RawValue: f.RawValue - n.RawValue} }
func (f Fp16) MulInt(v int32) Fp16 { return Fp16{RawValue: f.RawValue * v} }
func (f Fp16) DivInt(v int32) Fp16 { return Fp16{RawValue: f.RawValue / v} }

func (f Fp16) Mul(n Fp16) Fp16 {
	return Fp16{RawValue: int32(int64(f.RawValue) * int64(n.RawValue) >> 16)}
}

func (f Fp16) Div(n Fp16) Fp16 {
	return Fp16{RawValue: int32(int64(f.RawValue) << 16 / int64(n.RawValue))}
}

func Fp16MultiplyDivide(a, b, c Fp16) Fp16 {
	return Fp16{RawValue: int32(int64(a.RawValue) * int64(b.RawValue) / int64(c.RawValue))}
}

func Fp16DivideMultiply(a, b, c Fp16) Fp16 {
	return Fp16{RawValue: a.RawValue / b.RawValue * c.RawValue}
}

type Ufp16 struct {
	RawValue uint32
}

func Ufp16FromRaw(rawValue uint32) Ufp16 {
	return Ufp16{RawValue: rawValue}
}

func Ufp16Integer(v uint32) Ufp16 {
	return Ufp16{RawValue: v << 16}
}

func (f Ufp16) IntegerPart() uint32 {
	return f.RawValue >> 16
}

func (f Ufp16) FractionalPart() uint32 {
	return f.RawValue & (1<<16 - 1)
}

func (f Ufp16) AsSigned() Fp16 {
	return Fp16{RawValue: int32(f.RawValue)}
}

func (f Ufp16) Less(n Ufp16) bool         { return f.RawValue < n.RawValue }
func (f Ufp16) LessEqual(n Ufp16) bool    { return f.RawValue <= n.RawValue }
func (f Ufp16) Greater(n Ufp16) bool      { return f.RawValue > n.RawValue }
func (f Ufp16) GreaterEqual(n Ufp16) bool { return f.RawValue >= n.RawValue }

func (f Ufp16) Add(n Ufp16) Ufp16     { return Ufp16{RawValue: f.RawValue + n.RawValue} }
func (f Ufp16) Sub(n Ufp16) Ufp16     { return Ufp16{RawValue: f.RawValue - n.RawValue} }
func (f Ufp16) MulInt(v uint32) Ufp16 { return Ufp16{RawValue: f.RawValue * v} }
func (f Ufp16) DivInt(v uint32) Ufp16 { return Ufp16{RawValue: f.RawValue / v} }

func (f Ufp16) Mul(n Ufp16) Ufp16 {
	return Ufp16{RawValue: uint32(uint64(f.RawValue) * uint64(n.RawValue) >> 16)}
}

func (f Ufp16) Div(n Ufp16) Ufp16 {
	return Ufp16{RawValue: uint32(uint64(f.RawValue) << 16 / uint64(n.RawValue))}
}

type Fp32x8 struct {
	RawValue int64
}

func Fp32x8FromRaw(rawValue int64) Fp32x8 {
	return Fp32x8{RawValue: rawValue}
}

func Fp32x8Extend(f Fp8) Fp32x8 {
	return Fp32x8{RawValue: int64(f.RawValue)}
}

func (f Fp32x8) Less(n Fp32x8) bool         { return f.RawValue < n.RawValue }
func (f Fp32x8) LessEqual(n Fp32x8) bool    { return f.RawValue <= n.RawValue }
func (f Fp32x8) Greater(n Fp32x8) bool      { return f.RawValue > n.RawValue }
func (f Fp32x8) GreaterEqual(n Fp32x8) bool { return f.RawValue >= n.RawValue }

func (f Fp32x8) MulInt(v int32) Fp32x8 { return Fp32x8{RawValue: f.RawValue * int64(v)} }
func (f Fp32x8) DivInt(v int32) Fp32x8 { return Fp32x8{RawValue: f.RawValue / int64(v)} }

func (d DirectionT) IntegerPart() int8 {
	return int8(int32(d.RawValue) >> 8)
}

func (d DirectionT) FractionalPart() int8 {
	return d.RawValue
}

func (d DirectionT) Abs() DirectionT {
	if d.RawValue >= 0 {
		return d
	}
	return DirectionT{RawValue: -d.RawValue}
}

func (d DirectionT) Less(n DirectionT) bool         { return d.RawValue < n.RawValue }
func (d DirectionT) LessEqual(n DirectionT) bool    { return d.RawValue <= n.RawValue }
func (d DirectionT) Greater(n DirectionT) bool      { return d.RawValue > n.RawValue }
func (d DirectionT) GreaterEqual(n DirectionT) bool { return d.RawValue >= n.RawValue }

func (d DirectionT) Neg() DirectionT { return DirectionT{RawValue: -d.RawValue} }
func (d DirectionT) Add(n DirectionT) DirectionT {
	return DirectionT{RawValue: d.RawValue + n.RawValue}
}
func (d DirectionT) Sub(n DirectionT) DirectionT {
	return DirectionT{RawValue: d.RawValue - n.RawValue}
}
func (d DirectionT) MulInt(v int32) DirectionT {
	return DirectionT{RawValue: int8(int32(d.RawValue) * v)}
}
func (d DirectionT) DivInt(v int32) DirectionT {
	return DirectionT{RawValue: int8(int32(d.RawValue) / v)}
}

func DirectionTTruncate(f Fp8) DirectionT {
	return DirectionT{RawValue: int8(f.RawValue)}
}

type unsignedInteger interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uint
}

func isqrt[T unsignedInteger](n T) T {
	var r T
	var zero T
	bits := 8 * unsafe.Sizeof(zero)
	p := T(1) << (bits - 2)
	for p > n {
		p /= 4
	}
	for p != 0 {
		if n >= r+p {
			n -= r + p
			r += 2 * p
		}
		r /= 2
		p /= 4
	}
	return r
}

func getBestScore[E any, S cmp.Ordered](values []E, score func(E) S) int {
	if len(values) == 0 {
		return len(values)
	}
	best := 0
	bestScore := score(values[0])
	for i := 1; i < len(values); i++ {
		s := score(values[i])
		if s < bestScore {
			best = i
			bestScore = s
		}
	}
	return best
}

func getBestScoreSeq[E any, S cmp.Ordered](values iter.Seq[E], score func(E) S) (best E, found bool) {
	var bestScore S
	for v := range values {
		s := score(v)
		if !found || s < bestScore {
			best = v
			bestScore = s
			found = true
		}
	}
	return best, found
}

type StaticVectorStorage[T any] interface {
	[4]T | [5]T | [12]T | [16]T | [32]T | [150]T | [240]T | [250]T
}

type StaticVector[T any, A StaticVectorStorage[T]] struct {
	data A
	size int
}

func (v *StaticVector[T, A]) Capacity() int {
	return len(v.data)
}

func (v *StaticVector[T, A]) Size() int {
	return v.size
}

func (v *StaticVector[T, A]) Empty() bool {
	return v.size == 0
}

func (v *StaticVector[T, A]) Values() []T {
	return unsafe.Slice(&v.data[0], v.size)
}

func (v *StaticVector[T, A]) At(pos int) T {
	return *v.Ptr(pos)
}

func (v *StaticVector[T, A]) Ptr(pos int) *T {
	if pos < 0 || pos >= v.size {
		throwError("static_vector subscript out of range")
	}
	return &v.data[pos]
}

func (v *StaticVector[T, A]) Front() T {
	return v.data[0]
}

func (v *StaticVector[T, A]) Back() T {
	return v.data[v.size-1]
}

func (v *StaticVector[T, A]) PushBack(value T) {
	if v.size == v.Capacity() {
		throwError("static_vector resized beyond capacity")
	}
	v.data[v.size] = value
	v.size++
}

func (v *StaticVector[T, A]) PopBack() {
	var zero T
	v.size--
	v.data[v.size] = zero
}

func (v *StaticVector[T, A]) Erase(pos int) {
	for i := pos; i+1 < v.size; i++ {
		v.data[i] = v.data[i+1]
	}
	v.PopBack()
}

func (v *StaticVector[T, A]) Insert(pos int, value T) {
	v.PushBack(value)
	for i := v.size - 1; i > pos; i-- {
		v.data[i] = v.data[i-1]
	}
	v.data[pos] = value
}

func (v *StaticVector[T, A]) Resize(count int) {
	if count > v.Capacity() {
		throwError("static_vector resized beyond capacity")
	}
	for v.size > count {
		v.PopBack()
	}
	var zero T
	for v.size < count {
		v.PushBack(zero)
	}
}

func (v *StaticVector[T, A]) Clear() {
	v.Resize(0)
}

func (v *StaticVector[T, A]) All() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for i := 0; i < v.size; i++ {
			if !yield(i, v.data[i]) {
				return
			}
		}
	}
}

var tanTable = [64]uint32{
	7, 13, 19, 26, 32, 38, 45, 51, 58, 65, 71, 78, 85, 92,
	99, 107, 114, 122, 129, 137, 146, 154, 163, 172, 181,
	190, 200, 211, 221, 233, 244, 256, 269, 283, 297, 312,
	329, 346, 364, 384, 405, 428, 452, 479, 509, 542, 578,
	619, 664, 716, 775, 844, 926, 1023, 1141, 1287, 1476,
	1726, 2076, 2600, 3471, 5211, 10429, 0xffffffff,
}

var repulseAdjustTable = [16 * 2]int32{
	-5, -5, -5, 5, 5, -5, 5, 5, 5, -5, 5, 5, -5, -5, -5, 5, -5, 5, 5, -5, -5, -5, 5, 5, -5, 5, 5, -5, 5, 5, -5, -5,
}

var cardinalDirectionXy = [4]Xy{{X: 1, Y: 0}, {X: 0, Y: 1}, {X: -1, Y: 0}, {X: 0, Y: -1}}

var directionTable = xyFp8TableFromRaw([256][2]int32{
	{0, -256}, {6, -256}, {13, -256}, {19, -255}, {25, -255}, {31, -254}, {38, -253}, {44, -252},
	{50, -251}, {56, -250}, {62, -248}, {68, -247}, {74, -245}, {80, -243}, {86, -241}, {92, -239},
	{98, -237}, {104, -234}, {109, -231}, {115, -229}, {121, -226}, {126, -223}, {132, -220}, {137, -216},
	{142, -213}, {147, -209}, {152, -206}, {157, -202}, {162, -198}, {167, -194}, {172, -190}, {177, -185},
	{181, -181}, {185, -177}, {190, -172}, {194, -167}, {198, -162}, {202, -157}, {206, -152}, {209, -147},
	{213, -142}, {216, -137}, {220, -132}, {223, -126}, {226, -121}, {229, -115}, {231, -109}, {234, -104},
	{237, -98}, {239, -92}, {241, -86}, {243, -80}, {245, -74}, {247, -68}, {248, -62}, {250, -56},
	{251, -50}, {252, -44}, {253, -38}, {254, -31}, {255, -25}, {255, -19}, {256, -13}, {256, -6},
	{256, 0}, {256, 6}, {256, 13}, {255, 19}, {255, 25}, {254, 31}, {253, 38}, {252, 44},
	{251, 50}, {250, 56}, {248, 62}, {247, 68}, {245, 74}, {243, 80}, {241, 86}, {239, 92},
	{237, 98}, {234, 104}, {231, 109}, {229, 115}, {226, 121}, {223, 126}, {220, 132}, {216, 137},
	{213, 142}, {209, 147}, {206, 152}, {202, 157}, {198, 162}, {194, 167}, {190, 172}, {185, 177},
	{181, 181}, {177, 185}, {172, 190}, {167, 194}, {162, 198}, {157, 202}, {152, 206}, {147, 209},
	{142, 213}, {137, 216}, {132, 220}, {126, 223}, {121, 226}, {115, 229}, {109, 231}, {104, 234},
	{98, 237}, {92, 239}, {86, 241}, {80, 243}, {74, 245}, {68, 247}, {62, 248}, {56, 250},
	{50, 251}, {44, 252}, {38, 253}, {31, 254}, {25, 255}, {19, 255}, {13, 256}, {6, 256},
	{0, 256}, {-6, 256}, {-13, 256}, {-19, 255}, {-25, 255}, {-31, 254}, {-38, 253}, {-44, 252},
	{-50, 251}, {-56, 250}, {-62, 248}, {-68, 247}, {-74, 245}, {-80, 243}, {-86, 241}, {-92, 239},
	{-98, 237}, {-104, 234}, {-109, 231}, {-115, 229}, {-121, 226}, {-126, 223}, {-132, 220}, {-137, 216},
	{-142, 213}, {-147, 209}, {-152, 206}, {-157, 202}, {-162, 198}, {-167, 194}, {-172, 190}, {-177, 185},
	{-181, 181}, {-185, 177}, {-190, 172}, {-194, 167}, {-198, 162}, {-202, 157}, {-206, 152}, {-209, 147},
	{-213, 142}, {-216, 137}, {-220, 132}, {-223, 126}, {-226, 121}, {-229, 115}, {-231, 109}, {-234, 104},
	{-237, 98}, {-239, 92}, {-241, 86}, {-243, 80}, {-245, 74}, {-247, 68}, {-248, 62}, {-250, 56},
	{-251, 50}, {-252, 44}, {-253, 38}, {-254, 31}, {-255, 25}, {-255, 19}, {-256, 13}, {-256, 6},
	{-256, 0}, {-256, -6}, {-256, -13}, {-255, -19}, {-255, -25}, {-254, -31}, {-253, -38}, {-252, -44},
	{-251, -50}, {-250, -56}, {-248, -62}, {-247, -68}, {-245, -74}, {-243, -80}, {-241, -86}, {-239, -92},
	{-237, -98}, {-234, -104}, {-231, -109}, {-229, -115}, {-226, -121}, {-223, -126}, {-220, -132}, {-216, -137},
	{-213, -142}, {-209, -147}, {-206, -152}, {-202, -157}, {-198, -162}, {-194, -167}, {-190, -172}, {-185, -177},
	{-181, -181}, {-177, -185}, {-172, -190}, {-167, -194}, {-162, -198}, {-157, -202}, {-152, -206}, {-147, -209},
	{-142, -213}, {-137, -216}, {-132, -220}, {-126, -223}, {-121, -226}, {-115, -229}, {-109, -231}, {-104, -234},
	{-98, -237}, {-92, -239}, {-86, -241}, {-80, -243}, {-74, -245}, {-68, -247}, {-62, -248}, {-56, -250},
	{-50, -251}, {-44, -252}, {-38, -253}, {-31, -254}, {-25, -255}, {-19, -255}, {-13, -256}, {-6, -256},
})

func xyFp8TableFromRaw(raw [256][2]int32) [256]XyFp8 {
	var table [256]XyFp8
	for i, v := range raw {
		table[i] = XyFp8{X: Fp8FromRaw(v[0]), Y: Fp8FromRaw(v[1])}
	}
	return table
}

var hitNearTargetPositions = [14]Xy{
	{X: -1, Y: -1}, {X: -51, Y: -31}, {X: -7, Y: -12}, {X: 36, Y: 52}, {X: 49, Y: -25}, {X: -40, Y: 46}, {X: 4, Y: 19}, {X: -28, Y: -50},
	{X: -50, Y: 15}, {X: 54, Y: 13}, {X: -11, Y: 53}, {X: -15, Y: 9}, {X: 19, Y: -50}, {X: 17, Y: -8},
}

var psiFieldMask = [5][8]bool{
	{true, true, true, true, true, true, true, true},
	{true, true, true, true, true, true, true, true},
	{true, true, true, true, true, true, true, false},
	{true, true, true, true, true, true, false, false},
	{true, true, true, false, false, false, false, false},
}

var allPlayerSlots = [12]int32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}

func contourLowerBound(contours []Contour, value int32) int {
	return sort.Search(len(contours), func(i int) bool { return contours[i].V[0] >= value })
}

func contourUpperBound(contours []Contour, value int32) int {
	return sort.Search(len(contours), func(i int) bool { return value < contours[i].V[0] })
}

func unitFinderLowerBound(entries []UnitFinderEntry, value int32) int {
	return sort.Search(len(entries), func(i int) bool { return entries[i].Value >= value })
}

func unitFinderUpperBound(entries []UnitFinderEntry, value int32) int {
	return sort.Search(len(entries), func(i int) bool { return value < entries[i].Value })
}

func absInt32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

func resizeSlice[T any](s []T, n int) []T {
	if n <= len(s) {
		return s[:n]
	}
	return append(s, make([]T, n-len(s))...)
}
