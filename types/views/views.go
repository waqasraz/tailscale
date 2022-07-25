// Copyright (c) 2022 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package views provides read-only accessors for commonly used
// value types.
package views

import (
	"encoding/json"
	"errors"

	"tailscale.com/net/netaddr"
	"tailscale.com/net/tsaddr"
)

func unmarshalJSON[T any](b []byte, x *[]T) error {
	if *x != nil {
		return errors.New("already initialized")
	}
	if len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, x)
}

// StructView represents the corresponding StructView of a Viewable. The concrete types are
// typically generated by tailscale.com/cmd/viewer.
type StructView[T any] interface {
	// Valid reports whether the underlying Viewable is nil.
	Valid() bool
	// AsStruct returns a deep-copy of the underlying value.
	// It returns nil, if Valid() is false.
	AsStruct() T
}

// ViewCloner is any type that has had View and Clone funcs generated using
// tailscale.com/cmd/viewer.
type ViewCloner[T any, V StructView[T]] interface {
	// View returns a read-only view of Viewable.
	// If Viewable is nil, View().Valid() reports false.
	View() V
	// Clone returns a deep-clone of Viewable.
	// It returns nil, when Viewable is nil.
	Clone() T
}

// SliceOfViews returns a ViewSlice for x.
func SliceOfViews[T ViewCloner[T, V], V StructView[T]](x []T) SliceView[T, V] {
	return SliceView[T, V]{x}
}

// SliceView is a read-only wrapper around a struct which should only be exposed
// as a View.
type SliceView[T ViewCloner[T, V], V StructView[T]] struct {
	// ж is the underlying mutable value, named with a hard-to-type
	// character that looks pointy like a pointer.
	// It is named distinctively to make you think of how dangerous it is to escape
	// to callers. You must not let callers be able to mutate it.
	ж []T
}

// MarshalJSON implements json.Marshaler.
func (v SliceView[T, V]) MarshalJSON() ([]byte, error) { return json.Marshal(v.ж) }

// UnmarshalJSON implements json.Unmarshaler.
func (v *SliceView[T, V]) UnmarshalJSON(b []byte) error { return unmarshalJSON(b, &v.ж) }

// IsNil reports whether the underlying slice is nil.
func (v SliceView[T, V]) IsNil() bool { return v.ж == nil }

// Len returns the length of the slice.
func (v SliceView[T, V]) Len() int { return len(v.ж) }

// At returns a View of the element at index `i` of the slice.
func (v SliceView[T, V]) At(i int) V { return v.ж[i].View() }

// AppendTo appends the underlying slice values to dst.
func (v SliceView[T, V]) AppendTo(dst []V) []V {
	for _, x := range v.ж {
		dst = append(dst, x.View())
	}
	return dst
}

// AsSlice returns a copy of underlying slice.
func (v SliceView[T, V]) AsSlice() []V {
	return v.AppendTo(nil)
}

// Slice is a read-only accessor for a slice.
type Slice[T any] struct {
	// ж is the underlying mutable value, named with a hard-to-type
	// character that looks pointy like a pointer.
	// It is named distinctively to make you think of how dangerous it is to escape
	// to callers. You must not let callers be able to mutate it.
	ж []T
}

// SliceOf returns a Slice for the provided slice for immutable values.
// It is the caller's responsibility to make sure V is immutable.
func SliceOf[T any](x []T) Slice[T] {
	return Slice[T]{x}
}

// MarshalJSON implements json.Marshaler.
func (v Slice[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.ж)
}

// UnmarshalJSON implements json.Unmarshaler.
func (v *Slice[T]) UnmarshalJSON(b []byte) error {
	return unmarshalJSON(b, &v.ж)
}

// IsNil reports whether the underlying slice is nil.
func (v Slice[T]) IsNil() bool { return v.ж == nil }

// Len returns the length of the slice.
func (v Slice[T]) Len() int { return len(v.ж) }

// At returns the element at index `i` of the slice.
func (v Slice[T]) At(i int) T { return v.ж[i] }

// AppendTo appends the underlying slice values to dst.
func (v Slice[T]) AppendTo(dst []T) []T {
	return append(dst, v.ж...)
}

// AsSlice returns a copy of underlying slice.
func (v Slice[T]) AsSlice() []T {
	return v.AppendTo(v.ж[:0:0])
}

// IndexFunc returns the first index of an element in v satisfying f(e),
// or -1 if none do.
//
// As it runs in O(n) time, use with care.
func (v Slice[T]) IndexFunc(f func(T) bool) int {
	for i := 0; i < v.Len(); i++ {
		if f(v.At(i)) {
			return i
		}
	}
	return -1
}

// ContainsFunc reports whether any element in v satisfies f(e).
//
// As it runs in O(n) time, use with care.
func (v Slice[T]) ContainsFunc(f func(T) bool) bool {
	for i := 0; i < v.Len(); i++ {
		if f(v.At(i)) {
			return true
		}
	}
	return false
}

// SliceContains reports whether v contains element e.
//
// As it runs in O(n) time, use with care.
func SliceContains[T comparable](v Slice[T], e T) bool {
	for i := 0; i < v.Len(); i++ {
		if v.At(i) == e {
			return true
		}
	}
	return false
}

// IPPrefixSlice is a read-only accessor for a slice of netaddr.IPPrefix.
type IPPrefixSlice struct {
	ж Slice[netaddr.IPPrefix]
}

// IPPrefixSliceOf returns a IPPrefixSlice for the provided slice.
func IPPrefixSliceOf(x []netaddr.IPPrefix) IPPrefixSlice { return IPPrefixSlice{SliceOf(x)} }

// IsNil reports whether the underlying slice is nil.
func (v IPPrefixSlice) IsNil() bool { return v.ж.IsNil() }

// Len returns the length of the slice.
func (v IPPrefixSlice) Len() int { return v.ж.Len() }

// At returns the IPPrefix at index `i` of the slice.
func (v IPPrefixSlice) At(i int) netaddr.IPPrefix { return v.ж.At(i) }

// AppendTo appends the underlying slice values to dst.
func (v IPPrefixSlice) AppendTo(dst []netaddr.IPPrefix) []netaddr.IPPrefix {
	return v.ж.AppendTo(dst)
}

// Unwrap returns the underlying Slice[netaddr.IPPrefix].
func (v IPPrefixSlice) Unwrap() Slice[netaddr.IPPrefix] {
	return v.ж
}

// AsSlice returns a copy of underlying slice.
func (v IPPrefixSlice) AsSlice() []netaddr.IPPrefix {
	return v.ж.AsSlice()
}

// PrefixesContainsIP reports whether any IPPrefix contains IP.
func (v IPPrefixSlice) ContainsIP(ip netaddr.IP) bool {
	return tsaddr.PrefixesContainsIP(v.ж.ж, ip)
}

// PrefixesContainsFunc reports whether f is true for any IPPrefix in the slice.
func (v IPPrefixSlice) ContainsFunc(f func(netaddr.IPPrefix) bool) bool {
	return tsaddr.PrefixesContainsFunc(v.ж.ж, f)
}

// ContainsExitRoutes reports whether v contains ExitNode Routes.
func (v IPPrefixSlice) ContainsExitRoutes() bool {
	return tsaddr.ContainsExitRoutes(v.ж.ж)
}

// MarshalJSON implements json.Marshaler.
func (v IPPrefixSlice) MarshalJSON() ([]byte, error) {
	return v.ж.MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler.
func (v *IPPrefixSlice) UnmarshalJSON(b []byte) error {
	return v.ж.UnmarshalJSON(b)
}

// MapOf returns a view over m. It is the caller's responsibility to make sure K
// and V is immutable, if this is being used to provide a read-only view over m.
func MapOf[K comparable, V comparable](m map[K]V) Map[K, V] {
	return Map[K, V]{m}
}

// Map is a view over a map whose values are immutable.
type Map[K comparable, V any] struct {
	// ж is the underlying mutable value, named with a hard-to-type
	// character that looks pointy like a pointer.
	// It is named distinctively to make you think of how dangerous it is to escape
	// to callers. You must not let callers be able to mutate it.
	ж map[K]V
}

// Has reports whether k has an entry in the map.
func (m Map[K, V]) Has(k K) bool {
	_, ok := m.ж[k]
	return ok
}

// IsNil reports whether the underlying map is nil.
func (m Map[K, V]) IsNil() bool {
	return m.ж == nil
}

// Len returns the number of elements in the map.
func (m Map[K, V]) Len() int { return len(m.ж) }

// Get returns the element with key k.
func (m Map[K, V]) Get(k K) V {
	return m.ж[k]
}

// GetOk returns the element with key k and a bool representing whether the key
// is in map.
func (m Map[K, V]) GetOk(k K) (V, bool) {
	v, ok := m.ж[k]
	return v, ok
}

// MapRangeFn is the func called from a Map.Range call.
// Implementations should return false to stop range.
type MapRangeFn[K comparable, V any] func(k K, v V) (cont bool)

// Range calls f for every k,v pair in the underlying map.
// It stops iteration immediately if f returns false.
func (m Map[K, V]) Range(f MapRangeFn[K, V]) {
	for k, v := range m.ж {
		if !f(k, v) {
			return
		}
	}
}

// MapFnOf returns a MapFn for m.
func MapFnOf[K comparable, T any, V any](m map[K]T, f func(T) V) MapFn[K, T, V] {
	return MapFn[K, T, V]{
		ж:     m,
		wrapv: f,
	}
}

// MapFn is like Map but with a func to convert values from T to V.
// It is used to provide map of slices and views.
type MapFn[K comparable, T any, V any] struct {
	// ж is the underlying mutable value, named with a hard-to-type
	// character that looks pointy like a pointer.
	// It is named distinctively to make you think of how dangerous it is to escape
	// to callers. You must not let callers be able to mutate it.
	ж     map[K]T
	wrapv func(T) V
}

// Has reports whether k has an entry in the map.
func (m MapFn[K, T, V]) Has(k K) bool {
	_, ok := m.ж[k]
	return ok
}

// Get returns the element with key k.
func (m MapFn[K, T, V]) Get(k K) V {
	return m.wrapv(m.ж[k])
}

// IsNil reports whether the underlying map is nil.
func (m MapFn[K, T, V]) IsNil() bool {
	return m.ж == nil
}

// Len returns the number of elements in the map.
func (m MapFn[K, T, V]) Len() int { return len(m.ж) }

// GetOk returns the element with key k and a bool representing whether the key
// is in map.
func (m MapFn[K, T, V]) GetOk(k K) (V, bool) {
	v, ok := m.ж[k]
	return m.wrapv(v), ok
}

// Range calls f for every k,v pair in the underlying map.
// It stops iteration immediately if f returns false.
func (m MapFn[K, T, V]) Range(f MapRangeFn[K, V]) {
	for k, v := range m.ж {
		if !f(k, m.wrapv(v)) {
			return
		}
	}
}
