package main

import (
	"context"
	"go-tg-transcriber/config"
	"golang.org/x/sync/errgroup"
	"log"
	"time"

	tele "gopkg.in/telebot.v3"
)

func main() {
	cfg := config.MustLoad()

	pref := tele.Settings{
		Token:  cfg.TelegramApiToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle("/hello", func(c tele.Context) error {
		return c.Send("Hello!")
	})

	b.Handle(tele.OnVoice, func(c tele.Context) error {
		g, _ := errgroup.WithContext(context.Background())

		for _, msg := range []string{".", "..", "...", "transcribed as ass"} {
			g.Go(func() error {
				return c.Send(msg)
			})
			time.Sleep(time.Second)
		}

		return g.Wait()
	})

	b.Start()
}
