package resource_test

import (
	"bytes"
	"encoding/json"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	resource "github.com/cappyzawa/romver-resource"
)

var _ = Describe("Check", func() {
	var req struct {
		Source  resource.Source
		Version *resource.Version
	}

	var res []resource.Version

	BeforeEach(func() {
		req.Source = resource.Source{}
		req.Version = &resource.Version{}

		res = []resource.Version{}
	})

	JustBeforeEach(func() {
		cmd := exec.Command(bins.Check)

		payload, err := json.Marshal(req)
		Expect(err).ToNot(HaveOccurred())

		checkBuf := new(bytes.Buffer)

		cmd.Stdin = bytes.NewBuffer(payload)
		cmd.Stdout = checkBuf
		cmd.Stderr = GinkgoWriter

		err = cmd.Run()
		Expect(err).ToNot(HaveOccurred())

		err = json.Unmarshal(checkBuf.Bytes(), &res)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("when target file exists in github", func() {
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
		})

		It("works", func() {
			Expect(len(res)).NotTo(BeZero())
		})
	})
	Context("when target file does not exists in github", func() {
		BeforeEach(func() {
			req.Source = resource.Source{
				Driver: "git",

				InitialVersion: "0",

				URI:      githubURI,
				Branch:   githubBranch,
				Username: githubUsername,
				Password: githubPassword,
				File:     "missingFile",
			}
		})

		It("works", func() {
			Expect(len(res)).NotTo(BeZero())
		})
		It("response is initialVersion", func() {
			Expect(res[0].Number).To(Equal(req.Source.InitialVersion))
		})
	})
})
