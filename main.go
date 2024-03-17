package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/m1ker1n-transcriber/go-telegram/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"time"

	tele "gopkg.in/telebot.v3"
)

func main() {
	cfg := config.MustLoad()

	amqpConn, err := NewAMQPConn(cfg.AMQP)
	if err != nil {
		panic(err)
	}
	defer amqpConn.Close()

	ch, err := amqpConn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		cfg.AMQP.SendQueueName, // name
		false,                  // durable
		false,                  // delete when unused
		false,                  // exclusive
		false,                  // no-wait
		nil,                    // arguments
	)
	if err != nil {
		panic(err)
	}

	minioClient, err := NewMinioClient(cfg.Minio)
	if err != nil {
		panic(err)
	}

	pref := tele.Settings{
		Token:  cfg.Telegram.ApiToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle("/hello", func(c tele.Context) error {
		_, err = c.Bot().Send(&tele.User{ID: 910754150}, "zhopa", &tele.SendOptions{ReplyTo: &tele.Message{ID: 203}})
		if err != nil {
			return err
		}
		return c.Send("Hello!")
	})

	b.Handle(tele.OnVoice, func(c tele.Context) error {
		err := c.Reply("Received voice message to transcribe.")
		if err != nil {
			return err
		}

		voice := c.Message().Voice
		rc, err := b.File(voice.MediaFile())
		if err != nil {
			return err
		}
		defer rc.Close()

		minioCtx, cancel := context.WithTimeout(context.Background(), cfg.Minio.UploadTimeout)
		defer cancel()
		uploadInfo, err := minioClient.PutObject(minioCtx, cfg.Minio.BucketName, voice.UniqueID, rc, voice.FileSize, minio.PutObjectOptions{})
		if err != nil {
			return c.Reply(err)
		}

		amqpCtx, cancel := context.WithTimeout(context.Background(), cfg.AMQP.SendTimeout)
		defer cancel()

		body, err := json.Marshal(map[string]any{
			"telegram-user-id": c.Sender().ID,
			"telegram-msg-id":  c.Message().ID,
			"voice-unique-id":  voice.UniqueID,
		})
		if err != nil {
			return err
		}
		err = ch.PublishWithContext(amqpCtx,
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate

			amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			})
		if err != nil {
			return err
		}
		return c.Reply(fmt.Sprintf("Downloaded voice message: %d bytes, unique ID: %s. It will be transcribed later.", uploadInfo.Size, voice.UniqueID))
	})

	b.Start()
}

func NewMinioClient(cfg config.MinioConfig) (*minio.Client, error) {
	ctx := context.Background()

	// Initialize minio client object.
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: false,
		Region: cfg.Region,
	})
	if err != nil {
		log.Fatalln(err)
	}

	// Create bucket if not exist
	err = minioClient.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{
		Region: cfg.Region,
	})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, cfg.BucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", cfg.BucketName)
			return minioClient, nil
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s\n", cfg.BucketName)
	}

	return minioClient, err
}

func NewAMQPConn(cfg config.AMQPConfig) (*amqp.Connection, error) {
	return amqp.Dial(cfg.URL)
}
