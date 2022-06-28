package storage_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/bbl-state-resource/fakes"
	"github.com/cloudfoundry/bbl-state-resource/storage"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Storage", func() {
	var (
		storageDir      string
		filename        string
		nestedDirectory string
		store           storage.Storage
		fakeTarrer      *fakes.Tarrer
		fakeObject      *fakes.Object
		fakeBucket      *fakes.Bucket
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
		fakeObject.VersionCall.Returns.Version = storage.Version{Name: "passionfruit", Ref: "fresh-version", Updated: time.Unix(1, 0)}

		fakeObject2 := &fakes.Object{}
		fakeObject2.VersionCall.Returns.Version = storage.Version{Name: "kiwi", Ref: "rotten-version", Updated: time.Unix(-1, 0)}

		fakeObject3 := &fakes.Object{}
		fakeObject3.VersionCall.Returns.Version = storage.Version{Name: "breadfruit", Ref: "fresh-version", Updated: time.Unix(1, 0)}

		fakeObject4 := &fakes.Object{}
		fakeObject4.VersionCall.Returns.Version = storage.Version{Name: "noni", Ref: "rotten-version", Updated: time.Unix(-1, 0)}

		fakeObject5 := &fakes.Object{}
		fakeObject5.VersionCall.Returns.Version = storage.Version{Name: "passionfruit", Ref: "ripe-version", Updated: time.Unix(0, 0)}

		fakeBucket = &fakes.Bucket{}
		fakeBucket.ObjectsCall.Returns.Objects = []storage.Object{fakeObject, fakeObject2, fakeObject3, fakeObject4, fakeObject5}

		By("creating a temporary directory to walk", func() {
			var err error
			storageDir, err = ioutil.TempDir("", "storage_dir")
			Expect(err).NotTo(HaveOccurred())
			filename = filepath.Join(storageDir, "bbl-state.json")
			bblStateFile, err := os.Create(filename)
			Expect(err).NotTo(HaveOccurred())
			defer bblStateFile.Close()

			bblStateContents := fmt.Sprintf(`{"version": 14, "randomDir": "%s"}`, storageDir)
			_, err = bblStateFile.Write([]byte(bblStateContents))
			Expect(err).NotTo(HaveOccurred())

			nestedDirectory = filepath.Join(storageDir, "nested-dir")
			err = os.MkdirAll(nestedDirectory, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			nestedFilename := filepath.Join(nestedDirectory, "nested-data.json")
			nestedDataFile, err := os.Create(nestedFilename)
			Expect(err).NotTo(HaveOccurred())
			defer nestedDataFile.Close()

			nestedDataContents := fmt.Sprintf(`{"version": 999, "randomDir": "%s"}`, nestedDirectory)
			_, err = nestedDataFile.Write([]byte(nestedDataContents))
			Expect(err).NotTo(HaveOccurred())
		})

		store = storage.Storage{
			Name:     "passionfruit",
			Bucket:   fakeBucket,
			Object:   fakeObject,
			Archiver: fakeTarrer,
		}
	})

	AfterEach(func() {
		_ = os.RemoveAll(storageDir) // ignore the error
	})

	Describe("Upload", func() {
		It("tars the contents of filepath and uploads them", func() {
			version, err := store.Upload(storageDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(version.Ref).To(Equal("fresh-version"))

			Expect(fakeTarrer.ArchiveCall.Receives.Output).To(Equal(fakeWriteCloser))
			Expect(fakeTarrer.ArchiveCall.Receives.Files[1].NameInArchive).To(Equal(filepath.Base(filename)))
			Expect(fakeTarrer.ArchiveCall.Receives.Files[2].NameInArchive).To(Equal(filepath.Base(nestedDirectory)))

			Expect(fakeWriteCloser.CloseCall.CallCount).To(Equal(1))
		})

		Context("when archiving the file returns an error", func() {
			BeforeEach(func() {
				fakeTarrer.ArchiveCall.Returns.Error = errors.New("coconut")
			})

			It("returns an error", func() {
				_, err := store.Upload(storageDir)
				Expect(err).To(MatchError("coconut"))

				Expect(fakeWriteCloser.CloseCall.CallCount).To(Equal(0))
			})
		})

		Context("when closing the writer returns an error", func() {
			BeforeEach(func() {
				fakeWriteCloser.CloseCall.Returns.Error = errors.New("mango")
			})

			It("returns an error", func() {
				_, err := store.Upload(storageDir)
				Expect(err).To(MatchError("mango"))
			})
		})
	})

	Describe("Download", func() {
		Context("when the object already exists", func() {
			It("downloads the object and untars it", func() {
				version, err := store.Download(storageDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(version.Ref).To(Equal("fresh-version"))

				Expect(fakeTarrer.ExtractCall.Receives.SourceArchive).To(Equal(fakeReadCloser))

				Expect(fakeReadCloser.CloseCall.CallCount).To(Equal(1))
				Expect(fakeWriteCloser.CloseCall.CallCount).To(Equal(0))
			})
		})

		Context("when the object does not exist", func() {
			BeforeEach(func() {
				fakeObject.NewReaderCall.Returns.Error = storage.ObjectNotFoundError
			})

			It("uploads an the appropriate object", func() {
				version, err := store.Download(storageDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(version.Ref).To(Equal("fresh-version"))

				Expect(fakeTarrer.ExtractCall.CallCount).To(Equal(0))

				Expect(fakeTarrer.ArchiveCall.Receives.Output).To(Equal(fakeWriteCloser))
				Expect(fakeTarrer.ArchiveCall.Receives.Files[1].Name()).To(Equal(filepath.Base(filename)))
				Expect(fakeTarrer.ArchiveCall.Receives.Files[2].Name()).To(Equal(filepath.Base(nestedDirectory)))

				Expect(fakeReadCloser.CloseCall.CallCount).To(Equal(0))
				Expect(fakeWriteCloser.CloseCall.CallCount).To(Equal(1))
			})
		})

		Context("when reading the object returns an error", func() {
			BeforeEach(func() {
				fakeObject.NewReaderCall.Returns.Error = errors.New("papaya")
			})

			It("returns the error", func() {
				_, err := store.Download(storageDir)
				Expect(err).To(MatchError("papaya"))

				Expect(fakeTarrer.ExtractCall.CallCount).To(Equal(0))
				Expect(fakeTarrer.ArchiveCall.CallCount).To(Equal(0))

				Expect(fakeReadCloser.CloseCall.CallCount).To(Equal(0))
				Expect(fakeWriteCloser.CloseCall.CallCount).To(Equal(0))
			})
		})

		Context("when reading the archive returns an error", func() {
			BeforeEach(func() {
				fakeTarrer.ExtractCall.Returns.Error = errors.New("mango")
			})

			It("returns the error", func() {
				_, err := store.Download(storageDir)
				Expect(err).To(MatchError("mango"))

				Expect(fakeTarrer.ArchiveCall.CallCount).To(Equal(0))

				Expect(fakeReadCloser.CloseCall.CallCount).To(Equal(1))
				Expect(fakeWriteCloser.CloseCall.CallCount).To(Equal(0))
			})
		})

		Context("when the fetching the version of the object errors", func() {
			BeforeEach(func() {
				fakeObject.VersionCall.Returns.Error = errors.New("mango")
			})

			It("returns the error", func() {
				_, err := store.Download(storageDir)
				Expect(err).To(MatchError("mango"))
			})
		})
	})

	Describe("Version", func() {
		It("returns the objects version", func() {
			version, err := store.Version()
			Expect(err).NotTo(HaveOccurred())
			Expect(version.Ref).To(Equal("fresh-version"))
			Expect(version.Name).To(Equal("passionfruit"))
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

	Describe("GetAllNewerVersions", func() {
		var version storage.Version
		BeforeEach(func() {
			version = storage.Version{Name: "passionfruit", Ref: "old-version", Updated: time.Unix(0, 0)}
		})

		It("returns the versions for each newer object in the bucket", func() {
			versions, err := store.GetAllNewerVersions(version)
			Expect(err).NotTo(HaveOccurred())
			Expect(versions).To(ConsistOf([]storage.Version{
				{Name: "passionfruit", Ref: "fresh-version", Updated: time.Unix(1, 0)},
				{Name: "breadfruit", Ref: "fresh-version", Updated: time.Unix(1, 0)},
				{Name: "passionfruit", Ref: "ripe-version", Updated: time.Unix(0, 0)},
			}))
		})

		Context("when we fail to list buckets", func() {
			BeforeEach(func() {
				fakeBucket.ObjectsCall.Returns.Error = errors.New("durian")
			})

			It("returns the error", func() {
				_, err := store.GetAllNewerVersions(version)
				Expect(err).To(MatchError("durian"))
			})
		})

		Context("when we fail to query an object's version", func() {
			BeforeEach(func() {
				fakeBucket.ObjectsCall.Returns.Error = errors.New("durian")
			})

			It("returns the error", func() {
				_, err := store.GetAllNewerVersions(version)
				Expect(err).To(MatchError("durian"))
			})
		})
	})
})
