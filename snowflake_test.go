package snowflake

import (
	"testing"
	"time"
)

func TestOption(t *testing.T) {
	epoch := time.Date(2021, 01, 01, 0, 0, 0, 0, time.UTC)
	snowflake, _ := NewSnowflake(
		WithEpoch(epoch),
		WithWorkerIdBits(9),
		WithSequenceBits(11),
	)

	if snowflake.epoch != epoch.UnixMilli() {
		t.Fatalf("WithEpoch not work")
	}

	if snowflake.workerIdBits != 9 {
		t.Fatalf("WithWorkerIdBits not work")
	}

	if snowflake.sequenceBits != 11 {
		t.Fatalf("WithSequenceBits not work")
	}
}

func TestNewNode(t *testing.T) {
	_, err := NewSnowflake(WithWorkerId(0))
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	_, err = NewSnowflake(WithWorkerId(5000))
	if err == nil {
		t.Fatalf("no error creating NewNode, %s", err)
	}

}

// lazy check if Generate will create duplicate IDs
// would be good to later enhance this with more smarts
func TestGenerateDuplicateID(t *testing.T) {

	node, _ := NewSnowflake(WithWorkerId(0))

	var x, y int64
	for i := 0; i < 1000000; i++ {
		y = node.Next()
		if x == y {
			t.Errorf("x(%d) & y(%d) are the same", x, y)
		}
		x = y
	}
}

// I feel like there's probably a better way
func TestRace(t *testing.T) {

	node, _ := NewSnowflake(WithWorkerId(1))

	go func() {
		for i := 0; i < 1000000000; i++ {
			_, _ = NewSnowflake(WithWorkerId(1))
		}
	}()

	for i := 0; i < 4000; i++ {
		node.Next()
	}

}

//func BenchmarkGenerate(b *testing.B) {
//
//	vuuvv, _ := NewSnowflake()
//	bwmarrin, _ := snowflake.NewNode(1)
//
//	b.ReportAllocs()
//
//	b.ResetTimer()
//
//	b.Run("snowflake", func(b *testing.B) {
//		b.Run("bwmarrin", func(b *testing.B) {
//			for i := 0; i < b.N; i++ {
//				_ = bwmarrin.Generate()
//			}
//		})
//		b.Run("vuuvv", func(b *testing.B) {
//			for i := 0; i < b.N; i++ {
//				_ = vuuvv.Next()
//			}
//		})
//	})
//}
