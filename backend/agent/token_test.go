package agent

import (
	"testing"
)

func TestDefaultEstimator_Estimate(t *testing.T) {
	estimator := NewTokenEstimator()

	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "空字符串",
			text:     "",
			expected: 0,
		},
		{
			name:     "纯英文短句",
			text:     "Hello",
			expected: 8, // 1 word + 4 overhead
		},
		{
			name:     "纯英文句子",
			text:     "Hello World",
			expected: 12, // 2 words * 0.75 + 4 overhead
		},
		{
			name:     "纯中文",
			text:     "你好世界",
			expected: 12, // 4 chars * 2 + 4 overhead
		},
		{
			name:     "中英文混合",
			text:     "Hello 你好",
			expected: 12,
		},
		{
			name:     "包含数字",
			text:     "12345",
			expected: 7, // 5 digits * 0.5 + 4 overhead
		},
		{
			name:     "包含标点",
			text:     "Hello, World!",
			expected: 13,
		},
		{
			name:     "长文本",
			text:     "这是一个比较长的中文文本，用于测试Token估算算法的准确性。",
			expected: 47,
		},
		{
			name:     "代码片段",
			text:     "func main() { fmt.Println(\"Hello\") }",
			expected: 27,
		},
		{
			name:     "日文",
			text:     "こんにちは",
			expected: 14, // 5 chars * 2 + 4 overhead
		},
		{
			name:     "韩文",
			text:     "안녕하세요",
			expected: 14, // 5 chars * 2 + 4 overhead
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := estimator.Estimate(tt.text)
			// 允许一定的误差范围
			diff := got - tt.expected
			if diff < -2 || diff > 2 {
				t.Errorf("Estimate() = %v, expected around %v", got, tt.expected)
			}
			t.Logf("Estimate(%q) = %d", tt.text, got)
		})
	}
}

func TestIsCJK(t *testing.T) {
	tests := []struct {
		name     string
		r        rune
		expected bool
	}{
		{
			name:     "中文",
			r:        '中',
			expected: true,
		},
		{
			name:     "英文",
			r:        'A',
			expected: false,
		},
		{
			name:     "日文平假名",
			r:        'あ',
			expected: true,
		},
		{
			name:     "日文片假名",
			r:        'ア',
			expected: true,
		},
		{
			name:     "韩文",
			r:        '한',
			expected: true,
		},
		{
			name:     "数字",
			r:        '1',
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCJK(tt.r)
			if got != tt.expected {
				t.Errorf("isCJK(%q) = %v, want %v", tt.r, got, tt.expected)
			}
		})
	}
}

func BenchmarkEstimator_Estimate(b *testing.B) {
	estimator := NewTokenEstimator()
	text := "这是一个用于性能测试的文本，包含中英文混合内容。This is a mixed Chinese and English text for performance testing."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		estimator.Estimate(text)
	}
}

func BenchmarkEstimateTokensFunc(b *testing.B) {
	text := "这是一个用于性能测试的文本，包含中英文混合内容。This is a mixed Chinese and English text for performance testing."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		estimateTokens(text)
	}
}
