package build

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// cloneGitHubRepo clones a public GitHub repository into the specified
// destination directory using a shallow, single-branch clone.
// Shallow clone (--depth 1) downloads only the latest commit, which is
// all the deployment pipeline needs. no history, tags, or other branches.
//
// The logWriter receives git's stderr output (progress messages, errors)
// so that clone activity is captured in the deployment's log file alongside
// build output. Git writes progress to stderr, not stdout.
//
// This function shells out to the system `git` binary via exec.Command()
// rather than using a pure-Go git library (go-git). The native binary is
// faster, handles all protocol edge cases, and avoids pulling in ~30+
// transitive dependencies for a single fire-and-forget clone operation.
// The Go backend's Docker image must include git (one `apk add git` line).
func cloneGitHubRepo(repoURL string, branch string, destinationDir string, logWriter io.Writer) error {
	// exec.Command constructs the command but does not run it yet.
	// the command is equivalent to:
	//   git clone --depth 1 --single-branch --branch <branch> <repoURL> <destinationDir>
	//
	// --depth 1:        shallow clone, only the latest commit (no history)
	// --single-branch:  fetch only the specified branch, not all remote branches
	// --branch:         which branch to clone (user-configured, defaults to "main")
	//
	// the destination directory must NOT already exist, git clone creates it.
	// the caller is responsible for ensuring the path is available.
	gitCloneCommand := exec.Command(
		"git", "clone",
		"--depth", "1",
		"--single-branch",
		"--branch", branch,
		repoURL,
		destinationDir,
	)

	// git writes clone progress (remote counting, receiving objects, resolving deltas)
	// to stderr, not stdout. stdout is used for plumbing commands (git log, git diff)
	// that produce machine-readable output.
	// routing stderr to the logWriter captures clone progress in the deployment log file.
	// also capture stderr into a buffer so we can include git's error message
	// (e.g. "Remote branch main not found") in the returned error for better diagnostics.
	var stderrBuf bytes.Buffer
	gitCloneCommand.Stderr = io.MultiWriter(logWriter, &stderrBuf)

	// stdout from `git clone` is typically empty for a normal clone operation.
	// routing it to the same logWriter ensures nothing is silently discarded
	// in case git emits unexpected output.
	gitCloneCommand.Stdout = logWriter

	// Run() starts the command and waits for it to finish.
	// it returns a non-nil error if
	//   - the git binary is not found on the system PATH (exec.ErrNotFound)
	//   - the clone fails (invalid URL, branch not found, network error, auth required for private repo)
	//   - the process exits with a non-zero exit code
	// the error message from git (written to stderr and captured in the log file)
	// provides the specific failure reason for debugging.
	errGitClone := gitCloneCommand.Run()
	if errGitClone != nil {
		// extract the last meaningful line from git's stderr for a user-facing hint.
		// git stderr typically ends with something like:
		//   "fatal: Remote branch main not found in upstream origin"
		//   "fatal: repository 'https://...' not found"
		stderrOutput := strings.TrimSpace(stderrBuf.String())
		hint := extractGitErrorHint(stderrOutput)

		if hint != "" {
			return fmt.Errorf("git clone failed for %q (branch %q): %s", repoURL, branch, hint)
		}
		return fmt.Errorf("git clone failed for %q (branch %q): %w", repoURL, branch, errGitClone)
	}

	return nil
}

// extractGitErrorHint parses git's stderr output and returns a concise,
// user-facing error message. Git prefixes fatal errors with "fatal: "
// which contains the actual reason for the failure.
func extractGitErrorHint(stderr string) string {
	// look for the last "fatal:" line, which is the most specific error
	lines := strings.Split(stderr, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "fatal:") {
			// return the message after "fatal: "
			return strings.TrimSpace(strings.TrimPrefix(line, "fatal:"))
		}
	}
	return ""
}
