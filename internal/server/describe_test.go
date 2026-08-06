package server

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/tomasz-tomczyk/crit/internal/review"
)

// `crit describe` writes review.json behind the daemon's back, so the daemon
// has to pick the header up on its next disk sync and serve it to the browser.
func TestDescribe_ExternalWriteReachesTheSessionAPI(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "plan.md")
	if err := os.WriteFile(filePath, []byte("# Plan\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	critPath := filepath.Join(dir, ".crit")
	if err := review.SaveCritJSON(critPath, CritJSON{
		Files: map[string]CritJSONFile{"plan.md": {Comments: []Comment{}}},
	}); err != nil {
		t.Fatal(err)
	}

	sess := &Session{
		Mode:      "files",
		OutputDir: dir,
		RepoRoot:  dir,
		Files:     []*FileEntry{{Path: "plan.md", AbsPath: filePath}},
	}
	sess.InitTestChannels()

	s, err := NewServer(sess, frontendFS, "", "test", 0, "")
	if err != nil {
		t.Fatal(err)
	}
	s.projectDir = dir
	s.homeDir = t.TempDir()

	if title, description := sess.GetDescribe(); title != "" || description != "" {
		t.Fatalf("fresh session already has a header: %q / %q", title, description)
	}

	cj, err := review.LoadCritJSON(critPath)
	if err != nil {
		t.Fatal(err)
	}
	cj.Title = "PR #412 — auth race"
	cj.Description = "Fixes the token refresh race.\n\nSee the retry loop."
	if err := review.SaveCritJSON(critPath, cj); err != nil {
		t.Fatal(err)
	}

	sess.SyncCommentsFromDisk()

	title, description := sess.GetDescribe()
	if title != cj.Title {
		t.Errorf("title = %q, want %q", title, cj.Title)
	}
	if description != cj.Description {
		t.Errorf("description = %q, want %q", description, cj.Description)
	}

	req := httptest.NewRequest("GET", "/api/session", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	var resp struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Title != cj.Title || resp.Description != cj.Description {
		t.Errorf("/api/session returned %q / %q", resp.Title, resp.Description)
	}
}

// The daemon rewrites review.json on every comment change; the header is not
// part of its in-memory snapshot, so it has to survive that write.
func TestDescribe_SurvivesADaemonWrite(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "plan.md")
	if err := os.WriteFile(filePath, []byte("# Plan\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	critPath := filepath.Join(dir, ".crit")
	if err := review.SaveCritJSON(critPath, CritJSON{
		Title:       "Kept title",
		Description: "Kept description",
		Files:       map[string]CritJSONFile{"plan.md": {Comments: []Comment{}}},
	}); err != nil {
		t.Fatal(err)
	}

	sess := &Session{
		Mode:      "files",
		OutputDir: dir,
		RepoRoot:  dir,
		Files:     []*FileEntry{{Path: "plan.md", AbsPath: filePath}},
	}
	sess.InitTestChannels()
	sess.AddComment("plan.md", 1, 1, "", "a comment", "", "tester", "")
	if err := sess.SyncWriteFiles(); err != nil {
		t.Fatal(err)
	}

	cj, err := review.LoadCritJSON(critPath)
	if err != nil {
		t.Fatal(err)
	}
	if cj.Title != "Kept title" || cj.Description != "Kept description" {
		t.Errorf("daemon write dropped the header: %q / %q", cj.Title, cj.Description)
	}
}
