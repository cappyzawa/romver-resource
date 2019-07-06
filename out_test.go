package resource_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"

	resource "github.com/cappyzawa/romver-resource"
	"github.com/cappyzawa/romver-resource/driver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Out", func() {
	var srcDir string
	var req struct {
		Source resource.Source
		Params resource.OutParams
	}

	var res struct {
		Version  resource.Version
		Metadata resource.Metadata
	}

	BeforeEach(func() {
		var err error
		srcDir, err = ioutil.TempDir("", "romver-resource-out-dir")
		Expect(err).NotTo(HaveOccurred())

		req.Source = resource.Source{}
		req.Params = resource.OutParams{}

		res.Version = resource.Version{}
		res.Metadata = resource.Metadata{}
	})

	AfterEach(func() {
		Expect(os.RemoveAll(srcDir)).To(Succeed())
	})

	JustBeforeEach(func() {
		cmd := exec.Command(bins.Out, srcDir)

		payload, err := json.Marshal(req)
		Expect(err).ToNot(HaveOccurred())

		outBuf := new(bytes.Buffer)

		cmd.Stdin = bytes.NewBuffer(payload)
		cmd.Stdout = outBuf
		cmd.Stderr = GinkgoWriter

		err = cmd.Run()
		Expect(err).ToNot(HaveOccurred())

		err = json.Unmarshal(outBuf.Bytes(), &res)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("bump based on git file", func() {
		BeforeEach(func() {
			req.Source = resource.Source{
				Driver: "git",

				InitialVersion: "0",

				URI:      githubURI,
				Branch:   githubBranch,
				Username: githubUsername,
				Password: githubPassword,
				File:     "version",
			}
			req.Params = resource.OutParams{
				Bump: true,
			}
		})
		It("works", func() {
			_, err := driver.FromSource(req.Source)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
