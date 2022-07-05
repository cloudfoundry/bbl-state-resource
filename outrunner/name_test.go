package outrunner_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bbl-state-resource/concourse"
	"github.com/cloudfoundry/bbl-state-resource/outrunner"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Name", func() {
	Context("when passed empty params", func() {
		It("generates a random name", func() {
			name, err1 := outrunner.Name("", concourse.OutParams{})
			name2, err2 := outrunner.Name("", concourse.OutParams{})

			Expect(err1).NotTo(HaveOccurred())
			Expect(err2).NotTo(HaveOccurred())
			Expect(name).To(MatchRegexp(`\w+-\w+`))
			Expect(name).NotTo(Equal(name2))
		})
	})

	Context("when passed a name", func() {
		It("returns the name", func() {
			name, err := outrunner.Name("", concourse.OutParams{Name: "some-env-name", NameFile: "ignore-the-name-file", StateDir: "ignore-the-state-dir"})

			Expect(err).NotTo(HaveOccurred())
			Expect(name).To(Equal("some-env-name"))
		})
	})

	Context("when passed a name file", func() {
		Context("success", func() {
			var tempDir string
			BeforeEach(func() {
				var err error
				tempDir, err = ioutil.TempDir("", "")
				Expect(err).NotTo(HaveOccurred())

				nameFilePath := filepath.Join(tempDir, "name-file")

				err = ioutil.WriteFile(nameFilePath, []byte("some-env-name"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the name in the name file", func() {
				name, err := outrunner.Name(tempDir, concourse.OutParams{NameFile: "name-file", StateDir: "ignore-the-state-dir"})
				Expect(err).NotTo(HaveOccurred())

				Expect(name).To(Equal("some-env-name"))
			})
		})

		Context("failure", func() {
			It("returns an error", func() {
				_, err := outrunner.Name("", concourse.OutParams{NameFile: "not-a-real-file"})
				Expect(err).To(MatchError("Failure reading name file: open not-a-real-file: no such file or directory"))
			})
		})
	})

	Context("when passed only a state dir", func() {
		Context("success", func() {
			var nameFilePath string
			var tempDir string
			BeforeEach(func() {
				var err error
				tempDir, err = ioutil.TempDir("", "")
				Expect(err).NotTo(HaveOccurred())

				nameFilePath = filepath.Join(tempDir, "name")

				err = ioutil.WriteFile(nameFilePath, []byte("some-env-name"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the name in the name file", func() {
				name, err := outrunner.Name(tempDir, concourse.OutParams{StateDir: "."})
				Expect(err).NotTo(HaveOccurred())

				Expect(name).To(Equal("some-env-name"))
			})
		})

		Context("failure", func() {
			It("returns an error", func() {
				_, err := outrunner.Name("", concourse.OutParams{StateDir: "not-a-real-dir"})
				Expect(err).To(MatchError("Failure reading name file: open not-a-real-dir/name: no such file or directory"))
			})
		})
	})
})
