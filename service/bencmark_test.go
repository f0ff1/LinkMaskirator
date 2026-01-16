package service

import (
	"strings"
	"testing"
)

func BenchmarkMaskLinks_Short(b *testing.B) {
	input := "http://example.com"
	for i := 0; i < b.N; i++ {
		maskLink(input)
	}
}

func BenchmarkMaskLinks_LongText(b *testing.B) {
	input := strings.Repeat("текст с http://link.com ", 100) // 100 ссылок
	for i := 0; i < b.N; i++ {
		maskLink(input)
	}
}

func BenchmarkMaskLinks_NoLinks(b *testing.B) {
	input := "просто текст без ссылок"
	for i := 0; i < b.N; i++ {
		maskLink(input)
	}
}
func BenchmarkMaskLinks_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		input := "текст с http://ccumxoii"
		for pb.Next() {
			maskLink(input)
		}
	})
}
