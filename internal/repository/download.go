package repository

import (
	"context"
	"fmt"
	"io"
	"mf-importer/internal/logger"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

var l *zap.Logger

// downloader for csv
type downloader struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	BucketName      string
	BucketDir       string
	Endpoint        string
	SaveDir         string // CSVを保存するローカル側のパス
}

func NewDownloader(saveDir string) *downloader {
	return &downloader{
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Region:          os.Getenv("AWS_REGION"),
		BucketName:      os.Getenv("BUCKET_NAME"),
		BucketDir:       os.Getenv("BUCKET_DIR"),
		Endpoint:        os.Getenv("BUCKET_URL"),
		SaveDir:         saveDir,
	}
}

func (d *downloader) Start(ctx context.Context) error {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(d.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			d.AccessKeyID, d.SecretAccessKey, "",
		)),
		// S3 互換ストレージ (MinIO 等) はレスポンスにチェックサムを返さないため、
		// "WhenSupported" だと "Response has no supported checksum" 警告が出る。
		// "WhenRequired" で必須時のみ検証し、警告を抑制する。
		config.WithResponseChecksumValidation(aws.ResponseChecksumValidationWhenRequired),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(d.Endpoint)
	})

	// BucketDir の末尾のスラッシュを保証し、プレフィックスとして使用
	bucketDirPrefix := strings.TrimSuffix(d.BucketDir, "/") + "/"

	// ダウンロード先ディレクトリが存在しない場合は作成
	if err := os.MkdirAll(d.SaveDir, 0755); err != nil {
		return fmt.Errorf("failed to create save directory %s: %w", d.SaveDir, err)
	}

	// ファイル一覧を取得 (1 ページのみ。1000件超は warn)
	listOutput, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(d.BucketName),
		Prefix: aws.String(bucketDirPrefix),
		// MaxKeys = 1000 (default)
	})
	if err != nil {
		return fmt.Errorf("failed to list objects in s3://%s/%s: %w", d.BucketName, bucketDirPrefix, err)
	}

	if listOutput.IsTruncated != nil && *listOutput.IsTruncated {
		l.Warn("This list is contained more than 1000 objects", zap.String("bucketDirPrefix", bucketDirPrefix))
	}

	l.Info("get download file list complete", zap.Int("count", len(listOutput.Contents)))

	// 逐次ダウンロード
	for _, obj := range listOutput.Contents {
		objectKey := *obj.Key

		// ディレクトリ自体や末尾が "/" のオブジェクトはスキップ
		if objectKey == d.BucketDir || strings.HasSuffix(objectKey, "/") {
			continue
		}

		if err := d.downloadOne(ctx, s3Client, objectKey); err != nil {
			return err
		}
	}

	l.Info("download files complete", zap.Int("count", len(listOutput.Contents)))
	return nil
}

func (d *downloader) downloadOne(ctx context.Context, s3Client *s3.Client, objectKey string) error {
	out, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(d.BucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("s3 download failed for %s: %w", objectKey, err)
	}
	defer out.Body.Close()

	localFilePath := filepath.Join(d.SaveDir, filepath.Base(objectKey))
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

func init() {
	l = logger.NewLogger()
}
