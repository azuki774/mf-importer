package mawinter

import (
	"mf-importer/internal/model"
	"time"

	"go.uber.org/zap"
)

type MongoDBClient interface{}
type MawinterClient interface{}

type Mawinter struct {
	Logger      *zap.Logger
	DBClient    MongoDBClient
	MawClient   MawinterClient
	ExtractRule model.ExtractRule // 抽出するルール
	ProcessDate time.Time         // 処理するファイルの登録日を指定
}
