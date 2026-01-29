package git

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type FileStatus int

const (
	StatusStaged FileStatus = iota
	StatusModified
	StatusUntracked
	StatusConflict
)

type ChangedFile struct {
	Path   string
	Status FileStatus
}

type Repo struct {
	Name         string
	Path         string
	Branch       string
	Staged       int
	Modified     int
	Untracked    int
	Ahead        int
	Behind       int
	HasConflict  bool
	ChangedFiles []ChangedFile
}

func (r Repo) IsDirty() bool {
	return r.Staged > 0 || r.Modified > 0 || r.Untracked > 0
}

// ScanRepos finds all git repos in given paths up to depth.
func ScanRepos(paths []string, depth int) []Repo {
	var repos []Repo

	for _, basePath := range paths {
		filepath.WalkDir(basePath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			relPath, _ := filepath.Rel(basePath, path)
			currentDepth := strings.Count(relPath, string(os.PathSeparator))
			if currentDepth > depth {
				return filepath.SkipDir
			}

			if d.IsDir() && isGitRepo(path) {
				repo := GetRepoStatus(path)
				repos = append(repos, repo)
				return filepath.SkipDir
			}

			return nil
		})
	}

	return repos
}

func isGitRepo(path string) bool {
	gitPath := filepath.Join(path, ".git")
	_, err := os.Stat(gitPath)
	return err == nil
}

// GetRepoStatus gets full status for a repo.
func GetRepoStatus(path string) Repo {
	repo := Repo{
		Name: filepath.Base(path),
		Path: path,
	}

	repo.Branch = getBranch(path)

	if out, err := gitOutput(path, "status", "--porcelain=v1"); err == nil {
		parsePorcelain(&repo, out)
	}

	repo.Ahead, repo.Behind = getAheadBehind(path)

	return repo
}

func getBranch(path string) string {
	out, err := gitOutput(path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "unknown"
	}
	branch := strings.TrimSpace(out)
	if branch == "HEAD" {
		sha, _ := gitOutput(path, "rev-parse", "--short", "HEAD")
		sha = strings.TrimSpace(sha)
		if sha == "" {
			return "detached"
		}
		return "detached@" + sha
	}
	return branch
}

// getAheadBehind uses git CLI for reliable remote comparison.
func getAheadBehind(path string) (ahead, behind int) {
	out, err := gitOutput(path, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	if err != nil {
		return 0, 0
	}

	parts := strings.Fields(out)
	if len(parts) == 2 {
		ahead, _ = strconv.Atoi(parts[0])
		behind, _ = strconv.Atoi(parts[1])
	}

	return
}

// CommitAndPush stages all, commits, and pushes (call after user confirms stage-all).
func CommitAndPush(path, message string) error {
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = path
	if err := addCmd.Run(); err != nil {
		return err
	}

	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = path
	if err := commitCmd.Run(); err != nil {
		return err
	}

	pushCmd := exec.Command("git", "push")
	pushCmd.Dir = path
	return pushCmd.Run()
}

func Pull(path string) error {
	cmd := exec.Command("git", "pull")
	cmd.Dir = path
	return cmd.Run()
}

func FetchAll(path string) error {
	cmd := exec.Command("git", "fetch", "--all")
	cmd.Dir = path
	return cmd.Run()
}

func OpenInEditor(path, editor string) error {
	cmd := exec.Command(editor, path)
	return cmd.Start()
}

func parsePorcelain(repo *Repo, out string) {
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 3 {
			continue
		}
		code := line[:2]
		path := strings.TrimSpace(line[2:])
		cf := ChangedFile{Path: path}

		if code == "??" {
			repo.Untracked++
			cf.Status = StatusUntracked
			repo.ChangedFiles = append(repo.ChangedFiles, cf)
			continue
		}

		if isConflict(code) {
			repo.HasConflict = true
			cf.Status = StatusConflict
			repo.ChangedFiles = append(repo.ChangedFiles, cf)
			continue
		}

		if code[0] != ' ' {
			repo.Staged++
			cf.Status = StatusStaged
			repo.ChangedFiles = append(repo.ChangedFiles, cf)
		}
		if code[1] != ' ' {
			repo.Modified++
			cf.Status = StatusModified
			repo.ChangedFiles = append(repo.ChangedFiles, cf)
		}
	}
}

func isConflict(code string) bool {
	switch code {
	case "UU", "AA", "DD", "AU", "UA", "DU", "UD":
		return true
	default:
		return false
	}
}

func gitOutput(path string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = path
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.String(), err
}
