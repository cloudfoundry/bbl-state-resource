package storage_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bbl-state-resource/fakes"
	"github.com/cloudfoundry/bbl-state-resource/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Storage", func() {
	var (
		storageDir      string
		filename        string
		store           storage.Storage
		fakeTarrer      *fakes.Tarrer
		fakeObject      *fakes.Object
		fakeReadCloser  *fakes.ReadCloser
		fakeWriteCloser *fakes.WriteCloser
	)

	BeforeEach(func() {
		fakeTarrer = &fakes.Tarrer{}

		fakeReadCloser = &fakes.ReadCloser{}
		fakeWriteCloser = &fakes.WriteCloser{}

		fakeObject = &fakes.Object{}
		fakeObject.NewReaderCall.Returns.ReadCloser = fakeReadCloser
		fakeObject.NewWriterCall.Returns.WriteCloser = fakeWriteCloser
		fakeObject.VersionCall.Returns.Version = "fresh-version"

		By("creating a temporary directory to walk", func() {
			var err error
			storageDir, err = ioutil.TempDir("", "storage_dir")
			Expect(err).NotTo(HaveOccurred())
			filename = filepath.Join(storageDir, "bbl-state.json")
			f, err := os.Create(filename)
			Expect(err).NotTo(HaveOccurred())
			defer f.Close()

			bblStateContents := fmt.Sprintf(`{"version": 14, "randomDir": "%s"}`, storageDir)
			_, err = f.Write([]byte(bblStateContents))
			Expect(err).NotTo(HaveOccurred())
		})

		store = storage.Storage{
			DirectoryName: "dir-inside-tarball",
			Object:        fakeObject,
			Archiver:      fakeTarrer,
		}
	})

	AfterEach(func() {
		// _ = os.RemoveAll(storageDir) // ignore the error
	})

	Describe("Upload", func() {
		It("tars the contents of filepath and uploads them", func() {
			err := store.Upload(storageDir)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeTarrer.WriteCall.Receives.Output).To(Equal(fakeWriteCloser))
			Expect(fakeTarrer.WriteCall.Receives.Sources).To(HaveLen(1))
			Expect(filepath.Base(fakeTarrer.WriteCall.Receives.Sources[0])).To(Equal("dir-inside-tarball"))

			Expect(fakeWriteCloser.CloseCall.CallCount).To(Equal(1))
		})

		Context("when archiving the file returns an error", func() {
			BeforeEach(func() {
				fakeTarrer.WriteCall.Returns.Error = errors.New("coconut")
			})

			It("returns an error", func() {
				err := store.Upload(storageDir)
				Expect(err).To(MatchError("coconut"))

				Expect(fakeWriteCloser.CloseCall.CallCount).To(Equal(0))
			})
		})

		Context("when closing the writer returns an error", func() {
			BeforeEach(func() {
				fakeWriteCloser.CloseCall.Returns.Error = errors.New("mango")
			})

			It("returns an error", func() {
				err := store.Upload(storageDir)
				Expect(err).To(MatchError("mango"))
			})
		})
	})

	Describe("Download", func() {
		Context("when the object already exists", func() {
			FIt("downloads the object, tars it, and re-uploads it", func() {
				err := store.Download(storageDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeTarrer.ReadCall.Receives.Input).To(Equal(fakeReadCloser))
				Expect(fakeTarrer.ReadCall.Receives.Destination).To(Equal(storageDir))

				Expect(fakeTarrer.WriteCall.Receives.Output).To(Equal(fakeWriteCloser))
				Expect(fakeTarrer.WriteCall.Receives.Sources).To(HaveLen(1))
				Expect(filepath.Base(fakeTarrer.WriteCall.Receives.Sources[0])).To(Equal("bbl-state.json"))
				Expect(fakeTarrer.WriteCall.Receives.Sources[0]).NotTo(ContainSubstring("dir-inside-tarball"))

				Expect(fakeReadCloser.CloseCall.CallCount).To(Equal(1))
				Expect(fakeWriteCloser.CloseCall.CallCount).To(Equal(1))
			})
		})

		Context("when the object does not exist", func() {
			BeforeEach(func() {
				fakeObject.NewReaderCall.Returns.Error = storage.ObjectNotFoundError
			})

			It("uploads the object", func() {
				err := store.Download(storageDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeTarrer.ReadCall.CallCount).To(Equal(0))

				Expect(fakeTarrer.WriteCall.Receives.Output).To(Equal(fakeWriteCloser))
				Expect(fakeTarrer.WriteCall.Receives.Sources).To(HaveLen(1))
				Expect(filepath.Base(fakeTarrer.WriteCall.Receives.Sources[0])).To(Equal("bbl-state.json"))
				Expect(fakeTarrer.WriteCall.Receives.Sources[0]).NotTo(ContainSubstring("dir-inside-tarball"))

				Expect(fakeReadCloser.CloseCall.CallCount).To(Equal(0))
				Expect(fakeWriteCloser.CloseCall.CallCount).To(Equal(1))
			})
		})

		Context("when reading the object returns an error", func() {
			BeforeEach(func() {
				fakeObject.NewReaderCall.Returns.Error = errors.New("papaya")
			})

			It("returns the error", func() {
				err := store.Download(storageDir)
				Expect(err).To(MatchError("papaya"))

				Expect(fakeTarrer.ReadCall.CallCount).To(Equal(0))
				Expect(fakeTarrer.WriteCall.CallCount).To(Equal(0))

				Expect(fakeReadCloser.CloseCall.CallCount).To(Equal(0))
				Expect(fakeWriteCloser.CloseCall.CallCount).To(Equal(0))
			})
		})

		Context("when reading the object returns an error", func() {
			BeforeEach(func() {
				fakeObject.NewReaderCall.Returns.Error = errors.New("papaya")
			})

			It("returns the error", func() {
				err := store.Download(storageDir)
				Expect(err).To(MatchError("papaya"))

				Expect(fakeTarrer.ReadCall.CallCount).To(Equal(0))
				Expect(fakeTarrer.WriteCall.CallCount).To(Equal(0))

				Expect(fakeReadCloser.CloseCall.CallCount).To(Equal(0))
				Expect(fakeWriteCloser.CloseCall.CallCount).To(Equal(0))
			})
		})

		Context("when reading the archive returns an error", func() {
			BeforeEach(func() {
				fakeTarrer.ReadCall.Returns.Error = errors.New("mango")
			})

			It("returns the error", func() {
				err := store.Download(storageDir)
				Expect(err).To(MatchError("mango"))

				Expect(fakeTarrer.WriteCall.CallCount).To(Equal(0))

				Expect(fakeReadCloser.CloseCall.CallCount).To(Equal(1))
				Expect(fakeWriteCloser.CloseCall.CallCount).To(Equal(0))
			})
		})
	})

	Describe("Version", func() {
		It("returns the objects version", func() {
			version, err := store.Version()
			Expect(err).NotTo(HaveOccurred())
			Expect(version.Ref).To(Equal("fresh-version"))
		})

		Context("when the underlying object errors", func() {
			BeforeEach(func() {
				fakeObject.VersionCall.Returns.Error = errors.New("mango")
			})

			It("returns the error", func() {
				_, err := store.Version()
				Expect(err).To(MatchError("mango"))
			})
		})
	})

	Describe("CopyDir", func() {
		var (
			source string
			dest   string
		)
		BeforeEach(func() {
			var err error
			source, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			err = os.Mkdir(filepath.Join(source, "subdirectory"), os.ModePerm)

			dest = filepath.Join(os.TempDir(), "bbl-concourse-resource-copy-dir-dest")

			filePath := filepath.Join(source, "some-file")
			err = ioutil.WriteFile(filePath, []byte("some-contents"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			nestedFilePath := filepath.Join(source, "subdirectory", "some-other-file")
			err = ioutil.WriteFile(nestedFilePath, []byte("some-more-contents"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.RemoveAll(source)
			os.RemoveAll(dest)
		})

		It("copies the contents of sourcePath to destPath", func() {
			err := storage.CopyDir(source, dest)
			Expect(err).NotTo(HaveOccurred())

			filePath := filepath.Join(dest, "some-file")
			contents, err := ioutil.ReadFile(filePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("some-contents"))

			nestedFilePath := filepath.Join(dest, "subdirectory", "some-other-file")
			contents, err = ioutil.ReadFile(nestedFilePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("some-more-contents"))
		})
	})
})
