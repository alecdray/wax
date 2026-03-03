package utils

import "cmp"

func NewPointer[T any](value T) *T {
	return &value
}

func Clamp[T cmp.Ordered](input, minVal, maxVal T) T {
	clamped := max(minVal, min(maxVal, input))
	return clamped
}
