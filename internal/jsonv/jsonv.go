// Package jsonv is a dynamic JSON value for responses that are traversed ad hoc: a missing key or a wrong type
// yields an empty Value rather than a panic, and numbers keep their full precision.
package jsonv

import (
	"bytes"
	"encoding/json"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/ac1982/haul/internal/errs"
)

// Value wraps a decoded JSON value: map[string]any, []any, string, json.Number, bool or nil.
type Value struct{ v any }

// Null is the empty value.
var Null = Value{}

// Of wraps a Go value built from maps, slices, strings, numbers and bools.
func Of(v any) Value { return Value{v} }

// Parse decodes JSON text.
func Parse(data []byte) (Value, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return Null, errs.New("Invalid JSON: %v", err)
	}
	return Value{v}, nil
}

// ParseString decodes JSON text.
func ParseString(s string) (Value, error) { return Parse([]byte(s)) }

// Raw is the underlying Go value.
func (j Value) Raw() any { return j.v }

// Get is the value under key, or Null.
func (j Value) Get(key string) Value {
	if m, ok := j.v.(map[string]any); ok {
		return Value{m[key]}
	}
	return Null
}

// At is the element at index i, or Null.
func (j Value) At(i int) Value {
	if a, ok := j.v.([]any); ok && i >= 0 && i < len(a) {
		return Value{a[i]}
	}
	return Null
}

// Path follows keys: j.Path("data", "list") is j.Get("data").Get("list").
func (j Value) Path(keys ...string) Value {
	for _, k := range keys {
		j = j.Get(k)
	}
	return j
}

// Require is the value under key; a missing key is an error. Use for fields the response must have.
func (j Value) Require(key string) (Value, error) {
	if m, ok := j.v.(map[string]any); ok {
		if v, ok := m[key]; ok {
			return Value{v}, nil
		}
	}
	return Null, &MissingKeyError{Key: key}
}

// MissingKeyError is what Require returns.
type MissingKeyError struct{ Key string }

func (e *MissingKeyError) Error() string { return "Missing JSON field: " + e.Key }

// Has reports whether the object has key.
func (j Value) Has(key string) bool {
	m, ok := j.v.(map[string]any)
	if !ok {
		return false
	}
	_, ok = m[key]
	return ok
}

func (j Value) IsNull() bool   { return j.v == nil }
func (j Value) Exists() bool   { return j.v != nil }
func (j Value) IsObject() bool { _, ok := j.v.(map[string]any); return ok }
func (j Value) IsArray() bool  { _, ok := j.v.([]any); return ok }

// Str is the string, when the value is one.
func (j Value) Str() (string, bool) {
	s, ok := j.v.(string)
	return s, ok
}

// String is the text form of any value: scalars as text, containers serialized, null as "".
// It matches how responses are string-searched.
func (j Value) String() string {
	switch v := j.v.(type) {
	case nil:
		return ""
	case string:
		return v
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return strconv.FormatInt(i, 10)
		}
		if f, err := v.Float64(); err == nil && f == math.Trunc(f) && math.Abs(f) < 1e15 {
			return strconv.FormatInt(int64(f), 10)
		}
		return v.String()
	case bool:
		return strconv.FormatBool(v)
	default:
		return j.Serialized()
	}
}

// Int64 is the value as an integer: whole numbers, numeric strings and bools.
func (j Value) Int64() (int64, bool) {
	switch v := j.v.(type) {
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return i, true
		}
		if f, err := v.Float64(); err == nil && f == math.Trunc(f) && math.Abs(f) < 9.2e18 {
			return int64(f), true
		}
	case float64:
		if v == math.Trunc(v) && math.Abs(v) < 9.2e18 {
			return int64(v), true
		}
	case int:
		return int64(v), true
	case int64:
		return v, true
	case string:
		if i, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err == nil {
			return i, true
		}
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil && f == math.Trunc(f) {
			return int64(f), true
		}
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	}
	return 0, false
}

// Int is Int64 as an int.
func (j Value) Int() (int, bool) {
	i, ok := j.Int64()
	return int(i), ok
}

// IntOr is the integer, or def.
func (j Value) IntOr(def int) int {
	if i, ok := j.Int(); ok {
		return i
	}
	return def
}

// Int64Or is the integer, or def.
func (j Value) Int64Or(def int64) int64 {
	if i, ok := j.Int64(); ok {
		return i
	}
	return def
}

// Float is the value as a number: numbers and numeric strings.
func (j Value) Float() (float64, bool) {
	switch v := j.v.(type) {
	case json.Number:
		f, err := v.Float64()
		return f, err == nil
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		return f, err == nil
	}
	return 0, false
}

// FloatOr is the number, or def.
func (j Value) FloatOr(def float64) float64 {
	if f, ok := j.Float(); ok {
		return f
	}
	return def
}

// Bool is the value as a bool: bools, numbers (non-zero) and "true" / "false".
func (j Value) Bool() (bool, bool) {
	switch v := j.v.(type) {
	case bool:
		return v, true
	case json.Number, float64, int, int64:
		i, ok := j.Int64()
		return i != 0, ok
	case string:
		switch v {
		case "true":
			return true, true
		case "false":
			return false, true
		}
	}
	return false, false
}

// Array is the elements, or none.
func (j Value) Array() []Value {
	a, ok := j.v.([]any)
	if !ok {
		return nil
	}
	out := make([]Value, len(a))
	for i, v := range a {
		out[i] = Value{v}
	}
	return out
}

// Object is the members, or none.
func (j Value) Object() map[string]Value {
	m, ok := j.v.(map[string]any)
	if !ok {
		return nil
	}
	out := make(map[string]Value, len(m))
	for k, v := range m {
		out[k] = Value{v}
	}
	return out
}

// Keys are the object's keys, sorted.
func (j Value) Keys() []string {
	m, ok := j.v.(map[string]any)
	if !ok {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Serialized is compact JSON with sorted keys and unescaped HTML characters.
func (j Value) Serialized() string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(j.v); err != nil {
		return ""
	}
	return strings.TrimSuffix(buf.String(), "\n")
}
