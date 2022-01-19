package oss

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
	"io"
	"net/url"
	"time"
)

// object storage: minio
type MinioOss struct {
	ops    MinioOptions
	client *minio.Client
}

func NewMinio(options ...func(*MinioOptions)) (*MinioOss, error) {
	// Initialize minio client object.
	ops := getMinioOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}

	minioClient, err := minio.New(ops.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(ops.accessId, ops.secret, ""),
		Secure: ops.https,
	})

	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &MinioOss{
		ops:    *ops,
		client: minioClient,
	}, nil
}

// make a bucket without location
func (mo *MinioOss) MakeBucket(ctx context.Context, bucketName string) {
	mo.MakeBucketWithLocation(ctx, bucketName, "")
}

// make a bucket with location
func (mo *MinioOss) MakeBucketWithLocation(ctx context.Context, bucketName, location string) {
	err := mo.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := mo.client.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			mo.ops.logger.Warn("bucket %s(location %s) already exists", bucketName, location)
		} else {
			mo.ops.logger.Error("make bucket failed: %+v", err)
		}
	}
}

// find objects by prefix
func (mo *MinioOss) Find(ctx context.Context, bucketName, prefix string, recursive bool) <-chan minio.ObjectInfo {
	return mo.client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	})
}

// upload object from local file
func (mo *MinioOss) PutLocal(ctx context.Context, bucketName, objectName, filePath string) error {
	_, err := mo.client.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{})
	return err
}

// upload object from file stream
func (mo *MinioOss) Put(ctx context.Context, bucketName, objectName string, file io.Reader, fileSize int64) error {
	_, err := mo.client.PutObject(ctx, bucketName, objectName, file, fileSize, minio.PutObjectOptions{})
	return err
}

// batch remove object
func (mo *MinioOss) BatchRemove(ctx context.Context, bucketName string, objectNames []string) error {
	objectsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objectsCh)
		for _, name := range objectNames {
			objectsCh <- minio.ObjectInfo{
				Key: name,
			}
		}
	}()

	for rErr := range mo.client.RemoveObjects(ctx, bucketName, objectsCh, minio.RemoveObjectsOptions{}) {
		if rErr.Err != nil {
			return rErr.Err
		}
	}
	return nil
}

// get object preview url
func (mo *MinioOss) GetPreviewUrl(ctx context.Context, bucketName, objectName string) string {
	u, err := mo.client.PresignedGetObject(ctx, bucketName, objectName, time.Second*24*60*60, url.Values{})
	if err != nil {
		return ""
	}
	return u.String()
}

// check object is exists
func (mo *MinioOss) Exists(ctx context.Context, bucketName, objectName string) bool {
	_, err := mo.client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	return err == nil
}
