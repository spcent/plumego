package sharding

import (
	"fmt"
	"testing"
)

// BenchmarkHashStrategy benchmarks the hash-based sharding strategy.
func BenchmarkHashStrategy(b *testing.B) {
	strategy := NewHashStrategy()
	numShards := 4

	b.Run("String", func(b *testing.B) {
		key := "user123"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = strategy.Shard(key, numShards)
		}
	})

	b.Run("Int64", func(b *testing.B) {
		key := int64(12345)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = strategy.Shard(key, numShards)
		}
	})

	b.Run("Uint64", func(b *testing.B) {
		key := uint64(12345)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = strategy.Shard(key, numShards)
		}
	})

	b.Run("ByteSlice", func(b *testing.B) {
		key := []byte("user123")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = strategy.Shard(key, numShards)
		}
	})
}

// BenchmarkModStrategy benchmarks the modulo-based sharding strategy.
func BenchmarkModStrategy(b *testing.B) {
	strategy := NewModStrategy()
	numShards := 4

	b.Run("Int64", func(b *testing.B) {
		key := int64(12345)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = strategy.Shard(key, numShards)
		}
	})

	b.Run("Int", func(b *testing.B) {
		key := 12345
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = strategy.Shard(key, numShards)
		}
	})

	b.Run("Uint64", func(b *testing.B) {
		key := uint64(12345)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = strategy.Shard(key, numShards)
		}
	})
}

// BenchmarkRangeStrategy benchmarks the range-based sharding strategy.
func BenchmarkRangeStrategy(b *testing.B) {
	ranges := []RangeDefinition{
		{Start: int64(0), End: int64(10000), Shard: 0},
		{Start: int64(10000), End: int64(20000), Shard: 1},
		{Start: int64(20000), End: int64(30000), Shard: 2},
		{Start: int64(30000), End: int64(40000), Shard: 3},
	}

	strategy, err := NewRangeStrategy(ranges)
	if err != nil {
		b.Fatalf("NewRangeStrategy() error = %v", err)
	}

	numShards := 4

	b.Run("FirstRange", func(b *testing.B) {
		key := int64(5000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = strategy.Shard(key, numShards)
		}
	})

	b.Run("MiddleRange", func(b *testing.B) {
		key := int64(25000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = strategy.Shard(key, numShards)
		}
	})

	b.Run("LastRange", func(b *testing.B) {
		key := int64(35000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = strategy.Shard(key, numShards)
		}
	})
}

// BenchmarkRangeStrategyScalability tests range strategy with different numbers of ranges.
func BenchmarkRangeStrategyScalability(b *testing.B) {
	testSizes := []int{4, 16, 64, 256}

	for _, size := range testSizes {
		b.Run(fmt.Sprintf("Ranges%d", size), func(b *testing.B) {
			ranges := make([]RangeDefinition, size)
			rangeSize := int64(10000)
			for i := 0; i < size; i++ {
				ranges[i] = RangeDefinition{
					Start: int64(i) * rangeSize,
					End:   int64(i+1) * rangeSize,
					Shard: i % 4, // Distribute across 4 shards
				}
			}

			strategy, err := NewRangeStrategy(ranges)
			if err != nil {
				b.Fatalf("NewRangeStrategy() error = %v", err)
			}

			key := int64(size/2) * rangeSize // Middle key
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = strategy.Shard(key, 4)
			}
		})
	}
}

// BenchmarkListStrategy benchmarks the list-based sharding strategy.
func BenchmarkListStrategy(b *testing.B) {
	mapping := map[any]int{
		"US": 0,
		"EU": 1,
		"CN": 2,
		"JP": 3,
		"UK": 4,
		"DE": 5,
		"FR": 6,
		"AU": 7,
	}

	strategy := NewListStrategy(mapping)
	numShards := 8

	b.Run("ExistingKey", func(b *testing.B) {
		key := "US"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = strategy.Shard(key, numShards)
		}
	})

	b.Run("WithDefault", func(b *testing.B) {
		strategyWithDefault := NewListStrategyWithDefault(mapping, 0)
		key := "XX" // Non-existent key
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = strategyWithDefault.Shard(key, numShards)
		}
	})
}

// BenchmarkListStrategyScalability tests list strategy with different mapping sizes.
func BenchmarkListStrategyScalability(b *testing.B) {
	testSizes := []int{10, 100, 1000, 10000}

	for _, size := range testSizes {
		b.Run(fmt.Sprintf("Mappings%d", size), func(b *testing.B) {
			mapping := make(map[any]int)
			for i := 0; i < size; i++ {
				mapping[fmt.Sprintf("key%d", i)] = i % 4
			}

			strategy := NewListStrategy(mapping)
			key := fmt.Sprintf("key%d", size/2) // Middle key
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = strategy.Shard(key, 4)
			}
		})
	}
}

// BenchmarkAllStrategies compares performance of all strategies.
func BenchmarkAllStrategies(b *testing.B) {
	numShards := 4

	// Hash strategy
	hashStrategy := NewHashStrategy()
	b.Run("Hash", func(b *testing.B) {
		key := "user123"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = hashStrategy.Shard(key, numShards)
		}
	})

	// Mod strategy
	modStrategy := NewModStrategy()
	b.Run("Mod", func(b *testing.B) {
		key := int64(12345)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = modStrategy.Shard(key, numShards)
		}
	})

	// Range strategy
	ranges := []RangeDefinition{
		{Start: int64(0), End: int64(10000), Shard: 0},
		{Start: int64(10000), End: int64(20000), Shard: 1},
		{Start: int64(20000), End: int64(30000), Shard: 2},
		{Start: int64(30000), End: int64(40000), Shard: 3},
	}
	rangeStrategy, _ := NewRangeStrategy(ranges)
	b.Run("Range", func(b *testing.B) {
		key := int64(15000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = rangeStrategy.Shard(key, numShards)
		}
	})

	// List strategy
	mapping := map[any]int{
		"US": 0,
		"EU": 1,
		"CN": 2,
		"JP": 3,
	}
	listStrategy := NewListStrategy(mapping)
	b.Run("List", func(b *testing.B) {
		key := "US"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = listStrategy.Shard(key, numShards)
		}
	})
}

// BenchmarkShardRange benchmarks range query performance.
func BenchmarkShardRange(b *testing.B) {
	numShards := 4

	b.Run("Hash", func(b *testing.B) {
		strategy := NewHashStrategy()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = strategy.ShardRange("start", "end", numShards)
		}
	})

	b.Run("Mod", func(b *testing.B) {
		strategy := NewModStrategy()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = strategy.ShardRange(int64(10), int64(50), numShards)
		}
	})

	ranges := []RangeDefinition{
		{Start: int64(0), End: int64(10000), Shard: 0},
		{Start: int64(10000), End: int64(20000), Shard: 1},
		{Start: int64(20000), End: int64(30000), Shard: 2},
		{Start: int64(30000), End: int64(40000), Shard: 3},
	}
	rangeStrategy, _ := NewRangeStrategy(ranges)
	b.Run("Range", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = rangeStrategy.ShardRange(int64(5000), int64(25000), numShards)
		}
	})

	mapping := map[any]int{
		"US": 0,
		"EU": 1,
		"CN": 2,
		"JP": 3,
	}
	listStrategy := NewListStrategy(mapping)
	b.Run("List", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = listStrategy.ShardRange("start", "end", numShards)
		}
	})
}
