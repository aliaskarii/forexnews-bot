package main

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/go-telegram/bot"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

func buildBot(log zerolog.Logger) func(*cli.Context) error {
	return func(cliCtx *cli.Context) error {
		ctx, cancel := signal.NotifyContext(cliCtx.Context, os.Interrupt)
		defer cancel()

		if err := godotenv.Load(); nil != err {
			if !errors.Is(err, os.ErrNotExist) {
				log.Fatal().Err(err).Msg("unexpected error while loading .env file")
			}
			log.Warn().Msg(".env file not found")
		}
		loadAllowedUserIDs(log)
		handler := Handler{
			logger: log.With().Str("app", "handler").Logger(),
		}

		httpTransport := http.Transport{IdleConnTimeout: 10 * time.Second, ResponseHeaderTimeout: 30 * time.Second}
		httpClient := http.Client{Timeout: time.Second * 35, Transport: &httpTransport}
		proxyURL, ok := os.LookupEnv(BotHTTPProxyURL)
		if ok {
			httpProxyURL, err := url.Parse(proxyURL)
			if nil != err {
				log.Fatal().Err(err).Msg("failed to parse bot http proxy url")
			}
			httpTransport.Proxy = http.ProxyURL(httpProxyURL)
		}

		opts := []bot.Option{
			bot.WithCheckInitTimeout(30 * time.Second),
			bot.WithHTTPClient(30*time.Second, &httpClient),
		}
		token, ok := os.LookupEnv(BotTokenEnvKey)
		if !ok {
			log.Fatal().Str("key", BotTokenEnvKey).Msg("required environment variable is not set")
		}
		b, err := bot.New(token, opts...)
		if nil != err {
			log.Fatal().Err(err).Msg("failed to initialize bot instance")
		}
		b.RegisterHandler(bot.HandlerTypeMessageText, CommandStart, bot.MatchTypeExact, handler.handleStartCommand)
		b.RegisterHandler(bot.HandlerTypeMessageText, CommandNewsList, bot.MatchTypePrefix, handler.handleListCommand)
		b.Start(ctx)

		return nil
	}
}
