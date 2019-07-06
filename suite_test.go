package resource_test

import (
	"encoding/json"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var bins struct {
	In    string `json:"in"`
	Out   string `json:"out"`
	Check string `json:"check"`
}

var (
	githubURI      = os.Getenv("ROMVER_TESTING_GITHUB_URI")
	githubBranch   = os.Getenv("ROMVER_TESTING_GITHUB_BRANCH")
	githubUsername = os.Getenv("ROMVER_TESTING_GITHUB_USERNAME")
	githubPassword = os.Getenv("ROMVER_TESTING_GITHUB_PASSWORD")
)

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error

	b := bins
	if _, err := os.Stat("/opt/resource/in"); err != nil {
		b.In, err = gexec.Build("github.com/cappyzawa/romver-resource/cmd/in")
		Expect(err).NotTo(HaveOccurred())
	} else {
		b.In = "/opt/resource/in"
	}

	if _, err := os.Stat("/opt/resource/out"); err != nil {
		b.Out, err = gexec.Build("github.com/cappyzawa/romver-resource/cmd/out")
		Expect(err).NotTo(HaveOccurred())
	} else {
		b.Out = "/opt/resource/out"
	}

	if _, err := os.Stat("/opt/resource/check"); err != nil {
		b.Check, err = gexec.Build("github.com/cappyzawa/romver-resource/cmd/check")
		Expect(err).NotTo(HaveOccurred())
	} else {
		b.Check = "/opt/resource/check"
	}

	j, err := json.Marshal(b)
	Expect(err).ToNot(HaveOccurred())
	return j
}, func(bp []byte) {
	err := json.Unmarshal(bp, &bins)
	Expect(err).ToNot(HaveOccurred())
})

var _ = SynchronizedAfterSuite(func() {

}, func() {
	gexec.CleanupBuildArtifacts()
})

func TestRomverResource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RomverResource Suite")
}
