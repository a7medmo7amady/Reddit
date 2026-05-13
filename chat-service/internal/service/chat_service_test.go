package service

import (
	"testing"
)

func TestExtractMentions(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "SingleMention",
			content: "hey @john what's up",
			want:    []string{"john"},
		},
		{
			name:    "MultipleMentions",
			content: "@alice and @bob are here",
			want:    []string{"alice", "bob"},
		},
		{
			name:    "DuplicateMention",
			content: "@dup hello @dup",
			want:    []string{"dup"},
		},
		{
			name:    "NoMentions",
			content: "no mentions here",
			want:    nil,
		},
		{
			name:    "EmptyString",
			content: "",
			want:    nil,
		},
		{
			name:    "EmailNotMatched",
			content: "email@example.com should not match",
			want:    nil,
		},
		{
			name:    "MentionAtStartOfLine",
			content: "@startuser is first",
			want:    []string{"startuser"},
		},
		{
			name:    "MentionWithHyphen",
			content: "hello @user-name test",
			want:    []string{"user-name"},
		},
		{
			name:    "MentionWithUnderscore",
			content: "hello @user_name test",
			want:    []string{"user_name"},
		},
		{
			name:    "MentionWithNumbers",
			content: "hello @user123 test",
			want:    []string{"user123"},
		},
		{
			name:    "MultilineMentions",
			content: "@line1\n@line2",
			want:    []string{"line1", "line2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractMentions(tt.content)
			if len(got) != len(tt.want) {
				t.Fatalf("extractMentions(%q) = %v, want %v", tt.content, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("extractMentions(%q)[%d] = %q, want %q", tt.content, i, got[i], tt.want[i])
				}
			}
		})
	}
}
