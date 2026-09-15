package repository

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

type nrknDownloader struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	BucketName      string
	BucketDir       string
	Endpoint        string
	SaveDir         string
}

func NewNrknDownloader(saveDir string) *nrknDownloader {
	return &nrknDownloader{
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Region:          os.Getenv("AWS_REGION"),
		BucketName:      os.Getenv("NRKN_BUCKET_NAME"),
		BucketDir:       os.Getenv("NRKN_BUCKET_DIR"),
		Endpoint:        envOrNrkn("NRKN_BUCKET_URL", os.Getenv("BUCKET_URL")),
		SaveDir:         saveDir,
	}
}

func envOrNrkn(nrknKey, fallback string) string {
	if v := os.Getenv(nrknKey); v != "" {
		return v
	}
	return fallback
}

// isNrknBucketConfigured は NRKN 用バケット設定が揃っているかを返す
// Endpoint は任意 (docs/s3-download.md と同様、未設定なら AWS 標準の
// virtual-hosted-style アクセスを使う)
func (d *nrknDownloader) isNrknBucketConfigured() bool {
	return d.BucketName != "" && d.BucketDir != "" && d.Region != ""
}

// StartMonth downloads objects below the configured bucket directory for one
// calendar month in YYYYMM form.
func (d *nrknDownloader) StartMonth(ctx context.Context, month string) error {
	bucketDirPrefix, err := nrknMonthPrefix(d.BucketDir, month)
	if err != nil {
		return err
	}
	return d.start(ctx, bucketDirPrefix)
}

func nrknMonthPrefix(bucketDir, month string) (string, error) {
	if len(month) != 6 {
		return "", fmt.Errorf("invalid NRKN month %q: want YYYYMM", month)
	}
	for _, r := range month {
		if r < '0' || r > '9' {
			return "", fmt.Errorf("invalid NRKN month %q: want YYYYMM", month)
		}
	}
	if _, err := time.Parse("200601", month); err != nil {
		return "", fmt.Errorf("invalid NRKN month %q: want YYYYMM: %w", month, err)
	}
	base := strings.TrimSuffix(bucketDir, "/")
	if base == "" {
		return month[:4] + "/" + month[4:] + "/", nil
	}
	return base + "/" + month[:4] + "/" + month[4:] + "/", nil
}

func (d *nrknDownloader) start(ctx context.Context, bucketDirPrefix string) error {
	if !d.isNrknBucketConfigured() {
		return fmt.Errorf("NRKN_BUCKET_NAME, NRKN_BUCKET_DIR and AWS_REGION are required")
	}

	opts := []func(*config.LoadOptions) error{
		config.WithRegion(d.Region),
		config.WithResponseChecksumValidation(aws.ResponseChecksumValidationWhenRequired),
	}
	// 静的クレデンシャルが揃う場合のみ provider を上書きし、
	// そうでなければ SDK 既定の chain (プロファイル等) を温存する
	if d.AccessKeyID != "" && d.SecretAccessKey != "" {
		opts = append(opts, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			d.AccessKeyID, d.SecretAccessKey, "",
		)))
	}
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to load AWS config for nrkn: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if d.Endpoint != "" {
			o.BaseEndpoint = aws.String(d.Endpoint)
		}
	})

	if err := os.MkdirAll(d.SaveDir, 0700); err != nil {
		return fmt.Errorf("failed to create nrkn save directory %s: %w", d.SaveDir, err)
	}

	paginator := s3.NewListObjectsV2Paginator(s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(d.BucketName),
		Prefix: aws.String(bucketDirPrefix),
	})

	count := 0
	for paginator.HasMorePages() {
		listOutput, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list nrkn objects in s3://%s/%s: %w", d.BucketName, bucketDirPrefix, err)
		}
		for _, obj := range listOutput.Contents {
			if obj.Key == nil {
				continue
			}
			objectKey := *obj.Key
			if !strings.HasPrefix(objectKey, bucketDirPrefix) || !strings.EqualFold(filepath.Ext(objectKey), ".json") {
				continue
			}
			if err := d.downloadOne(ctx, s3Client, bucketDirPrefix, objectKey); err != nil {
				return err
			}
			count++
		}
	}

	l.Info("download nrkn files complete", zap.Int("count", count))
	return nil
}

func (d *nrknDownloader) downloadOne(ctx context.Context, s3Client *s3.Client, bucketDirPrefix, objectKey string) error {
	rel := strings.TrimPrefix(objectKey, bucketDirPrefix)
	if rel == objectKey || !isSafeRelPath(rel) {
		return fmt.Errorf("nrkn s3 object key escapes save directory")
	}
	out, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(d.BucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("nrkn s3 download failed for %s: %w", objectKey, err)
	}
	defer out.Body.Close()

	localFilePath := filepath.Join(d.SaveDir, rel)

	if err := os.MkdirAll(filepath.Dir(localFilePath), 0700); err != nil {
		return fmt.Errorf("failed to create dir for %s: %w", localFilePath, err)
	}

	f, err := os.OpenFile(localFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open local file %s: %w", localFilePath, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, out.Body); err != nil {
		return fmt.Errorf("failed to write %s: %w", localFilePath, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close NRKN download: %w", err)
	}
	return nil
}
