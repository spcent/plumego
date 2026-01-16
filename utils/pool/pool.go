package pool

import (
	"bytes"
	"encoding/json"
	"sync"
)

// JSONBufferPool provides a pool of bytes.Buffer for JSON encoding/decoding
var JSONBufferPool = &sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, 1024)) // Pre-allocate 1KB
	},
}

// MapPool provides a pool of map[string]any for temporary JSON operations
var MapPool = &sync.Pool{
	New: func() any {
		return make(map[string]any, 16) // Pre-allocate for 16 fields
	},
}

// ByteSlicePool provides a pool of []byte for temporary operations
var ByteSlicePool = &sync.Pool{
	New: func() any {
		return make([]byte, 0, 512) // Pre-allocate 512 bytes
	},
}

// StringSlicePool provides a pool of []string for temporary operations
var StringSlicePool = &sync.Pool{
	New: func() any {
		return make([]string, 0, 16) // Pre-allocate for 16 strings
	},
}

// IntSlicePool provides a pool of []int for temporary operations
var IntSlicePool = &sync.Pool{
	New: func() any {
		return make([]int, 0, 16) // Pre-allocate for 16 ints
	},
}

// Int64SlicePool provides a pool of []int64 for temporary operations
var Int64SlicePool = &sync.Pool{
	New: func() any {
		return make([]int64, 0, 16) // Pre-allocate for 16 int64s
	},
}

// Float64SlicePool provides a pool of []float64 for temporary operations
var Float64SlicePool = &sync.Pool{
	New: func() any {
		return make([]float64, 0, 16) // Pre-allocate for 16 float64s
	},
}

// BoolSlicePool provides a pool of []bool for temporary operations
var BoolSlicePool = &sync.Pool{
	New: func() any {
		return make([]bool, 0, 16) // Pre-allocate for 16 bools
	},
}

// MapStringSlicePool provides a pool of []map[string]any for temporary operations
var MapStringSlicePool = &sync.Pool{
	New: func() any {
		return make([]map[string]any, 0, 8) // Pre-allocate for 8 maps
	},
}

// GetBuffer retrieves a buffer from the pool
func GetBuffer() *bytes.Buffer {
	return JSONBufferPool.Get().(*bytes.Buffer)
}

// PutBuffer returns a buffer to the pool after resetting it
func PutBuffer(buf *bytes.Buffer) {
	if buf != nil {
		buf.Reset()
		JSONBufferPool.Put(buf)
	}
}

// GetMap retrieves a map from the pool
func GetMap() map[string]any {
	return MapPool.Get().(map[string]any)
}

// PutMap returns a map to the pool after clearing it
func PutMap(m map[string]any) {
	if m != nil {
		// Clear the map but keep capacity
		for k := range m {
			delete(m, k)
		}
		MapPool.Put(m)
	}
}

// GetBytes retrieves a byte slice from the pool
func GetBytes() []byte {
	return ByteSlicePool.Get().([]byte)
}

// PutBytes returns a byte slice to the pool after resetting it
func PutBytes(b []byte) {
	if b != nil {
		// Reset length but keep capacity
		b = b[:0]
		ByteSlicePool.Put(b)
	}
}

// GetStringSlice retrieves a string slice from the pool
func GetStringSlice() []string {
	return StringSlicePool.Get().([]string)
}

// PutStringSlice returns a string slice to the pool after resetting it
func PutStringSlice(s []string) {
	if s != nil {
		// Reset length but keep capacity
		s = s[:0]
		StringSlicePool.Put(s)
	}
}

// GetIntSlice retrieves an int slice from the pool
func GetIntSlice() []int {
	return IntSlicePool.Get().([]int)
}

// PutIntSlice returns an int slice to the pool after resetting it
func PutIntSlice(i []int) {
	if i != nil {
		// Reset length but keep capacity
		i = i[:0]
		IntSlicePool.Put(i)
	}
}

// GetInt64Slice retrieves an int64 slice from the pool
func GetInt64Slice() []int64 {
	return Int64SlicePool.Get().([]int64)
}

// PutInt64Slice returns an int64 slice to the pool after resetting it
func PutInt64Slice(i []int64) {
	if i != nil {
		// Reset length but keep capacity
		i = i[:0]
		Int64SlicePool.Put(i)
	}
}

// GetFloat64Slice retrieves a float64 slice from the pool
func GetFloat64Slice() []float64 {
	return Float64SlicePool.Get().([]float64)
}

// PutFloat64Slice returns a float64 slice to the pool after resetting it
func PutFloat64Slice(f []float64) {
	if f != nil {
		// Reset length but keep capacity
		f = f[:0]
		Float64SlicePool.Put(f)
	}
}

// GetBoolSlice retrieves a bool slice from the pool
func GetBoolSlice() []bool {
	return BoolSlicePool.Get().([]bool)
}

// PutBoolSlice returns a bool slice to the pool after resetting it
func PutBoolSlice(b []bool) {
	if b != nil {
		// Reset length but keep capacity
		b = b[:0]
		BoolSlicePool.Put(b)
	}
}

// GetMapStringSlice retrieves a map[string]any slice from the pool
func GetMapStringSlice() []map[string]any {
	return MapStringSlicePool.Get().([]map[string]any)
}

// PutMapStringSlice returns a map[string]any slice to the pool after resetting it
func PutMapStringSlice(m []map[string]any) {
	if m != nil {
		// Reset length but keep capacity
		m = m[:0]
		MapStringSlicePool.Put(m)
	}
}

// JSONMarshal uses pooled buffers for JSON marshaling
func JSONMarshal(v any) ([]byte, error) {
	buf := GetBuffer()
	defer PutBuffer(buf)

	if err := json.NewEncoder(buf).Encode(v); err != nil {
		return nil, err
	}

	// Return a copy of the bytes
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result, nil
}

// JSONUnmarshal uses pooled maps for JSON unmarshaling
func JSONUnmarshal(data []byte, v any) error {
	// For map[string]any targets, use pooled map
	if m, ok := v.(*map[string]any); ok {
		tempMap := GetMap()
		defer PutMap(tempMap)

		// json.Unmarshal needs a pointer, so we pass &tempMap
		if err := json.Unmarshal(data, &tempMap); err != nil {
			return err
		}

		// Copy to target map
		*m = make(map[string]any, len(tempMap))
		for k, val := range tempMap {
			(*m)[k] = val
		}
		return nil
	}

	return json.Unmarshal(data, v)
}

// ExtractField uses pooled resources to extract a field from JSON
func ExtractField(data []byte, key string) (any, error) {
	tempMap := GetMap()
	defer PutMap(tempMap)

	// json.Unmarshal needs a pointer, so we pass &tempMap
	if err := json.Unmarshal(data, &tempMap); err != nil {
		return nil, err
	}

	return tempMap[key], nil
}

// ExtractStringSlice uses pooled resources to extract a string slice from JSON
func ExtractStringSlice(data []byte, key string) ([]string, error) {
	tempMap := GetMap()
	defer PutMap(tempMap)

	if err := json.Unmarshal(data, &tempMap); err != nil {
		return nil, err
	}

	v, ok := tempMap[key]
	if !ok {
		return nil, nil
	}

	arr, ok := v.([]any)
	if !ok {
		return nil, nil
	}

	result := GetStringSlice()
	defer PutStringSlice(result)

	for _, item := range arr {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}

	// Return a copy to avoid pool interference
	copyResult := make([]string, len(result))
	copy(copyResult, result)
	return copyResult, nil
}

// ExtractIntSlice uses pooled resources to extract an int slice from JSON
func ExtractIntSlice(data []byte, key string) ([]int, error) {
	tempMap := GetMap()
	defer PutMap(tempMap)

	if err := json.Unmarshal(data, &tempMap); err != nil {
		return nil, err
	}

	v, ok := tempMap[key]
	if !ok {
		return nil, nil
	}

	arr, ok := v.([]any)
	if !ok {
		return nil, nil
	}

	result := GetIntSlice()
	defer PutIntSlice(result)

	for _, item := range arr {
		switch val := item.(type) {
		case float64:
			result = append(result, int(val))
		case string:
			if i, err := json.Number(val).Int64(); err == nil {
				result = append(result, int(i))
			}
		}
	}

	// Return a copy to avoid pool interference
	copyResult := make([]int, len(result))
	copy(copyResult, result)
	return copyResult, nil
}

// ExtractInt64Slice uses pooled resources to extract an int64 slice from JSON
func ExtractInt64Slice(data []byte, key string) ([]int64, error) {
	tempMap := GetMap()
	defer PutMap(tempMap)

	if err := json.Unmarshal(data, &tempMap); err != nil {
		return nil, err
	}

	v, ok := tempMap[key]
	if !ok {
		return nil, nil
	}

	arr, ok := v.([]any)
	if !ok {
		return nil, nil
	}

	result := GetInt64Slice()
	defer PutInt64Slice(result)

	for _, item := range arr {
		switch val := item.(type) {
		case float64:
			result = append(result, int64(val))
		case string:
			if i, err := json.Number(val).Int64(); err == nil {
				result = append(result, i)
			}
		}
	}

	// Return a copy to avoid pool interference
	copyResult := make([]int64, len(result))
	copy(copyResult, result)
	return copyResult, nil
}

// ExtractFloat64Slice uses pooled resources to extract a float64 slice from JSON
func ExtractFloat64Slice(data []byte, key string) ([]float64, error) {
	tempMap := GetMap()
	defer PutMap(tempMap)

	if err := json.Unmarshal(data, &tempMap); err != nil {
		return nil, err
	}

	v, ok := tempMap[key]
	if !ok {
		return nil, nil
	}

	arr, ok := v.([]any)
	if !ok {
		return nil, nil
	}

	result := GetFloat64Slice()
	defer PutFloat64Slice(result)

	for _, item := range arr {
		switch val := item.(type) {
		case float64:
			result = append(result, val)
		case string:
			if f, err := json.Number(val).Float64(); err == nil {
				result = append(result, f)
			}
		}
	}

	// Return a copy to avoid pool interference
	copyResult := make([]float64, len(result))
	copy(copyResult, result)
	return copyResult, nil
}

// ExtractBoolSlice uses pooled resources to extract a bool slice from JSON
func ExtractBoolSlice(data []byte, key string) ([]bool, error) {
	tempMap := GetMap()
	defer PutMap(tempMap)

	if err := json.Unmarshal(data, &tempMap); err != nil {
		return nil, err
	}

	v, ok := tempMap[key]
	if !ok {
		return nil, nil
	}

	arr, ok := v.([]any)
	if !ok {
		return nil, nil
	}

	result := GetBoolSlice()
	defer PutBoolSlice(result)

	for _, item := range arr {
		if b, ok := item.(bool); ok {
			result = append(result, b)
		}
	}

	// Return a copy to avoid pool interference
	copyResult := make([]bool, len(result))
	copy(copyResult, result)
	return copyResult, nil
}

// ExtractMapString uses pooled resources to extract a map[string]string from JSON
func ExtractMapString(data []byte, key string) (map[string]string, error) {
	tempMap := GetMap()
	defer PutMap(tempMap)

	if err := json.Unmarshal(data, &tempMap); err != nil {
		return nil, err
	}

	v, ok := tempMap[key]
	if !ok {
		return nil, nil
	}

	obj, ok := v.(map[string]any)
	if !ok {
		return nil, nil
	}

	result := make(map[string]string, len(obj))
	for k, val := range obj {
		if s, ok := val.(string); ok {
			result[k] = s
		}
	}

	return result, nil
}

// ExtractMapInt uses pooled resources to extract a map[string]int from JSON
func ExtractMapInt(data []byte, key string) (map[string]int, error) {
	tempMap := GetMap()
	defer PutMap(tempMap)

	if err := json.Unmarshal(data, &tempMap); err != nil {
		return nil, err
	}

	v, ok := tempMap[key]
	if !ok {
		return nil, nil
	}

	obj, ok := v.(map[string]any)
	if !ok {
		return nil, nil
	}

	result := make(map[string]int, len(obj))
	for k, val := range obj {
		switch val := val.(type) {
		case float64:
			result[k] = int(val)
		case string:
			if i, err := json.Number(val).Int64(); err == nil {
				result[k] = int(i)
			}
		}
	}

	return result, nil
}

// ExtractMapInt64 uses pooled resources to extract a map[string]int64 from JSON
func ExtractMapInt64(data []byte, key string) (map[string]int64, error) {
	tempMap := GetMap()
	defer PutMap(tempMap)

	if err := json.Unmarshal(data, &tempMap); err != nil {
		return nil, err
	}

	v, ok := tempMap[key]
	if !ok {
		return nil, nil
	}

	obj, ok := v.(map[string]any)
	if !ok {
		return nil, nil
	}

	result := make(map[string]int64, len(obj))
	for k, val := range obj {
		switch val := val.(type) {
		case float64:
			result[k] = int64(val)
		case string:
			if i, err := json.Number(val).Int64(); err == nil {
				result[k] = i
			}
		}
	}

	return result, nil
}

// ExtractMapFloat64 uses pooled resources to extract a map[string]float64 from JSON
func ExtractMapFloat64(data []byte, key string) (map[string]float64, error) {
	tempMap := GetMap()
	defer PutMap(tempMap)

	if err := json.Unmarshal(data, &tempMap); err != nil {
		return nil, err
	}

	v, ok := tempMap[key]
	if !ok {
		return nil, nil
	}

	obj, ok := v.(map[string]any)
	if !ok {
		return nil, nil
	}

	result := make(map[string]float64, len(obj))
	for k, val := range obj {
		switch val := val.(type) {
		case float64:
			result[k] = val
		case string:
			if f, err := json.Number(val).Float64(); err == nil {
				result[k] = f
			}
		}
	}

	return result, nil
}

// ExtractMapBool uses pooled resources to extract a map[string]bool from JSON
func ExtractMapBool(data []byte, key string) (map[string]bool, error) {
	tempMap := GetMap()
	defer PutMap(tempMap)

	if err := json.Unmarshal(data, &tempMap); err != nil {
		return nil, err
	}

	v, ok := tempMap[key]
	if !ok {
		return nil, nil
	}

	obj, ok := v.(map[string]any)
	if !ok {
		return nil, nil
	}

	result := make(map[string]bool, len(obj))
	for k, val := range obj {
		if b, ok := val.(bool); ok {
			result[k] = b
		}
	}

	return result, nil
}

// ExtractArrayMapString uses pooled resources to extract an array of map[string]string from JSON
func ExtractArrayMapString(data []byte, key string) ([]map[string]string, error) {
	tempMap := GetMap()
	defer PutMap(tempMap)

	if err := json.Unmarshal(data, &tempMap); err != nil {
		return nil, err
	}

	v, ok := tempMap[key]
	if !ok {
		return nil, nil
	}

	arr, ok := v.([]any)
	if !ok {
		return nil, nil
	}

	result := make([]map[string]string, 0, len(arr))

	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		m := make(map[string]string, len(obj))
		for k, val := range obj {
			if s, ok := val.(string); ok {
				m[k] = s
			}
		}
		result = append(result, m)
	}

	return result, nil
}

// ExtractArrayMapInt uses pooled resources to extract an array of map[string]int from JSON
func ExtractArrayMapInt(data []byte, key string) ([]map[string]int, error) {
	tempMap := GetMap()
	defer PutMap(tempMap)

	if err := json.Unmarshal(data, &tempMap); err != nil {
		return nil, err
	}

	v, ok := tempMap[key]
	if !ok {
		return nil, nil
	}

	arr, ok := v.([]any)
	if !ok {
		return nil, nil
	}

	result := make([]map[string]int, 0, len(arr))

	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		m := make(map[string]int, len(obj))
		for k, val := range obj {
			switch val := val.(type) {
			case float64:
				m[k] = int(val)
			case string:
				if i, err := json.Number(val).Int64(); err == nil {
					m[k] = int(i)
				}
			}
		}
		result = append(result, m)
	}

	return result, nil
}

// ExtractArrayMapInt64 uses pooled resources to extract an array of map[string]int64 from JSON
func ExtractArrayMapInt64(data []byte, key string) ([]map[string]int64, error) {
	tempMap := GetMap()
	defer PutMap(tempMap)

	if err := json.Unmarshal(data, &tempMap); err != nil {
		return nil, err
	}

	v, ok := tempMap[key]
	if !ok {
		return nil, nil
	}

	arr, ok := v.([]any)
	if !ok {
		return nil, nil
	}

	result := make([]map[string]int64, 0, len(arr))

	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		m := make(map[string]int64, len(obj))
		for k, val := range obj {
			switch val := val.(type) {
			case float64:
				m[k] = int64(val)
			case string:
				if i, err := json.Number(val).Int64(); err == nil {
					m[k] = i
				}
			}
		}
		result = append(result, m)
	}

	return result, nil
}

// ExtractArrayMapFloat64 uses pooled resources to extract an array of map[string]float64 from JSON
func ExtractArrayMapFloat64(data []byte, key string) ([]map[string]float64, error) {
	tempMap := GetMap()
	defer PutMap(tempMap)

	if err := json.Unmarshal(data, &tempMap); err != nil {
		return nil, err
	}

	v, ok := tempMap[key]
	if !ok {
		return nil, nil
	}

	arr, ok := v.([]any)
	if !ok {
		return nil, nil
	}

	result := make([]map[string]float64, 0, len(arr))

	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		m := make(map[string]float64, len(obj))
		for k, val := range obj {
			switch val := val.(type) {
			case float64:
				m[k] = val
			case string:
				if f, err := json.Number(val).Float64(); err == nil {
					m[k] = f
				}
			}
		}
		result = append(result, m)
	}

	return result, nil
}

// ExtractArrayMapBool uses pooled resources to extract an array of map[string]bool from JSON
func ExtractArrayMapBool(data []byte, key string) ([]map[string]bool, error) {
	tempMap := GetMap()
	defer PutMap(tempMap)

	if err := json.Unmarshal(data, &tempMap); err != nil {
		return nil, err
	}

	v, ok := tempMap[key]
	if !ok {
		return nil, nil
	}

	arr, ok := v.([]any)
	if !ok {
		return nil, nil
	}

	result := make([]map[string]bool, 0, len(arr))

	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		m := make(map[string]bool, len(obj))
		for k, val := range obj {
			if b, ok := val.(bool); ok {
				m[k] = b
			}
		}
		result = append(result, m)
	}

	return result, nil
}
