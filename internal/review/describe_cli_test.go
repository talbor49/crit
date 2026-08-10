package review

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// describePath mirrors what RunDescribe resolves --output to, so tests read
// back the file the command actually wrote.
func describePath(t *testing.T, outputDir string) string {
	t.Helper()
	critPath, err := ResolveCommandReviewPathWithSession("", outputDir, "")
	if err != nil {
		t.Fatalf("ResolveReviewPath: %v", err)
	}
	return critPath
}

func loadDescribe(t *testing.T, dir string) CritJSON {
	t.Helper()
	cj, err := LoadCritJSON(describePath(t, dir))
	if err != nil {
		t.Fatalf("LoadCritJSON: %v", err)
	}
	return cj
}

func TestRunDescribe(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		title       string
		description string
	}{
		{
			name:        "title and body",
			args:        []string{"--title", "PR #412 — auth race", "--body", "  Summary of the change.  "},
			title:       "PR #412 — auth race",
			description: "Summary of the change.",
		},
		{
			name:  "title only",
			args:  []string{"--title", "Just a title"},
			title: "Just a title",
		},
		{
			name:        "body only",
			args:        []string{"--body", "## Why\n\nBecause."},
			description: "## Why\n\nBecause.",
		},
		{name: "no arguments is a usage error", args: nil, wantErr: true},
		{name: "unknown flag is a usage error", args: []string{"--nope"}, wantErr: true},
		{name: "missing flag value is a usage error", args: []string{"--title"}, wantErr: true},
		{
			name:    "clear cannot be combined",
			args:    []string{"--clear", "--title", "x"},
			wantErr: true,
		},
		{
			name:    "oversized description is rejected",
			args:    []string{"--body", strings.Repeat("x", MaxDescribeBodyBytes+1)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			err := RunDescribe(append(append([]string{}, tt.args...), "--output", dir))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("RunDescribe: %v", err)
			}
			cj := loadDescribe(t, dir)
			if cj.Title != tt.title {
				t.Errorf("title = %q, want %q", cj.Title, tt.title)
			}
			if cj.Description != tt.description {
				t.Errorf("description = %q, want %q", cj.Description, tt.description)
			}
		})
	}
}

func TestRunDescribe_PartialUpdateKeepsTheOtherField(t *testing.T) {
	dir := t.TempDir()
	if err := RunDescribe([]string{"--title", "Original", "--body", "Original body", "--output", dir}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := RunDescribe([]string{"--title", "Renamed", "--output", dir}); err != nil {
		t.Fatalf("retitle: %v", err)
	}
	cj := loadDescribe(t, dir)
	if cj.Title != "Renamed" {
		t.Errorf("title = %q, want %q", cj.Title, "Renamed")
	}
	if cj.Description != "Original body" {
		t.Errorf("description = %q, want it preserved", cj.Description)
	}
}

func TestRunDescribe_ClearRemovesBoth(t *testing.T) {
	dir := t.TempDir()
	if err := RunDescribe([]string{"--title", "T", "--body", "B", "--output", dir}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := RunDescribe([]string{"--clear", "--output", dir}); err != nil {
		t.Fatalf("clear: %v", err)
	}
	cj := loadDescribe(t, dir)
	if cj.Title != "" || cj.Description != "" {
		t.Errorf("clear left title=%q description=%q", cj.Title, cj.Description)
	}
}

func TestRunDescribe_ReadsBodyFromFile(t *testing.T) {
	dir := t.TempDir()
	body := filepath.Join(dir, "summary.md")
	if err := os.WriteFile(body, []byte("# Summary\n\nRefactors the retry loop.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RunDescribe([]string{"--title", "Retry loop", "--file", body, "--output", dir}); err != nil {
		t.Fatalf("RunDescribe: %v", err)
	}
	cj := loadDescribe(t, dir)
	if cj.Description != "# Summary\n\nRefactors the retry loop." {
		t.Errorf("description = %q", cj.Description)
	}
}

func TestRunDescribe_MissingFileIsAnError(t *testing.T) {
	dir := t.TempDir()
	err := RunDescribe([]string{"--file", filepath.Join(dir, "nope.md"), "--output", dir})
	if err == nil {
		t.Fatal("expected an error for a missing --file")
	}
}

// The review file is a shared artefact; describing a review must not disturb
// comments already in it.
func TestRunDescribe_PreservesComments(t *testing.T) {
	dir := t.TempDir()
	cj := CritJSON{
		Branch:      "feature",
		ReviewRound: 2,
		Files: map[string]CritJSONFile{
			"main.go": {Comments: []Comment{{ID: "c1", Body: "keep me"}}},
		},
	}
	if err := SaveCritJSON(describePath(t, dir), cj); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := RunDescribe([]string{"--title", "Hello", "--output", dir}); err != nil {
		t.Fatalf("RunDescribe: %v", err)
	}
	got := loadDescribe(t, dir)
	if got.Title != "Hello" {
		t.Errorf("title = %q", got.Title)
	}
	if len(got.Files["main.go"].Comments) != 1 || got.Files["main.go"].Comments[0].Body != "keep me" {
		t.Errorf("comments were not preserved: %+v", got.Files)
	}
	if got.ReviewRound != 2 {
		t.Errorf("review_round = %d, want 2", got.ReviewRound)
	}
}

func TestParseDescribeArgs_NormalizesStdin(t *testing.T) {
	for _, args := range [][]string{
		{"--file", "-"},
		{"-"},
	} {
		got, err := parseDescribeArgs(args)
		if err != nil {
			t.Fatalf("parseDescribeArgs(%v): %v", args, err)
		}
		if !got.stdin {
			t.Errorf("parseDescribeArgs(%v) did not select stdin", args)
		}
		if got.file != "" {
			t.Errorf("parseDescribeArgs(%v) left file = %q", args, got.file)
		}
	}
}

func TestRunDescribe_SessionAndOutputConflict(t *testing.T) {
	err := RunDescribe([]string{"--title", "x", "--session", "abcd1234ef01", "--output", t.TempDir()})
	if err == nil {
		t.Fatal("expected --session + --output to be rejected")
	}
}

func TestRunDescribe_UnknownSessionIsAnError(t *testing.T) {
	err := RunDescribe([]string{"--title", "x", "--session", "000000000000"})
	if err == nil {
		t.Fatal("expected an error for a session that is not running")
	}
}
