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
		storageDir          string
		filename            string
		store               storage.Storage
		fakeTarrer          *fakes.Tarrer
		fakeGCSBucketObject *fakes.GCSBucketObject
		fakeReadCloser      *fakes.ReadCloser
		fakeWriteCloser     *fakes.WriteCloser
	)

	BeforeEach(func() {
		fakeTarrer = &fakes.Tarrer{}

		fakeReadCloser = &fakes.ReadCloser{}
		fakeWriteCloser = &fakes.WriteCloser{}

		fakeGCSBucketObject = &fakes.GCSBucketObject{}
		fakeGCSBucketObject.NewReaderCall.Returns.ReadCloser = fakeReadCloser
		fakeGCSBucketObject.NewWriterCall.Returns.WriteCloser = fakeWriteCloser

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
			Object:   fakeGCSBucketObject,
			Archiver: fakeTarrer,
		}
	})

	AfterEach(func() {
		_ = os.RemoveAll(storageDir) // ignore the error
	})

	Describe("Upload", func() {
		It("tars the contents of filepath and uploads them", func() {
			err := store.Upload(storageDir)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeTarrer.WriteCall.Receives.Output).To(Equal(fakeWriteCloser))
			Expect(fakeTarrer.WriteCall.Receives.Sources).To(ConsistOf(filename))

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
			It("downloads the object, tars it, and re-uploads it", func() {
				err := store.Download(storageDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeTarrer.ReadCall.Receives.Input).To(Equal(fakeReadCloser))
				Expect(fakeTarrer.ReadCall.Receives.Destination).To(Equal(storageDir))

				Expect(fakeTarrer.WriteCall.Receives.Output).To(Equal(fakeWriteCloser))
				Expect(fakeTarrer.WriteCall.Receives.Sources).To(ConsistOf(filename))

				Expect(fakeReadCloser.CloseCall.CallCount).To(Equal(1))
				Expect(fakeWriteCloser.CloseCall.CallCount).To(Equal(1))
			})
		})

		Context("when the object does not exist", func() {
			BeforeEach(func() {
				fakeGCSBucketObject.NewReaderCall.Returns.Error = storage.ObjectNotFoundError
			})

			It("uploads the object", func() {
				err := store.Download(storageDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeTarrer.ReadCall.CallCount).To(Equal(0))

				Expect(fakeTarrer.WriteCall.Receives.Output).To(Equal(fakeWriteCloser))
				Expect(fakeTarrer.WriteCall.Receives.Sources).To(ConsistOf(filename))

				Expect(fakeReadCloser.CloseCall.CallCount).To(Equal(0))
				Expect(fakeWriteCloser.CloseCall.CallCount).To(Equal(1))
			})
		})

		Context("when reading the object returns an error", func() {
			BeforeEach(func() {
				fakeGCSBucketObject.NewReaderCall.Returns.Error = errors.New("papaya")
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
				fakeGCSBucketObject.NewReaderCall.Returns.Error = errors.New("papaya")
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
})
