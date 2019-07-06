package resource_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	resource "github.com/cappyzawa/romver-resource"
)

var _ = Describe("In", func() {
	var desDir string
	var req struct {
		Source  resource.Source
		Version resource.Version
		Params  resource.InParams
	}

	var res struct {
		Version  resource.Version
		Metadata resource.Metadata
	}

	BeforeEach(func() {
		var err error
		desDir, err = ioutil.TempDir("", "romver-resource-in-dir")
		Expect(err).NotTo(HaveOccurred())

		req.Source = resource.Source{}
		req.Params = resource.InParams{}

		res.Version = resource.Version{}
		res.Metadata = resource.Metadata{}
	})
	AfterEach(func() {
		Expect(os.RemoveAll(desDir)).To(Succeed())
	})

	JustBeforeEach(func() {
		cmd := exec.Command(bins.In, desDir)

		payload, err := json.Marshal(req)
		Expect(err).ToNot(HaveOccurred())

		inBuf := new(bytes.Buffer)

		cmd.Stdin = bytes.NewBuffer(payload)
		cmd.Stdout = inBuf
		cmd.Stderr = GinkgoWriter

		err = cmd.Run()
		Expect(err).ToNot(HaveOccurred())

		err = json.Unmarshal(inBuf.Bytes(), &res)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("no bump", func() {
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

			req.Params = resource.InParams{
				Bump: false,
			}

			req.Version = resource.Version{
				Number: "111",
			}
		})

		It("works", func() {
			Expect(res.Version.Number).To(Equal(req.Version.Number))
		})

		It("the content of the version file is the request version number", func() {
			b, err := ioutil.ReadFile(filepath.Join(desDir, "version"))
			Expect(err).NotTo(HaveOccurred())
			number := strings.TrimSpace(string(b))
			Expect(number).To(Equal(req.Version.Number))
		})

	})

	Context("bump", func() {
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

			req.Params = resource.InParams{
				Bump: true,
			}

			req.Version = resource.Version{
				Number: "111",
			}
		})

		It("works", func() {
			Expect(res.Version.Number).To(Equal(req.Version.Number))
		})

		It("the content of the version file is the request version number + 1", func() {
			b, err := ioutil.ReadFile(filepath.Join(desDir, "version"))
			Expect(err).NotTo(HaveOccurred())
			number := strings.TrimSpace(string(b))
			reqVerInt, err := strconv.Atoi(req.Version.Number)
			Expect(err).NotTo(HaveOccurred())
			expectInt := reqVerInt + 1
			expect := strconv.Itoa(expectInt)
			Expect(number).To(Equal(expect))
		})
	})
})
