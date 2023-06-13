package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBClient struct {
	client *mongo.Client
}

func NewMongoDB(ctx context.Context, uri string) (*MongoDBClient, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	m := &MongoDBClient{
		client: client,
	}

	return m, nil
}
