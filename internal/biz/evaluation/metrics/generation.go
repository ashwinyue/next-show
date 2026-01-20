// Package metrics 提供生成质量评估指标.
package metrics

import (
	"fmt"
	"math"
	"strings"
)

// BLEUMetric BLEU 指标（用于评估机器翻译质量）.
type BLEUMetric struct {
	MaxN int // 最大 n-gram, 通常为 4
}

// NewBLEUMetric 创建 BLEU 指标.
func NewBLEUMetric(maxN int) *BLEUMetric {
	if maxN <= 0 {
		maxN = 4
	}
	return &BLEUMetric{MaxN: maxN}
}

func (m *BLEUMetric) Name() string {
	return fmt.Sprintf("bleu-%d", m.MaxN)
}

func (m *BLEUMetric) Compute(input *MetricInput) float64 {
	candidate := m.splitIntoWords(input.GeneratedText)
	reference := m.splitIntoWords(input.ExpectedText)

	if len(candidate) == 0 || len(reference) == 0 {
		return 0.0
	}

	// 计算 modified precision
	precisions := make([]float64, m.MaxN)
	for n := 1; n <= m.MaxN; n++ {
		precisions[n-1] = m.modifiedPrecision(candidate, reference, n)
	}

	// 计算 geometric mean
	logSum := 0.0
	weight := 1.0 / float64(m.MaxN)
	for _, p := range precisions {
		if p > 0 {
			logSum += weight * math.Log(p)
		}
	}

	// Brevity penalty
	bp := m.brevityPenalty(len(candidate), len(reference))

	return bp * math.Exp(logSum)
}

func (m *BLEUMetric) splitIntoWords(text string) []string {
	return strings.Fields(strings.ToLower(text))
}

func (m *BLEUMetric) modifiedPrecision(candidate, reference []string, n int) float64 {
	if n > len(candidate) || n > len(reference) {
		return 0.0
	}

	// 提取 n-grams
	candidateNGrams := m.getNGrams(candidate, n)
	referenceNGrams := m.getNGrams(reference, n)

	if len(candidateNGrams) == 0 {
		return 0.0
	}

	// 计算 clipped counts
	clippedCount := 0
	for ngram, count := range candidateNGrams {
		refCount := referenceNGrams[ngram]
		if refCount > count {
			clippedCount += refCount
		} else {
			clippedCount += count
		}
	}

	return float64(clippedCount) / float64(len(candidateNGrams))
}

func (m *BLEUMetric) getNGrams(words []string, n int) map[string]int {
	ngrams := make(map[string]int)
	for i := 0; i <= len(words)-n; i++ {
		ngram := strings.Join(words[i:i+n], " ")
		ngrams[ngram]++
	}
	return ngrams
}

func (m *BLEUMetric) brevityPenalty(candidateLen, referenceLen int) float64 {
	if candidateLen > referenceLen {
		return 1.0
	}
	if candidateLen == 0 {
		return 0.0
	}
	return math.Exp(1.0 - float64(referenceLen)/float64(candidateLen))
}

func (m *BLEUMetric) Validate(input *MetricInput) error {
	if input.GeneratedText == "" || input.ExpectedText == "" {
		return fmt.Errorf("generated_text and expected_text are required")
	}
	return nil
}

func (m *BLEUMetric) IsGenerationMetric() {}

// ROUGEMetric ROUGE 指标（用于评估摘要质量）.
type ROUGEMetric struct {
	ROUGETYPE ROUGETYPE
}

// ROUGETYPE ROUGE 类型.
type ROUGETYPE string

const (
	ROUGE1 ROUGETYPE = "rouge1" // Unigrams
	ROUGE2 ROUGETYPE = "rouge2" // Bigrams
	ROUGEL ROUGETYPE = "rougel" // Longest Common Subsequence
)

// NewROUGEMetric 创建 ROUGE 指标.
func NewROUGEMetric(rougeType ROUGETYPE) *ROUGEMetric {
	return &ROUGEMetric{ROUGETYPE: rougeType}
}

func (m *ROUGEMetric) Name() string {
	return string(m.ROUGETYPE)
}

func (m *ROUGEMetric) Compute(input *MetricInput) float64 {
	candidate := m.splitIntoWords(input.GeneratedText)
	reference := m.splitIntoWords(input.ExpectedText)

	if len(candidate) == 0 || len(reference) == 0 {
		return 0.0
	}

	switch m.ROUGETYPE {
	case ROUGE1:
		return m.computeROUGEN(candidate, reference, 1)
	case ROUGE2:
		return m.computeROUGEN(candidate, reference, 2)
	case ROUGEL:
		return m.computeROUGEL(candidate, reference)
	default:
		return 0.0
	}
}

func (m *ROUGEMetric) computeROUGEN(candidate, reference []string, n int) float64 {
	candidateNGrams := m.getNGrams(candidate, n)
	referenceNGrams := m.getNGrams(reference, n)

	if len(candidateNGrams) == 0 {
		return 0.0
	}

	// 计算重叠
	matchCount := 0
	for ngram, count := range candidateNGrams {
		if refCount, exists := referenceNGrams[ngram]; exists {
			if refCount > count {
				matchCount += count
			} else {
				matchCount += refCount
			}
		}
	}

	// ROUGE-N = match / max(|candidate|, |reference|)
	denominator := len(candidateNGrams)
	if len(referenceNGrams) > len(candidateNGrams) {
		denominator = len(referenceNGrams)
	}

	return float64(matchCount) / float64(denominator)
}

func (m *ROUGEMetric) computeROUGEL(candidate, reference []string) float64 {
	// 计算最长公共子序列
	lcs := m.longestCommonSubsequence(candidate, reference)

	if len(candidate) == 0 || len(reference) == 0 {
		return 0.0
	}

	// ROUGE-L = LCS / max(|candidate|, |reference|)
	denominator := len(candidate)
	if len(reference) > len(candidate) {
		denominator = len(reference)
	}

	return float64(lcs) / float64(denominator)
}

func (m *ROUGEMetric) longestCommonSubsequence(a, b []string) int {
	mLen, n := len(a), len(b)
	dp := make([][]int, mLen+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= mLen; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	return dp[mLen][n]
}

func (m *ROUGEMetric) splitIntoWords(text string) []string {
	return strings.Fields(strings.ToLower(text))
}

func (m *ROUGEMetric) getNGrams(words []string, n int) map[string]int {
	ngrams := make(map[string]int)
	for i := 0; i <= len(words)-n; i++ {
		ngram := strings.Join(words[i:i+n], " ")
		ngrams[ngram]++
	}
	return ngrams
}

func (m *ROUGEMetric) Validate(input *MetricInput) error {
	if input.GeneratedText == "" || input.ExpectedText == "" {
		return fmt.Errorf("generated_text and expected_text are required")
	}
	return nil
}

func (m *ROUGEMetric) IsGenerationMetric() {}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
