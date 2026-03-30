package interview

import "testing"

func TestLooksLikeQuestion(t *testing.T) {
	cases := []struct {
		text string
		want bool
	}{
		{text: "你介绍一下你在上家公司做的缓存项目", want: true},
		{text: "What trade-offs did you consider?", want: true},
		{text: "我主要负责 Go 后端开发", want: false},
		{text: "好的", want: false},
		{text: "为什么选择用 Redis", want: true},
	}

	for _, tc := range cases {
		if got := LooksLikeQuestion(tc.text); got != tc.want {
			t.Fatalf("LooksLikeQuestion(%q) = %v, want %v", tc.text, got, tc.want)
		}
	}
}
