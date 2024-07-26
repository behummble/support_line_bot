package redis

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	log *slog.Logger
	conn *redis.Client
}

func New(log *slog.Logger, host, port, password string) (Client, error) {
	conn, err := connect(host, port, password)
	if err != nil {
		return Client{}, err
	}

	return Client{log, conn}, nil
}

func (client Client) Topic(ctx context.Context, topic string) (string, error) {
	res, err := client.conn.Get(ctx, topic).Result()
	if err != nil && err == redis.Nil {
		err = nil
		res = ""
	}
	return res, err
}

func (client Client) NewTopic(ctx context.Context, topicUserKey, topicSupportKey, topicListKey, topicData string) error {
	err := client.set(ctx, topicUserKey, topicData)
	if err != nil {
		client.log.Error("Add topic user in DB", "Error", err)
	}

	err = client.set(ctx, topicSupportKey, topicData)
	if err != nil {
		client.log.Error("Add topic support in DB", "Error",err)
	}

	err = client.push(ctx, topicListKey, topicSupportKey)
	return err
}

func (client Client) ClearTopics(ctx context.Context) error {
	_, err := client.conn.FlushAll(ctx).Result()
	return err
}
func (client Client) AllTopics(ctx context.Context, key string) ([]string, error) {
	return client.conn.LRange(ctx, key, 0, -1).Result()
}

func connect(host, port, password string) (*redis.Client, error) {
	options := &redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port),
		Password: password,
	}

	conn := redis.NewClient(options)

	_, err := conn.Ping(context.Background()).Result()

	if err != nil {
		return nil, err
	}
	
	return conn, nil
}

func (client Client) set(ctx context.Context, key string, value interface{}) error {
	_, err := client.conn.Set(ctx, key, value, 0).Result()
	return err
}

func (client Client) push(ctx context.Context, key string, value interface{}) error {
	_, err := client.conn.LPush(ctx, key, value).Result()
	return err
}