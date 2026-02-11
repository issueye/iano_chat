package iano_agent

import (
	"math"
	"unicode"
	"unicode/utf8"
)

// TokenEstimator Token 估算器接口
type TokenEstimator interface {
	Estimate(text string) int
}

// defaultEstimator 默认 Token 估算器
type defaultEstimator struct{}

// NewTokenEstimator 创建新的 Token 估算器
func NewTokenEstimator() TokenEstimator {
	return &defaultEstimator{}
}

// Estimate 估算文本的 Token 数量
// 使用改进的算法，参考 OpenAI 的 Token 计算规则：
// 1. 英文单词：平均 0.75 tokens/词
// 2. 中文字符：约 2 tokens/字符
// 3. 数字和标点：约 0.5 tokens/字符
// 4. 代码和特殊字符：约 1-2 tokens/字符
func (e *defaultEstimator) Estimate(text string) int {
	if text == "" {
		return 0
	}

	var tokenCount float64
	var i int

	for i < len(text) {
		r, size := utf8.DecodeRuneInString(text[i:])
		if r == utf8.RuneError && size == 1 {
			// 无效 UTF-8 字符，按 1 字节处理
			tokenCount += 0.5
			i++
			continue
		}

		switch {
		// CJK 字符（中文、日文、韩文）
		case isCJK(r):
			tokenCount += 2.0

		// 英文单词字符
		case unicode.IsLetter(r):
			// 读取完整单词
			wordLen := 0
			for i < len(text) {
				r2, size2 := utf8.DecodeRuneInString(text[i:])
				if !unicode.IsLetter(r2) {
					break
				}
				wordLen++
				i += size2
			}
			// 英文单词平均 0.75 tokens/字符
			tokenCount += math.Max(1, float64(wordLen)*0.75)
			continue

		// 数字
		case unicode.IsNumber(r):
			// 数字序列
			numLen := 0
			for i < len(text) {
				r2, size2 := utf8.DecodeRuneInString(text[i:])
				if !unicode.IsNumber(r2) {
					break
				}
				numLen++
				i += size2
			}
			// 数字平均 0.5 tokens/字符
			tokenCount += math.Max(1, float64(numLen)*0.5)
			continue

		// 空白字符
		case unicode.IsSpace(r):
			// 空白字符通常不计入 token，但连续的空白可能产生 1 个 token
			if i > 0 {
				prevR, _ := utf8.DecodeRuneInString(text[i-size:])
				if !unicode.IsSpace(prevR) {
					tokenCount += 0.1
				}
			}

		// 标点符号
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			tokenCount += 0.5

		// 其他字符
		default:
			tokenCount += 1.0
		}

		i += size
	}

	// 添加基础开销（每个消息约 3-4 tokens 的格式开销）
	tokenCount += 4

	return int(math.Ceil(tokenCount))
}

// isCJK 检查字符是否为 CJK（中日韩）字符
func isCJK(r rune) bool {
	// CJK 统一表意文字范围
	if r >= 0x4E00 && r <= 0x9FFF {
		return true
	}
	// CJK 扩展 A
	if r >= 0x3400 && r <= 0x4DBF {
		return true
	}
	// CJK 扩展 B
	if r >= 0x20000 && r <= 0x2A6DF {
		return true
	}
	// 日文平假名
	if r >= 0x3040 && r <= 0x309F {
		return true
	}
	// 日文片假名
	if r >= 0x30A0 && r <= 0x30FF {
		return true
	}
	// 韩文
	if r >= 0xAC00 && r <= 0xD7AF {
		return true
	}
	return false
}

// SimpleEstimateTokens 简单的 Token 估算（向后兼容）
func estimateTokens(text string) int {
	estimator := NewTokenEstimator()
	return estimator.Estimate(text)
}

// EstimateMessages 估算消息列表的总 Token 数
func EstimateMessages(messages []*ConversationRound) int {
	estimator := NewTokenEstimator()
	total := 0

	for _, round := range messages {
		if round.UserMessage != nil {
			total += estimator.Estimate(round.UserMessage.Content)
		}
		if round.AssistantMessage != nil {
			total += estimator.Estimate(round.AssistantMessage.Content)
		}
	}

	return total
}
