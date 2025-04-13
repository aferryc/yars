package gcs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
)

type UploadURLInfo struct {
	URL        string `json:"url"`
	Method     string `json:"method"`
	ObjectName string `json:"objectName"`
	Expiry     int64  `json:"expiry"`
}

type GCSRepo struct {
	bucketName string
	client     *storage.Client
}

func NewGCSRepository(bucketName string, client *storage.Client) (*GCSRepo, error) {
	return &GCSRepo{
		bucketName: bucketName,
		client:     client,
	}, nil
}

func (u *GCSRepo) GenerateUploadURL(objectName string, contentType string, expires time.Time) (string, error) {
	// Check if we're using the emulator
	if emulatorHost := os.Getenv("STORAGE_EMULATOR_HOST"); emulatorHost != "" {
		// For emulator, just construct a direct URL without signing
		return fmt.Sprintf("%s/upload/storage/v1/b/%s/o?name=%s&uploadType=media", "http://localhost:4443", u.bucketName, objectName), nil
	}

	opts := &storage.SignedURLOptions{
		GoogleAccessID: "some@example.com",
		Scheme:         storage.SigningSchemeV4,
		Method:         "PUT",
		Expires:        expires,
		ContentType:    contentType,
		Headers: []string{
			"Content-Type:" + contentType,
		},
	}

	uri, err := u.client.Bucket(u.bucketName).SignedURL(objectName, opts)
	if err != nil {
		return "", fmt.Errorf("bucket(%q).SignedURL: %v", u.bucketName, err)
	}

	return uri, err
}

func (u *GCSRepo) GenerateDownloadURL(objectName string) (string, error) {
	// Check if we're using the emulator
	if emulatorHost := os.Getenv("STORAGE_EMULATOR_HOST"); emulatorHost != "" {
		// For emulator, just construct a direct URL without signing
		return fmt.Sprintf("%s/storage/v1/b/%s/o/%s?alt=media", "http://localhost:4443", u.bucketName, objectName), nil
	}
	opts := &storage.SignedURLOptions{
		GoogleAccessID: "some@example.com",
		Scheme:         storage.SigningSchemeV4,
		Method:         "GET",
		Expires:        time.Now().Add(15 * time.Minute),
	}

	url, err := u.client.Bucket(u.bucketName).SignedURL(objectName, opts)
	if err != nil {
		return "", fmt.Errorf("bucket(%q).SignedURL: %v", u.bucketName, err)
	}

	return url, nil
}

func (u *GCSRepo) DownloadFromBucket(ctx context.Context, objectName string) (*os.File, error) {
	tempFile, err := os.CreateTemp("", "download-*.csv")
	if err != nil {
		return nil, fmt.Errorf("error creating temp file: %w", err)
	}
	defer func() {
		if err != nil {
			tempFile.Close()
			os.Remove(tempFile.Name())
		}
	}()

	// Generate download URL instead of using bucket directly
	downloadURL := fmt.Sprintf("%s/storage/v1/b/%s/o/%s?alt=media", "http://bucket:4443", u.bucketName, objectName)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error downloading file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status code: %d", resp.StatusCode)
	}

	// Copy content to temp file
	if _, err = io.Copy(tempFile, resp.Body); err != nil {
		return nil, fmt.Errorf("error copying object content: %w", err)
	}

	// Rewind file to beginning
	if _, err = tempFile.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("error rewinding temp file: %w", err)
	}

	return tempFile, nil
}

func (u *GCSRepo) Close() error {
	return u.client.Close()
}
