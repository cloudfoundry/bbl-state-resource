package storage

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	gcs "cloud.google.com/go/storage"
	"github.com/mholt/archiver"
	oauthgoogle "golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

// untested gcs api instantiation
type objectHandleWrapper struct {
	objectHandle *gcs.ObjectHandle
}

func (o objectHandleWrapper) Version() (string, error) {
	r, err := o.objectHandle.Attrs(context.Background())
	if err == gcs.ErrObjectNotExist {
		return "", ObjectNotFoundError
	} else if err != nil {
		return "", err
	}

	return hex.EncodeToString(r.MD5), nil
}

func (o objectHandleWrapper) NewReader() (io.ReadCloser, error) {
	r, err := o.objectHandle.NewReader(context.Background())
	if err == gcs.ErrObjectNotExist {
		return nil, ObjectNotFoundError
	}
	return r, err
}

func (o objectHandleWrapper) NewWriter() io.WriteCloser {
	return o.objectHandle.NewWriter(context.Background())
}

func NewGCSStorage(serviceAccountKey string, bucketName string) (Storage, error) {
	storageJwtConf, err := oauthgoogle.JWTConfigFromJSON([]byte(serviceAccountKey), gcs.ScopeReadWrite)
	if err != nil {
		return Storage{}, err
	}
	ctx := context.Background()
	tokenSource := storageJwtConf.TokenSource(ctx)

	storageClient, err := gcs.NewClient(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return Storage{}, fmt.Errorf("failed to instantiate storageclient: %s", err)
	}

	p := struct {
		ProjectId string `json:"project_id"`
	}{}
	if err := json.Unmarshal([]byte(serviceAccountKey), &p); err != nil {
		return Storage{}, fmt.Errorf("Unmarshalling account key for project id: %s", err)
	}
	bucket := storageClient.Bucket(bucketName).UserProject(p.ProjectId)

	_, err = bucket.Attrs(ctx)
	if err == gcs.ErrBucketNotExist {
		err = bucket.Create(ctx, p.ProjectId, nil)
	} else if err != nil {
		return Storage{}, fmt.Errorf("failed to get bucket: %s", err)
	}

	object := bucket.Object("bbl-state.tar.gz")

	return Storage{
		DirectoryName: bucketName,
		Object: objectHandleWrapper{
			objectHandle: object,
		},
		Archiver: archiver.TarGz,
	}, nil
}
