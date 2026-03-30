package interview

import "strings"

var questionHints = []string{
	"什么",
	"怎么",
	"为什么",
	"如何",
	"哪些",
	"哪个",
	"能否",
	"是否",
	"可以",
	"讲讲",
	"介绍一下",
	"tell me",
	"what",
	"why",
	"how",
	"when",
	"where",
	"walk me through",
	"could you",
	"can you",
}

func LooksLikeQuestion(text string) bool {
	normalized := strings.TrimSpace(strings.ToLower(text))
	if normalized == "" {
		return false
	}

	if strings.HasSuffix(normalized, "?") || strings.HasSuffix(normalized, "？") {
		return true
	}

	if len([]rune(normalized)) < 6 {
		return false
	}

	for _, hint := range questionHints {
		if strings.Contains(normalized, hint) {
			return true
		}
	}

	return false
}
