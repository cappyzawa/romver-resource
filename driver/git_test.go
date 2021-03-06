package driver_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	resource "github.com/cappyzawa/romver-resource"
	. "github.com/cappyzawa/romver-resource/driver"
)

var _ = Describe("Git", func() {
	var (
		initialVersion string

		uri           string
		branch        string
		privateKey    string
		username      string
		password      string
		file          string
		gitUser       string
		depth         string
		commitMessage string

		runner resource.Runner

		gitDriver *GitDriver
	)

	BeforeEach(func() {
		initialVersion = "0"

		uri = "https://github.com/owner/repo"
		branch = "version"
		privateKey = "privateKey"
		username = "username"
		password = "password"
		file = "version.txt"
		depth = ""
		commitMessage = ""

		runner = &mockRunner{
			run: func() error {
				return nil
			},
			combinedOutput: func() ([]byte, error) {
				return nil, nil
			},
			err: func() error {
				return nil
			},
		}

		gitDriver = &GitDriver{
			InitialVersion: initialVersion,

			URI:           uri,
			Branch:        branch,
			PrivateKey:    privateKey,
			Username:      username,
			Password:      password,
			File:          file,
			GitUser:       gitUser,
			Depth:         depth,
			CommitMessage: commitMessage,

			Runner: runner,
		}
	})

	Describe("Bump()", func() {
		Context("nothing file", func() {
			BeforeEach(func() {
				gitDriver.File = "missingFile"
			})
			AfterEach(func() {
				os.RemoveAll("../testdata/missingFile")
			})
			It("return InitialVersion + 1", func() {
				defer SetGitRepoDir("../testdata")()
				bumped, err := gitDriver.Bump()
				Expect(err).NotTo(HaveOccurred())
				initialVersionInt, err := strconv.Atoi(gitDriver.InitialVersion)
				Expect(err).NotTo(HaveOccurred())
				expectInt := initialVersionInt + 1
				expect := strconv.Itoa(expectInt)
				Expect(bumped).To(Equal(expect))
			})
		})
		Context("bump based on existing file", func() {
			var fileVerStr string
			var fileVer int
			BeforeEach(func() {
				fileVerStr = "4"
				err := ioutil.WriteFile("../testdata/version.txt", []byte(fileVerStr), 0755)
				Expect(err).NotTo(HaveOccurred())
				fileVer, err = strconv.Atoi(fileVerStr)
				Expect(err).NotTo(HaveOccurred())
			})
			It("return version in file + 1", func() {
				defer SetGitRepoDir("../testdata")()
				bumped, err := gitDriver.Bump()
				Expect(err).NotTo(HaveOccurred())
				expectInt := fileVer + 1
				expect := strconv.Itoa(expectInt)
				Expect(bumped).To(Equal(expect))
			})
		})
	})
	Describe("Set()", func() {
		It("error has not occurred", func() {
			defer SetGitRepoDir("../testdata")()
			err := gitDriver.Set("5")
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Describe("Check()", func() {
		var cursorVersion string
		Context("when cursor version is empty", func() {
			BeforeEach(func() {
				cursorVersion = ""
			})
			It("checked version is InitialVersion", func() {
				checkedList, err := gitDriver.Check(cursorVersion)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(checkedList)).To(Equal(1))
				Expect(checkedList[0]).To(Equal(gitDriver.InitialVersion))
			})
		})
		Context("when file version is 5 and cursor version is 4", func() {
			BeforeEach(func() {
				cursorVersion = "4"
			})
			It("checked version is 5", func() {
				defer SetGitRepoDir("../testdata")()
				checkedList, err := gitDriver.Check(cursorVersion)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(checkedList)).To(Equal(1))
				Expect(checkedList[0]).To(Equal("5"))
			})
		})
		Context("when file version is 5 and cursor version is 6", func() {
			BeforeEach(func() {
				cursorVersion = "6"
			})
			It("checkedVersion is empty", func() {
				defer SetGitRepoDir("../testdata")()
				checkedList, err := gitDriver.Check(cursorVersion)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(checkedList)).To(Equal(0))
			})
		})
	})
})

type mockRunner struct {
	run            func() error
	combinedOutput func() ([]byte, error)
	err            func() error
}

func (mr *mockRunner) Run(cmd *exec.Cmd) error {
	return mr.run()
}

func (mr *mockRunner) CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	return mr.combinedOutput()
}

func (mr *mockRunner) Error() error {
	return mr.err()
}
