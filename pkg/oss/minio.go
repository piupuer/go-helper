package oss

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/piupuer/go-helper/pkg/log"
	"io"
	"net/url"
	"time"
)

type MinioOss struct {
	ops    MinioOptions
	client *minio.Client
}

func NewMinio(options ...func(*MinioOptions)) (rp *MinioOss, err error) {
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
		return
	}
	rp = &MinioOss{
		ops:    *ops,
		client: minioClient,
	}
	return
}

func (mo *MinioOss) MakeBucket(ctx context.Context, bucketName string) {
	mo.MakeBucketWithLocation(ctx, bucketName, "")
}

func (mo *MinioOss) MakeBucketWithLocation(ctx context.Context, bucketName, location string) {
	err := mo.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := mo.client.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.WithContext(ctx).Warn("bucket %s(location %s) already exists", bucketName, location)
		} else {
			log.WithContext(ctx).WithError(err).Error("make bucket failed")
		}
	}
}

func (mo *MinioOss) Find(ctx context.Context, bucketName, prefix string, recursive bool) <-chan minio.ObjectInfo {
	return mo.client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	})
}

func (mo *MinioOss) PutLocal(ctx context.Context, bucketName, objectName, filePath string) error {
	_, err := mo.client.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{})
	return err
}

func (mo *MinioOss) Put(ctx context.Context, bucketName, objectName string, file io.Reader, fileSize int64) (err error) {
	_, err = mo.client.PutObject(ctx, bucketName, objectName, file, fileSize, minio.PutObjectOptions{})
	return
}

func (mo *MinioOss) BatchRemove(ctx context.Context, bucketName string, objectNames []string) (err error) {
	objectsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objectsCh)
		for _, name := range objectNames {
			objectsCh <- minio.ObjectInfo{
				Key: name,
			}
		}
	}()

	for e := range mo.client.RemoveObjects(ctx, bucketName, objectsCh, minio.RemoveObjectsOptions{}) {
		if e.Err != nil {
			err = e.Err
			return
		}
	}
	return
}

func (mo *MinioOss) GetPreviewUrl(ctx context.Context, bucketName, objectName string) (rp string) {
	u, err := mo.client.PresignedGetObject(ctx, bucketName, objectName, time.Second*24*60*60, url.Values{})
	if err != nil {
		return
	}
	rp = u.String()
	return
}

func (mo *MinioOss) Exists(ctx context.Context, bucketName, objectName string) (ok bool) {
	_, err := mo.client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	ok = err == nil
	return
}
