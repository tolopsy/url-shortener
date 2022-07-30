package mongo

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/tolopsy/url-shortener/shortener"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type mongoRepository struct {
	client     *mongo.Client
	database   string
	collection string
	timeout    time.Duration
}

func newMongoClient(mongoURL string, mongoTimeout time.Duration) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		return nil, err
	}

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}
	return client, nil
}

func NewMongoRepository(mongoURL, mongoDB, mongoCollection string, mongoTimeout int) (shortener.RedirectRepository, error) {
	timeout := time.Duration(mongoTimeout) * time.Second
	client, err := newMongoClient(mongoURL, timeout)
	if err != nil {
		return nil, errors.Wrap(err, "repository.NewMongoRepository")
	}

	repo := &mongoRepository{
		timeout:    timeout,
		database:   mongoDB,
		collection: mongoCollection,
		client:     client,
	}
	return repo, nil
}

func (repo *mongoRepository) Find(code string) (*shortener.Redirect, error) {
	ctx, cancel := context.WithTimeout(context.Background(), repo.timeout)
	defer cancel()

	redirect := &shortener.Redirect{}
	collection := repo.client.Database(repo.database).Collection(repo.collection)
	filter := bson.M{"code": code}
	err := collection.FindOne(ctx, filter).Decode(redirect)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.Wrap(shortener.ErrRedirectNotFound, "repository.Find")
		}
		return nil, err
	}
	return redirect, nil
}

func (repo *mongoRepository) Store(redirect *shortener.Redirect) error {
	ctx, cancel := context.WithTimeout(context.Background(), repo.timeout)
	defer cancel()

	collection := repo.client.Database(repo.database).Collection(repo.collection)
	_, err := collection.InsertOne(
		ctx,
		bson.M{
			"code":       redirect.Code,
			"url":        redirect.URL,
			"created_at": redirect.CreatedAt,
		},
	)
	if err != nil {
		return errors.Wrap(err, "repository.Store")
	}
	return nil
}
