package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

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

func (client Client) push(ctx context.Context, key string, value interface{}) error {
	_, err := client.conn.LPush(ctx, key, value).Result()
	return err
}

func (client Client) Receive(ctx context.Context, botName string, msgs chan<- string) {
	for {
		messages, err := client.conn.BLPop(context.Background(), time.Second * 3, botName).Result()
		if err != nil {
			client.log.Error("PopMessage", err)
		}
		for _, msg := range messages {
			msgs<- msg
		}
	}
}

func (client Client) Topic(ctx context.Context, hashTopic string) (string, error) {
	res, err := client.conn.Get(ctx, hashTopic).Result()
	if err != nil && err == redis.Nil {
		err = nil
		res = ""
	}
	return res, err
}

func (client Client) NewTopic(ctx context.Context, hashTopic string, topicID int) error {
	err := client.push(ctx,hashTopic, topicID)
	if err != nil {
		return err
	}
	return client.addTopicToList(ctx, hashTopic)
}

func (client Client) addTopicToList(ctx context.Context, hashTopic string) error {
	return nil
}

