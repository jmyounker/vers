package main

import (
	"testing"
)

func TestGitParseBranchTracking(t *testing.T) {
	var cases = []struct {
		Status string
		Want   string
	}{
		{"## master...origin/master\nA  rcs_git_test.go\n", "master"},
		{"## master...origin/master\n", "master"},
		{"## master\n", "master"},
		{"## HEAD (no branch)\n", "HEAD"},
		{"## origin/master\n", "master"},
	}
	for _, tc := range cases {
		b, err := ParseGitStatus(tc.Status)
		failWhenErr(t, err)
		failWhen(t, b != tc.Want)
	}
}
