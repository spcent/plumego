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
