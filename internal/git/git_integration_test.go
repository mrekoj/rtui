package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetRepoStatusClean(t *testing.T) {
	dir := t.TempDir()
	repo := createRepo(t, dir, "clean")

	status := GetRepoStatus(repo)
	if status.Staged != 0 || status.Modified != 0 || status.Untracked != 0 {
		t.Fatalf("expected clean repo, got S=%d M=%d U=%d", status.Staged, status.Modified, status.Untracked)
	}
	if status.Branch == "" {
		t.Fatal("expected branch name")
	}
}

func TestGetRepoStatusDirtyAndStaged(t *testing.T) {
	dir := t.TempDir()
	repo := createRepo(t, dir, "dirty")

	// modify and stage
	writeFile(t, filepath.Join(repo, "a.txt"), "change")
	runGit(t, repo, "add", "a.txt")

	// modify untracked
	writeFile(t, filepath.Join(repo, "u.txt"), "new")

	status := GetRepoStatus(repo)
	if status.Staged == 0 {
		t.Fatal("expected staged > 0")
	}
	if status.Untracked == 0 {
		t.Fatal("expected untracked > 0")
	}
}

func createRepo(t *testing.T, root, name string) string {
	repo := filepath.Join(root, name)
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	runGit(t, repo, "init")
	writeFile(t, filepath.Join(repo, "a.txt"), "init")
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-m", "init")
	return repo
}

func writeFile(t *testing.T, path, content string) {
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	cmd := gitCmd(dir, args...)
	if err := cmd.Run(); err != nil {
		t.Fatalf("git %v failed: %v", args, err)
	}
}
