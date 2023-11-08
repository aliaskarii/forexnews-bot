package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
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
	Title          string `json:"title"`
	Country        string `json:"country"`
	Date           string `json:"date"`
	CurrencyImpact string `json:"impact"`
	Forecast       string `json:"forecast"`
	Previous       string `json:"previous"`
	Actual         string `json:"actual"`
}

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
				Description: "News List",
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

	// for _, news := range newsData {
	// 	message := "Top news by Currency Impact:\n\n"
	// 	message += fmt.Sprintf("Event: %s\n", news.Title)
	// 	message += fmt.Sprintf("Data: %s\n", news.Date)
	// 	message += fmt.Sprintf("Currency Impact: %s\n", news.CurrencyImpact)
	// 	message += fmt.Sprintf("Actual: %s\n", news.Actual)
	// 	message += fmt.Sprintf("Forecast: %s\n", news.Forecast)
	// 	message += fmt.Sprintf("Previous: %s\n\n", news.Previous)
	// 	msg := tgbotapi.NewMessage(chatID, message)

	// 	if _, err := bot.Send(msg); err != nil {
	// 		fmt.Printf("Failed to send message: %v\n", err)
	// 	}

	// }

	loadingMessage, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:           chatID,
		Text:             "Listing news. Please wait...",
		ParseMode:        ParseModeMarkdownV1,
		ReplyToMessageID: msgID,
	})
	if nil != err {
		h.logger.Error().Err(err).Msg("failed to send loading reply message to the user")
	}

	//list, err := h.listnews(ctx, pageNo)
	if _, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{ChatID: chatID, MessageID: loadingMessage.ID}); nil != err {
		h.logger.Error().Err(err).Int("previousLoadingMessageId", loadingMessage.ID).Msg("failed to delete previous loading message")
	}
	if nil != err {
		h.logger.Error().Err(err).Int64("chatID", chatID).Int("pageNo", pageNo).Msg("failed to get list of news")
		return
	}
	msg := bot.SendMessageParams{
		ChatID:    chatID,
		ParseMode: ParseModeMarkdownV1,
	}
	//msg
	if _, err := b.SendMessage(ctx, &msg); nil != err {
		h.logger.Error().Err(err).Msg("failed to send reply message")
		return
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

func fetch_news_data() {
	url := "https://nfs.faireconomy.media/ff_calendar_thisweek.json"

	client := resty.New()

	resp, err := client.R().EnableTrace().Get(url)
	if err != nil {
		fmt.Printf("Failed to fetch data: %v\n", err)
		return
	}
	if resp.StatusCode() == 200 {
		var newsData []NewsItem

		err := json.Unmarshal(resp.Body(), &newsData)
		if err != nil {
			fmt.Printf("Failed to parse JSON: %v\n", err)
			return
		}

		sort.SliceStable(newsData, func(i, j int) bool {
			return newsData[i].CurrencyImpact > newsData[j].CurrencyImpact
		})

		for _, news := range newsData {
			fmt.Printf("Event: %s\n", news.Title)
			fmt.Printf("Currency Impact: %s\n", news.CurrencyImpact)
			fmt.Printf("Date: %s\n", news.Date)
			fmt.Printf("Actual: %s\n", news.Actual)
			fmt.Printf("Forecast: %s\n", news.Forecast)
			fmt.Printf("Previous: %s\n\n", news.Previous)
		}
	} else {
		fmt.Printf("Failed to fetch data. Status code: %d\n", resp.StatusCode())
	}
}
