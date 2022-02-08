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

	resource "github.com/cappyzawa/romver-resource"
)

var (
	gitRepoDir     string
	privateKeyPATH string
	netRcPATH      string
)

var ErrEncryptedKey = errors.New("private keys with passphrases are not supported")

func init() {
	gitRepoDir = filepath.Join(os.TempDir(), "romver-git-repo")
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

	Runner resource.Runner
}

// Bump increments version and pushs
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

// Set pushs version, but does not increment
func (gd *GitDriver) Set(version string) error {
	if err := gd.setUpAuth(); err != nil {
		return err
	}

	if err := gd.setUserInfo(); err != nil {
		return err
	}

	for {
		if err := gd.setUpRepo(); err != nil {
			return err
		}

		wrote, err := gd.writeVersion(version)
		if err != nil {
			return err
		}

		if wrote {
			break
		}
	}

	return nil
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
			if err := ioutil.WriteFile(privateKeyPATH, []byte(gd.PrivateKey+"\n"), 0600); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if gd.isPrivateKeyEncrypted(privateKeyPATH) {
		return ErrEncryptedKey
	}
	return os.Setenv("GIT_SSH_COMMAND", "ssh -o StrictHostKeyChecking=no -i "+privateKeyPATH)
}

func (gd *GitDriver) isPrivateKeyEncrypted(path string) bool {
	passphrases := ``
	cmd := exec.Command(`ssh-keygen`, `-y`, `-f`, path, `-P`, passphrases)
	err := gd.Runner.Run(cmd)
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
		if err := gd.Runner.Run(gitName); err != nil {
			return fmt.Errorf("error: %v, detail: %v", err, gd.Runner.Error())
		}
	}

	gitEmail := exec.Command("git", "config", "--global", "user.email", user.Address)
	if err := gd.Runner.Run(gitEmail); err != nil {
		return fmt.Errorf("error: %v, detail: %v", err, gd.Runner.Error())
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
		if err := gd.Runner.Run(gitClone); err != nil {
			return fmt.Errorf("error: %v, detail: %v", err, gd.Runner.Error())
		}
	} else {
		gitFetch := exec.Command("git", "fetch", "origin", gd.Branch)
		gitFetch.Dir = gitRepoDir
		if err := gd.Runner.Run(gitFetch); err != nil {
			return fmt.Errorf("error: %v, detail: %v", err, gd.Runner.Error())
		}
	}

	gitCheckout := exec.Command("git", "reset", "--hard", "origin/"+gd.Branch)
	gitCheckout.Dir = gitRepoDir
	if err := gd.Runner.Run(gitCheckout); err != nil {
		return fmt.Errorf("error: %v, detail: %v", err, gd.Runner.Error())
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
	if err := gd.Runner.Run(gitAdd); err != nil {
		return false, fmt.Errorf("error: %v, detail: %v", err, gd.Runner.Error())
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
	commitOutput, err := gd.Runner.CombinedOutput(gitCommit)
	if err != nil {
		return false, err
	}
	if strings.Contains(string(commitOutput), nothingToCommitString) {
		return true, nil
	}

	gitPush := exec.Command("git", "push", "origin", "HEAD:"+gd.Branch)
	gitPush.Dir = gitRepoDir

	pushOutput, err := gd.Runner.CombinedOutput(gitPush)
	if err != nil {
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
