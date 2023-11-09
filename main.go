package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"github.com/xeptore/wireuse/pkg/funcutils"
)

var (
	allowedUserIDs []string
	AppName        = ""
	AppVersion     = ""
	AppCompileTime = ""
)

const (
	NewsUrlEnvKey        = "NEWS_URL"
	AllowedUserIDsEnvKey = "ALLOWED_USER_IDS"
	BotTokenEnvKey       = "BOT_TOKEN"
	BotHTTPProxyURL      = "BOT_HTTP_PROXY_URL"
	ParseModeMarkdownV1  = models.ParseMode("Markdown")
	CommandNewsList      = "/newslist"
	CommandStart         = "/start"
	CLICommandBotName    = "bot"
)

/*
"title":"Retail Sales m\/m",
"country":"AUD",
"date":"2023-10-29T20:30:00-04:00",
"impact":"Medium","forecast":"0.3%",
"previous":"0.2%
*/
type NewsItem struct {
	Title          string    `json:"title"`
	Country        string    `json:"country"`
	Date           time.Time `json:"date"`
	CurrencyImpact string    `json:"impact"`
	Forecast       string    `json:"forecast"`
	Previous       string    `json:"previous"`
	Actual         string    `json:"actual"`
}

//go:embed profile.png
var botPNG []byte

func main() {
	compileTime, err := time.Parse(time.RFC3339, AppCompileTime)
	if nil != err {
		panic(err)
	}

	log := zerolog.New(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) { w.Out = os.Stderr; w.TimeFormat = time.RFC3339 })).With().Timestamp().Logger().Level(zerolog.TraceLevel)

	app := &cli.App{
		Name:           AppName,
		Version:        AppVersion,
		Compiled:       compileTime,
		Suggest:        true,
		Usage:          "Conisma Forex Helper Telegram Bot",
		DefaultCommand: CLICommandBotName,

		Commands: []*cli.Command{
			{
				Name:    CLICommandBotName,
				Usage:   "Starts bot server",
				Aliases: []string{"b"},
				Action:  buildBot(log),
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msg("command failed")
	}
}

func (h *Handler) handleStartCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	if !isFromAllowedUser(update.Message.From.ID) {
		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text: strings.Join(
				[]string{
					fmt.Sprintf("Your not Allowed User"),
				},
				"\n",
			),
			ParseMode: models.ParseModeMarkdown,
		}); nil != err {
			h.logger.Error().Err(err).Msg("failed to send start command success reply message")
		}
		return
	}

	if _, err := b.SetMyName(ctx, &bot.SetMyNameParams{
		Name: "Conisma Forex Helper",
	}); nil != err {
		h.logger.Error().Err(err).Msg("failed to set bot name")
	}

	if _, err := b.SetMyCommands(ctx, &bot.SetMyCommandsParams{
		Scope: &models.BotCommandScopeAllPrivateChats{},
		Commands: []models.BotCommand{
			{
				Command:     CommandStart,
				Description: "Restart Me",
			},
			{
				Command:     CommandNewsList,
				Description: "Today Important News List",
			},
		},
	}); nil != err {
		h.logger.Error().Err(err).Msg("failed to set bot commands")
	}

	if _, err := b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID:    update.Message.Chat.ID,
		Caption:   "_I wish I could also set this photo as my profile photo, but currently, Telegram does not allow me to do so_ 😥\n_However, you can do this going to @BotFather, sending /setuserpic command, selecting my username, then sharing this photo with @BotFather_ 🥲",
		Photo:     &models.InputFileUpload{Filename: "profile.png", Data: bytes.NewBuffer(botPNG)},
		ParseMode: models.ParseModeMarkdown,
	}); nil != err {
		h.logger.Error().Err(err).Msg("failed to send bot profile photo message")
	}

	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text: strings.Join(
			[]string{
				fmt.Sprintf("*%s*", AppName),
				fmt.Sprintf("Compiled At: `%s`", bot.EscapeMarkdown(AppCompileTime)),
				fmt.Sprintf("Version: `%s`", bot.EscapeMarkdown(AppVersion)),
			},
			"\n",
		),
		ParseMode: models.ParseModeMarkdown,
	}); nil != err {
		h.logger.Error().Err(err).Msg("failed to send start command success reply message")
	}
}

func (h *Handler) handleListCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	if !isFromAllowedUser(update.Message.From.ID) {
		return
	}

	chatID := update.Message.Chat.ID
	msgID := update.Message.ID
	fetchNewsItem(h.logger)

	args := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, CommandNewsList))
	pageNo, err := strconv.Atoi(args)
	if len(args) != 0 && nil != err || pageNo < 0 {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      fmt.Sprintf("Use `%s page_no` message format.", CommandNewsList),
			ParseMode: ParseModeMarkdownV1,
		})
		if nil != err {
			h.logger.Error().Err(err).Msg("failed to send command help message")
		}
		return
	} else if len(args) == 0 {
		pageNo = 0
	}

	loadingMessage, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:           chatID,
		Text:             "Listing news. Please wait...",
		ParseMode:        ParseModeMarkdownV1,
		ReplyToMessageID: msgID,
	})
	if nil != err {
		h.logger.Error().Err(err).Msg("failed to send loading reply message to the user")
	}

	if _, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{ChatID: chatID, MessageID: loadingMessage.ID}); nil != err {
		h.logger.Error().Err(err).Int("previousLoadingMessageId", loadingMessage.ID).Msg("failed to delete previous loading message")
	}
	currentTime := time.Now()
	for _, news := range newsData {

		if news.Date.Month() == currentTime.Month() && news.Date.Year() == currentTime.Year() && news.Date.Day() == currentTime.Day() {
			if news.CurrencyImpact == "High" || news.CurrencyImpact == "Medium" {
				if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text: strings.Join(
						[]string{
							fmt.Sprintf("*%s*", string(news.Title)),
							fmt.Sprintf("%s  %s  %v  %v", news.Country, news.CurrencyImpact, news.Previous, news.Forecast),
							fmt.Sprintf("%s", news.Date.Format("15:04:00")),
						},
						"\n",
					),
					ParseMode: models.ParseModeMarkdown,
				}); nil != err {
					h.logger.Error().Err(err).Msg("failed to send news item  success reply message")
				}
			}
		}

	}
}

func loadAllowedUserIDs(log zerolog.Logger) {
	v, ok := os.LookupEnv(AllowedUserIDsEnvKey)
	if !ok {
		log.Fatal().Str("key", AllowedUserIDsEnvKey).Msg("required environment variable is not set")
	}
	parts := strings.Split(v, ",")
	allowedUserIDs = funcutils.Map(parts, func(p string) string {
		return strings.TrimSpace(p)
	})
}

func isFromAllowedUser(uid int64) bool {
	userID := fmt.Sprintf("%d", uid)
	for _, v := range allowedUserIDs {
		if userID == v {
			return true
		}
	}

	return false
}

type Handler struct {
	logger zerolog.Logger
}

var newsData []NewsItem

func fetchNewsItem(log zerolog.Logger) {
	url, ok := os.LookupEnv(NewsUrlEnvKey)
	if !ok {
		log.Fatal().Str("key", NewsUrlEnvKey).Msg("required environment variable is not set")
	}

	client := resty.New()

	resp, err := client.R().EnableTrace().Get(url)
	if err != nil {
		log.Error().Err(err).Msg("failed to get url data")
	}
	if resp.StatusCode() == 200 {

		err := json.Unmarshal(resp.Body(), &newsData)
		if err != nil {
			log.Error().Err(err).Msg("failed to parse json data")
		}
	} else {
		log.Error().Int("StatusCode: ", resp.StatusCode()).Msg("failed to fetch data StatusCode is not 200")
	}
}
