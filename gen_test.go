package snow

import "testing"

func TestSnow(t *testing.T) {
	snow, err := NewSnow(1, 1)
	if err != nil {
		t.Fatalf("init snow error")
	}

	id, err := snow.Gen()
	if err != nil {
		t.Fatalf("gen error")
	}

	t.Logf("id:%d", id)
}

func BenchmarkSnow(b *testing.B) {
	snow, err := NewSnow(1, 1)
	if err != nil {
		return
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		id, err := snow.Gen()
		if err != nil || id < 0 {
			return
		}
	}
}
