package app

import (
	"bytes"
	"io"
	"time"
	"context"
	"cloud.google.com/go/storage"
)

// NewStorageClient with google storage
func NewStorageClient(ctx context.Context) (*storage.Client, error) {
	return storage.NewClient(ctx)
}

func getBucket(ctx context.Context, req bucketImage) (*storage.BucketHandle, error) {
	client, err := NewStorageClient(ctx)
	if err != nil { return nil, err }
	ctx, cancel := context.WithTimeout(ctx, time.Second * 10)
	defer cancel()
	bucket := client.Bucket(req.BucketName)
	return bucket, nil
}

func getImageFromGCP(ctx context.Context, bucket *storage.BucketHandle, filename string) (*bytes.Buffer, error) {
	obj := bucket.Object(filename)

	r, err := obj.NewReader(ctx)
	if err != nil { return nil, err }
	defer r.Close()
	
	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, r) 
	if err != nil { return nil, err }
	return buf, nil
}