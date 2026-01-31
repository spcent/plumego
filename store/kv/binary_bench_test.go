package kvstore

import (
	"bufio"
	"bytes"
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
		serializer := &BinarySerializer{}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, err := serializer.EncodeWALEntry(entry)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("JSON_Encode", func(b *testing.B) {
		serializer := &JSONSerializer{}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, err := serializer.EncodeWALEntry(entry)
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

	binarySerializer := &BinarySerializer{}
	jsonSerializer := &JSONSerializer{}

	// Prepare binary data
	binaryData, _ := binarySerializer.EncodeWALEntry(entry)

	// Prepare JSON data
	jsonData, _ := jsonSerializer.EncodeWALEntry(entry)

	b.Run("Binary_Decode", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			reader := bufio.NewReader(bytes.NewReader(binaryData))
			_, err := binarySerializer.DecodeWALEntry(reader)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("JSON_Decode", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			reader := bufio.NewReader(bytes.NewReader(jsonData))
			_, err := jsonSerializer.DecodeWALEntry(reader)
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
		serializer := &BinarySerializer{}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, err := serializer.EncodeEntry(entry)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("JSON_Encode", func(b *testing.B) {
		serializer := &JSONSerializer{}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, err := serializer.EncodeEntry(entry)
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

	binarySerializer := &BinarySerializer{}
	jsonSerializer := &JSONSerializer{}

	// Prepare binary data
	binaryData, _ := binarySerializer.EncodeEntry(entry)

	// Prepare JSON data
	jsonData, _ := jsonSerializer.EncodeEntry(entry)

	b.Run("Binary_Decode", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			reader := bufio.NewReader(bytes.NewReader(binaryData))
			_, err := binarySerializer.DecodeEntry(reader)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("JSON_Decode", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			reader := bufio.NewReader(bytes.NewReader(jsonData))
			_, err := jsonSerializer.DecodeEntry(reader)
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

	binarySerializer := &BinarySerializer{}
	jsonSerializer := &JSONSerializer{}

	// Get binary size
	binaryData, _ := binarySerializer.EncodeWALEntry(entry)

	// Get JSON size
	jsonData, _ := jsonSerializer.EncodeWALEntry(entry)

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

	b.Run("Binary_Bulk_Encode", func(b *testing.B) {
		serializer := &BinarySerializer{}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, entry := range entries {
				_, err := serializer.EncodeWALEntry(entry)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})

	b.Run("JSON_Bulk_Encode", func(b *testing.B) {
		serializer := &JSONSerializer{}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, entry := range entries {
				_, err := serializer.EncodeWALEntry(entry)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}
