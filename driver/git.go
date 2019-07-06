package driver

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/mail"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	gitRepoDir     string
	privateKeyPATH string
	netRcPATH      string
)

var ErrEncryptedKey = errors.New("private keys with passphrases are not supported")

func init() {
	gitRepoDir = filepath.Join(os.TempDir(), "semver-git-repo")
	privateKeyPATH = filepath.Join(os.TempDir(), "private-key")
	netRcPATH = filepath.Join(os.Getenv("HOME"), ".netrc")
}

// GitDriver accesses git
type GitDriver struct {
	InitialVersion string

	URI           string
	Branch        string
	PrivateKey    string
	Username      string
	Password      string
	File          string
	GitUser       string
	Depth         string
	CommitMessage string
}

// Bump increments version
func (gd *GitDriver) Bump() (string, error) {
	if err := gd.setUpAuth(); err != nil {
		return "", err
	}
	if err := gd.setUserInfo(); err != nil {
		return "", err
	}

	var newVersion string
	for {
		if err := gd.setUpRepo(); err != nil {
			return "", err
		}

		currentVersion, exists, err := gd.readVersion()
		if err != nil {
			return "", err
		}
		if !exists {
			currentVersion = gd.InitialVersion
		}

		currentVersionInt, err := strconv.Atoi(currentVersion)
		if err != nil {
			return "", nil
		}

		newVersion = strconv.Itoa(currentVersionInt + 1)
		wrote, err := gd.writeVersion(newVersion)
		if wrote {
			break
		}
	}
	return newVersion, nil
}

func (gd *GitDriver) setUpAuth() error {
	if _, err := os.Stat(netRcPATH); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		if err := os.Remove(netRcPATH); err != nil {
			return err
		}
	}

	if gd.PrivateKey != "" {
		if err := gd.setUpKey(); err != nil {
			return err
		}
	}

	if gd.Username != "" && gd.Password != "" {
		if err := gd.setUpUsernamePassword(); err != nil {
			return err
		}
	}
	return nil
}

func (gd *GitDriver) setUpKey() error {
	if _, err := os.Stat(privateKeyPATH); err != nil {
		if os.IsNotExist(err) {
			if err := ioutil.WriteFile(privateKeyPATH, []byte(gd.PrivateKey), 0600); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if isPrivateKeyEncrypted(privateKeyPATH) {
		return ErrEncryptedKey
	}
	return os.Setenv("GIT_SSH_COMMAND", "ssh -o StrictHostKeyChecking=no -i "+privateKeyPATH)
}

func isPrivateKeyEncrypted(path string) bool {
	passphrases := ``
	cmd := exec.Command(`ssh-keygen`, `-y`, `-f`, path, `-P`, passphrases)
	err := cmd.Run()
	return err != nil
}

func (gd *GitDriver) setUpUsernamePassword() error {
	if _, err := os.Stat(netRcPATH); err != nil {
		if os.IsNotExist(err) {
			content := fmt.Sprintf("default login %s password %s", gd.Username, gd.Password)
			if err := ioutil.WriteFile(netRcPATH, []byte(content), 0600); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func (gd *GitDriver) setUserInfo() error {
	if gd.GitUser == "" {
		return nil
	}

	user, err := mail.ParseAddress(gd.GitUser)
	if err != nil {
		return err
	}

	if user.Name != "" {
		gitName := exec.Command("git", "config", "--global", "user.name", user.Name)
		gitName.Stdout = os.Stderr
		gitName.Stderr = os.Stderr
		if err := gitName.Run(); err != nil {
			return err
		}
	}

	gitEmail := exec.Command("git", "config", "--global", "user.email", user.Address)
	gitEmail.Stdout = os.Stderr
	gitEmail.Stderr = os.Stderr
	if err := gitEmail.Run(); err != nil {
		return err
	}
	return nil
}

func (gd *GitDriver) setUpRepo() error {
	if _, err := os.Stat(gitRepoDir); err != nil {
		gitClone := exec.Command("git", "clone", gd.URI, "--branch", gd.Branch)
		if gd.Depth != "" {
			gitClone.Args = append(gitClone.Args, "--depth", gd.Depth)
		}
		gitClone.Args = append(gitClone.Args, "--single-branch", gitRepoDir)
		gitClone.Stdout = os.Stderr
		gitClone.Stderr = os.Stderr
		if err := gitClone.Run(); err != nil {
			return err
		}
	} else {
		gitFetch := exec.Command("git", "fetch", "origin", gd.Branch)
		gitFetch.Dir = gitRepoDir
		gitFetch.Stdout = os.Stderr
		gitFetch.Stderr = os.Stderr
		if err := gitFetch.Run(); err != nil {
			return err
		}
	}

	gitCheckout := exec.Command("git", "reset", "--hard", "origin/"+gd.Branch)
	gitCheckout.Dir = gitRepoDir
	gitCheckout.Stdout = os.Stderr
	gitCheckout.Stderr = os.Stderr
	if err := gitCheckout.Run(); err != nil {
		return err
	}

	return nil
}

func (gd *GitDriver) readVersion() (string, bool, error) {
	var currentVersion string
	versionFile, err := os.Open(filepath.Join(gitRepoDir, gd.File))
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}
	defer versionFile.Close()

	if _, err := fmt.Fscanf(versionFile, "%s", &currentVersion); err != nil {
		return "", false, err
	}
	return currentVersion, true, nil
}

const (
	nothingToCommitString    = "nothing to commit"
	falsePushString          = "Everything up-to-date"
	pushRejectedString       = "[rejected]"
	pushRemoteRejectedString = "[remote rejected]"
)

func (gd *GitDriver) writeVersion(newVersion string) (bool, error) {
	if err := ioutil.WriteFile(filepath.Join(gitRepoDir, gd.File), []byte(newVersion), 0644); err != nil {
		return false, nil
	}

	gitAdd := exec.Command("git", "add", gd.File)
	gitAdd.Dir = gitRepoDir
	gitAdd.Stdout = os.Stderr
	gitAdd.Stderr = os.Stderr
	if err := gitAdd.Run(); err != nil {
		return false, err
	}
	var commitMessage string
	if gd.CommitMessage == "" {
		commitMessage = fmt.Sprintf("bump to %s", newVersion)
	} else {
		commitMessage = strings.Replace(gd.CommitMessage, "%version%", newVersion, -1)
		commitMessage = strings.Replace(gd.CommitMessage, "%file%", gd.File, -1)
	}

	gitCommit := exec.Command("git", "commit", "-m", commitMessage)
	gitCommit.Dir = gitRepoDir
	commitOutput, err := gitCommit.CombinedOutput()
	if err != nil {
		return false, err
	}
	if strings.Contains(string(commitOutput), nothingToCommitString) {
		return true, nil
	}

	gitPush := exec.Command("git", "push", "origin", "HEAD:"+gd.Branch)
	gitPush.Dir = gitRepoDir

	pushOutput, err := gitPush.CombinedOutput()
	if err != nil {
		os.Stderr.Write(pushOutput)
		return false, err
	}

	if strings.Contains(string(pushOutput), falsePushString) {
		return false, nil
	}

	if strings.Contains(string(pushOutput), pushRejectedString) {
		return false, nil
	}

	if strings.Contains(string(pushOutput), pushRemoteRejectedString) {
		return false, nil
	}

	return true, err
}

// Check checks new version
func (gd *GitDriver) Check(cursor string) ([]string, error) {
	if err := gd.setUpAuth(); err != nil {
		return nil, err
	}

	if err := gd.setUpRepo(); err != nil {
		return nil, err
	}

	currentVersion, exists, err := gd.readVersion()
	if err != nil {
		return nil, err
	}
	if !exists {
		return []string{gd.InitialVersion}, nil
	}

	if cursor == "" {
		cursor = gd.InitialVersion
	}

	isCurrentGreater, err := gte(currentVersion, cursor)
	if err != nil {
		return nil, err
	}

	if isCurrentGreater {
		return []string{currentVersion}, nil
	}

	return []string{}, nil
}

func gte(current, cursor string) (bool, error) {
	currentInt, err := strconv.Atoi(current)
	if err != nil {
		return false, err
	}
	cursorInt, err := strconv.Atoi(cursor)
	if err != nil {
		return false, err
	}

	return currentInt-cursorInt >= 0, nil
}
