package v1

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoOpts struct {
	MongoClientOptions *options.ClientOptions
}

type MongoConfig struct {
	DatabaseName   string
	CollectionName string
}

func NewMongo(ctx context.Context, config MongoOpts) (*mongo.Client, error) {
	m, err := mongo.NewClient(config.MongoClientOptions)
	if err != nil {
		return nil, err
	}
	err = m.Connect(ctx)
	if err != nil {
		return nil, err
	}

	err = m.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return m, nil
}
