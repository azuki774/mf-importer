package repository

import (
	"context"
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

func (c *MongoDBClient) GetCFRecords(ctx context.Context) (cfRecords []model.CFRecord, err error) {
	// 未登録の record を取得するための filter
	filter := bson.D{{"maw_status", bson.D{{"$ne", true}}}}
	// filter := bson.D{}
	coll := c.client.Database("mfimporter").Collection("detail")
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return []model.CFRecord{}, err
	}

	if err := cursor.All(ctx, &cfRecords); err != nil {
		return []model.CFRecord{}, err
	}

	return cfRecords, nil
}

func (c *MongoDBClient) CheckCFRecords(ctx context.Context, cfRecords []model.CFRecord) (err error) {
	// TODO
	return nil
}

func (c *MongoDBClient) RegistedCFRecords(ctx context.Context, cfRecords []model.CFRecord) (err error) {
	// TODO
	return nil
}
