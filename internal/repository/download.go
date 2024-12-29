package repository

import (
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// downloader for csv
type downloader struct {
	Token      string
	Region     string
	BucketName string
	BucketDir  string
	SaveDir    string // CSVを保存するローカル側のパス
}

// s3 から取得するCSVファイル名
var TargetCSVName = []string{
	"cf.csv",
	"cf_lastmonth.csv",
}

func NewDownloader(saveDir string) *downloader {
	return &downloader{
		Token:      os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Region:     os.Getenv("AWS_REGION"),
		BucketName: os.Getenv("BUCKET_NAME"),
		BucketDir:  os.Getenv("BUCKET_DIR"),
		SaveDir:    saveDir,
	}
}

func (d *downloader) Start() error {
	creds := credentials.NewStaticCredentials("AccessKey", "secret", d.Token)

	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String(d.Region),
	}))
	svc := s3manager.NewDownloader(sess)
	objects := []s3manager.BatchDownloadObject{}

	for _, fileName := range TargetCSVName {
		savePath := filepath.Join(d.SaveDir, fileName)
		s3Key := filepath.Join(d.BucketDir, fileName)
		f, err := os.OpenFile(savePath, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		defer f.Close()
		objects = append(objects, s3manager.BatchDownloadObject{
			Object: &s3.GetObjectInput{
				Bucket: aws.String(d.BucketName),
				Key:    aws.String(s3Key),
			},
			Writer: f,
		})
	}
	iter := &s3manager.DownloadObjectsIterator{Objects: objects}
	err := svc.DownloadWithIterator(aws.BackgroundContext(), iter)
	if err != nil {
		return err
	}

	return nil
}
