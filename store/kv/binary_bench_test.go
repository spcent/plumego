package kvstore

import (
	"bufio"
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

// Benchmark binary vs JSON serialization for WAL entries

func BenchmarkWALEntrySerialization(b *testing.B) {
	entry := WALEntry{
		Op:       opSet,
		Key:      "benchmark_key_with_reasonable_length_0001",
		Value:    make([]byte, 1024), // 1KB value
		ExpireAt: time.Now().Add(time.Hour),
		Version:  12345,
		CRC:      0x12345678,
	}

	b.Run("Binary_Encode", func(b *testing.B) {
		kv := &KVStore{}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, err := kv.encodeWALEntryBinary(entry)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("JSON_Encode", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			buf := new(bytes.Buffer)
			err := json.NewEncoder(buf).Encode(entry)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkWALEntryDeserialization(b *testing.B) {
	entry := WALEntry{
		Op:       opSet,
		Key:      "benchmark_key_with_reasonable_length_0001",
		Value:    make([]byte, 1024), // 1KB value
		ExpireAt: time.Now().Add(time.Hour),
		Version:  12345,
		CRC:      0x12345678,
	}

	kv := &KVStore{}

	// Prepare binary data
	binaryData, _ := kv.encodeWALEntryBinary(entry)

	// Prepare JSON data
	jsonBuf := new(bytes.Buffer)
	json.NewEncoder(jsonBuf).Encode(entry)
	jsonData := jsonBuf.Bytes()

	b.Run("Binary_Decode", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			reader := bufio.NewReader(bytes.NewReader(binaryData))
			_, err := kv.decodeWALEntryBinary(reader)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("JSON_Decode", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			reader := bytes.NewReader(jsonData)
			var decoded WALEntry
			err := json.NewDecoder(reader).Decode(&decoded)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkEntrySerialization(b *testing.B) {
	entry := &Entry{
		Key:      "benchmark_key_with_reasonable_length_0001",
		Value:    make([]byte, 1024), // 1KB value
		ExpireAt: time.Now().Add(time.Hour),
		Version:  12345,
		Size:     1088,
	}

	b.Run("Binary_Encode", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, err := encodeEntryBinary(entry)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("JSON_Encode", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			buf := new(bytes.Buffer)
			err := json.NewEncoder(buf).Encode(entry)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkEntryDeserialization(b *testing.B) {
	entry := &Entry{
		Key:      "benchmark_key_with_reasonable_length_0001",
		Value:    make([]byte, 1024), // 1KB value
		ExpireAt: time.Now().Add(time.Hour),
		Version:  12345,
		Size:     1088,
	}

	// Prepare binary data
	binaryData, _ := encodeEntryBinary(entry)

	// Prepare JSON data
	jsonBuf := new(bytes.Buffer)
	json.NewEncoder(jsonBuf).Encode(entry)
	jsonData := jsonBuf.Bytes()

	b.Run("Binary_Decode", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			reader := bufio.NewReader(bytes.NewReader(binaryData))
			_, err := decodeEntryBinary(reader)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("JSON_Decode", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			reader := bytes.NewReader(jsonData)
			var decoded Entry
			err := json.NewDecoder(reader).Decode(&decoded)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// Benchmark data size comparison
func BenchmarkSerializedSize(b *testing.B) {
	entry := WALEntry{
		Op:       opSet,
		Key:      "benchmark_key_with_reasonable_length_0001",
		Value:    make([]byte, 1024),
		ExpireAt: time.Now().Add(time.Hour),
		Version:  12345,
		CRC:      0x12345678,
	}

	kv := &KVStore{}

	// Get binary size
	binaryData, _ := kv.encodeWALEntryBinary(entry)

	// Get JSON size
	jsonBuf := new(bytes.Buffer)
	json.NewEncoder(jsonBuf).Encode(entry)
	jsonData := jsonBuf.Bytes()

	b.Logf("Binary size: %d bytes", len(binaryData))
	b.Logf("JSON size: %d bytes", len(jsonData))
	b.Logf("Size reduction: %.2f%%", float64(len(jsonData)-len(binaryData))/float64(len(jsonData))*100)
}

// Benchmark realistic workload: multiple entries
func BenchmarkBulkSerialization(b *testing.B) {
	entries := make([]WALEntry, 1000)
	for i := 0; i < 1000; i++ {
		entries[i] = WALEntry{
			Op:       opSet,
			Key:      "key_" + string(rune(i)),
			Value:    make([]byte, 256),
			ExpireAt: time.Now().Add(time.Hour),
			Version:  int64(i),
			CRC:      uint32(i),
		}
	}

	kv := &KVStore{}

	b.Run("Binary_Bulk_Encode", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, entry := range entries {
				_, err := kv.encodeWALEntryBinary(entry)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})

	b.Run("JSON_Bulk_Encode", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, entry := range entries {
				buf := new(bytes.Buffer)
				err := json.NewEncoder(buf).Encode(entry)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}
