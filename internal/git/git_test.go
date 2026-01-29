package git

import "testing"

func TestParsePorcelain(t *testing.T) {
	repo := Repo{}
	out := " M modified.txt\n" +
		"M  staged.txt\n" +
		"A  added.txt\n" +
		"?? untracked.txt\n" +
		"UU conflict.txt\n"

	parsePorcelain(&repo, out)

	if repo.Modified != 1 {
		t.Fatalf("Modified = %d, want 1", repo.Modified)
	}
	if repo.Staged != 2 {
		t.Fatalf("Staged = %d, want 2", repo.Staged)
	}
	if repo.Untracked != 1 {
		t.Fatalf("Untracked = %d, want 1", repo.Untracked)
	}
	if !repo.HasConflict {
		t.Fatal("expected HasConflict true")
	}
	if len(repo.ChangedFiles) != 5 {
		t.Fatalf("ChangedFiles len = %d, want 5", len(repo.ChangedFiles))
	}
}

func TestParsePorcelainBothStages(t *testing.T) {
	repo := Repo{}
	out := "MM both.txt\n"
	parsePorcelain(&repo, out)

	if repo.Staged != 1 || repo.Modified != 1 {
		t.Fatalf("Staged=%d Modified=%d, want 1/1", repo.Staged, repo.Modified)
	}
}
