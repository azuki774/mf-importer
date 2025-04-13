package repository

import (
	"context"
	"errors"
	"fmt"
	"mf-importer/internal/logger"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
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
	creds := credentials.NewStaticCredentials(d.AccessKeyID, d.SecretAccessKey, "")

	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String(d.Region),
		Endpoint:    aws.String(d.Endpoint),
	}))

	getDownloadList, err := d.getDownloadList(ctx, sess)
	if err != nil {
		return err
	}

	if err := d.download(ctx, sess, getDownloadList); err != nil {
		return err
	}

	return nil
}

// getDownloadList: そのbucketName/Dirに存在するファイルリストを取得。ただし、1000件以下であることを前提とする
func (d *downloader) getDownloadList(ctx context.Context, sess *session.Session) (*s3.ListObjectsV2Output, error) {
	s3Client := s3.New(sess)

	// BucketDir の末尾のスラッシュを保証し、プレフィックスとして使用
	bucketDirPrefix := strings.TrimSuffix(d.BucketDir, "/") + "/"

	listInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(d.BucketName),
		Prefix: aws.String(bucketDirPrefix),
		// MaxKeys = 1000 (default)
	}

	listOutput, err := s3Client.ListObjectsV2WithContext(ctx, listInput) // ctx を使用
	if err != nil {
		return nil, fmt.Errorf("failed to list objects in s3://%s/%s: %w", d.BucketName, bucketDirPrefix, err)
	}

	if listOutput.IsTruncated != nil && *listOutput.IsTruncated {
		// 前提条件チェック：万が一1000件を超えていたら警告
		l.Warn("This list is contained more than 1000 objects", zap.String("bucketDirPrefix", bucketDirPrefix))
		// 表示だけ
	}

	l.Info("get download file list complete", zap.Int("count", len(listOutput.Contents)))
	return listOutput, nil
}

func (d *downloader) download(ctx context.Context, sess *session.Session, list *s3.ListObjectsV2Output) error {
	objects := []s3manager.BatchDownloadObject{}
	openedFiles := []*os.File{} // 開いたファイルを保持するスライス (後で閉じるため)

	// ダウンロード先ディレクトリが存在しない場合は作成
	if err := os.MkdirAll(d.SaveDir, 0644); err != nil {
		return fmt.Errorf("failed to create save directory %s: %w", d.SaveDir, err)
	}

	// エラー発生時に備え、開いたファイルを確実に閉じるための defer を設定
	defer func() {
		for _, f := range openedFiles {
			if f != nil { // nil チェックを追加 (OpenFile失敗時など)
				if closeErr := f.Close(); closeErr != nil {
					zap.Error(errors.New("error closing file during cleanup"))
				}
			}
		}
	}()

	// リストされたオブジェクトをループしてダウンロード準備
	for _, obj := range list.Contents {
		objectKey := *obj.Key

		// ディレクトリ自体や末尾が "/" のオブジェクトはスキップ (サブディレクトリなし前提でも念のため)
		if objectKey == d.BucketDir || strings.HasSuffix(objectKey, "/") {
			continue
		}

		fileName := filepath.Base(objectKey)
		localFilePath := filepath.Join(d.SaveDir, fileName)

		f, err := os.OpenFile(localFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			// defer func() が後で閉じるので、ここではエラーを返すだけで良い
			return fmt.Errorf("failed to open local file %s for writing: %w", localFilePath, err)
		}
		openedFiles = append(openedFiles, f)

		// s3manager用のダウンロードオブジェクトを作成
		objects = append(objects, s3manager.BatchDownloadObject{
			Object: &s3.GetObjectInput{
				Bucket: aws.String(d.BucketName),
				Key:    aws.String(objectKey),
			},
			Writer: f, // 開いた *os.File は io.WriterAt を満たす
		})
	}

	if len(objects) == 0 {
		return nil // 正常終了
	}

	iter := &s3manager.DownloadObjectsIterator{Objects: objects}
	svc := s3manager.NewDownloader(sess)
	err := svc.DownloadWithIterator(ctx, iter) // 引数の ctx を使用するのが望ましい
	if err != nil {
		return fmt.Errorf("s3 download failed: %w", err)
	}

	l.Info("download files complete", zap.Int("count", len(list.Contents)))
	return nil
}

func init() {
	l = logger.NewLogger()
}
