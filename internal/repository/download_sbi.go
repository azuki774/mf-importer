package repository

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

type sbiDownloader struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	BucketName      string
	BucketDir       string
	Endpoint        string
	SaveDir         string
}

func NewSbiDownloader(saveDir string) *sbiDownloader {
	return &sbiDownloader{
		AccessKeyID:     envOrSbi("AWS_ACCESS_KEY_ID", os.Getenv("AWS_ACCESS_KEY_ID")),
		SecretAccessKey: envOrSbi("AWS_SECRET_ACCESS_KEY", os.Getenv("AWS_SECRET_ACCESS_KEY")),
		Region:          envOrSbi("AWS_REGION", os.Getenv("AWS_REGION")),
		BucketName:      envOrSbi("SBI_BUCKET_NAME", os.Getenv("BUCKET_NAME")),
		BucketDir:       envOrSbi("SBI_BUCKET_DIR", os.Getenv("BUCKET_DIR")),
		Endpoint:        envOrSbi("SBI_BUCKET_URL", os.Getenv("BUCKET_URL")),
		SaveDir:         saveDir,
	}
}

func envOrSbi(sbiKey, fallback string) string {
	if v := os.Getenv(sbiKey); v != "" {
		return v
	}
	return fallback
}

// isSbiBucketConfigured は SBI 用バケット設定が揃っているかを返す
// Endpoint は任意 (docs/s3-download.md と同様、未設定なら AWS 標準の
// virtual-hosted-style アクセスを使う)
func (d *sbiDownloader) isSbiBucketConfigured() bool {
	return d.BucketName != "" && d.BucketDir != "" && d.Region != ""
}

func (d *sbiDownloader) Start(ctx context.Context) error {
	if !d.isSbiBucketConfigured() {
		l.Info("sbi downloader skipped: SBI bucket not configured")
		return nil
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
		return fmt.Errorf("failed to load AWS config for sbi: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if d.Endpoint != "" {
			o.BaseEndpoint = aws.String(d.Endpoint)
		}
	})

	bucketDirPrefix := strings.TrimSuffix(d.BucketDir, "/") + "/"

	if err := os.MkdirAll(d.SaveDir, 0755); err != nil {
		return fmt.Errorf("failed to create sbi save directory %s: %w", d.SaveDir, err)
	}

	listOutput, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(d.BucketName),
		Prefix: aws.String(bucketDirPrefix),
	})
	if err != nil {
		return fmt.Errorf("failed to list sbi objects in s3://%s/%s: %w", d.BucketName, bucketDirPrefix, err)
	}

	if listOutput.IsTruncated != nil && *listOutput.IsTruncated {
		l.Warn("sbi list is truncated (>1000 objects)", zap.String("bucketDirPrefix", bucketDirPrefix))
	}

	l.Info("get sbi download file list complete", zap.Int("count", len(listOutput.Contents)))

	for _, obj := range listOutput.Contents {
		objectKey := *obj.Key
		if objectKey == d.BucketDir || strings.HasSuffix(objectKey, "/") {
			continue
		}
		if err := d.downloadOne(ctx, s3Client, bucketDirPrefix, objectKey); err != nil {
			return err
		}
	}

	l.Info("download sbi files complete", zap.Int("count", len(listOutput.Contents)))
	return nil
}

// isSafeRelPath は rel が SaveDir 配下に収まる相対パスかを返す
// ".." 要素や絶対パスを含む場合は false (path traversal 防止)
func isSafeRelPath(rel string) bool {
	if filepath.IsAbs(rel) {
		return false
	}
	for _, part := range strings.Split(filepath.ToSlash(rel), "/") {
		if part == ".." {
			return false
		}
	}
	return rel != ""
}

func (d *sbiDownloader) downloadOne(ctx context.Context, s3Client *s3.Client, bucketDirPrefix, objectKey string) error {
	out, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(d.BucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("sbi s3 download failed for %s: %w", objectKey, err)
	}
	defer out.Body.Close()

	// ネストを保持: bucketDirPrefix からの相対パスを SaveDir 配下に再現
	rel := strings.TrimPrefix(objectKey, bucketDirPrefix)
	// 相対パスが空なら basename にフォールバック
	if rel == "" || rel == objectKey {
		rel = filepath.Base(objectKey)
	}
	if !isSafeRelPath(rel) {
		return fmt.Errorf("sbi s3 object key escapes save directory: %s", objectKey)
	}
	localFilePath := filepath.Join(d.SaveDir, rel)

	if err := os.MkdirAll(filepath.Dir(localFilePath), 0755); err != nil {
		return fmt.Errorf("failed to create dir for %s: %w", localFilePath, err)
	}

	f, err := os.OpenFile(localFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open local file %s: %w", localFilePath, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, out.Body); err != nil {
		return fmt.Errorf("failed to write %s: %w", localFilePath, err)
	}
	return nil
}
