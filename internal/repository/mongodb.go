package repository

import (
	"context"
	"fmt"
	"mf-importer/internal/model"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBClient struct {
	client *mongo.Client
}

func NewMongoDB(ctx context.Context, uri string) (*MongoDBClient, error) {
	credential := options.Credential{
		Username: "root",
		Password: os.Getenv("db_pass"),
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri).SetAuth(credential))
	if err != nil {
		return nil, err
	}

	m := &MongoDBClient{
		client: client,
	}

	return m, nil
}

func (c *MongoDBClient) Disconnect(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

func (c *MongoDBClient) GetCFRecords(ctx context.Context) (cfRecords []model.CFRecords, err error) {
	// 未登録の record を取得するための filter
	// filter := bson.D{{"maw_regist", bson.D{{"$ne", true}}}}
	filter := bson.D{}
	coll := c.client.Database("mfimporter").Collection("detail")
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return []model.CFRecords{}, err
	}

	if err := cursor.All(ctx, &cfRecords); err != nil {
		return []model.CFRecords{}, err
	}

	fmt.Println(cfRecords) // For test
	// cfRecords の中から 実際に登録すべきレコードを抽出
	return cfRecords, nil
}
