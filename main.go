package main

import (
	"log/slog"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"

	"github.com/sysnote8main/simplebot/internal/wrapme"
)

var (
	messageLinkRegex = regexp.MustCompile(`https:\/\/(ptb\.|canary\.|)discord.com\/channels\/\d+\/\d+\/\d+`)
)

func main() {
	// slog.SetLogLoggerLevel(slog.LevelDebug)
	// Load token
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		slog.Error("Failed to load discord bot token. please check token was set.")
		os.Exit(1)
	}

	// Create bot instance
	client, err := discordgo.New("Bot " + token)
	if err != nil {
		slog.Error("Failed to create discordgo instance", wrapme.Error(err))
		os.Exit(1)
	}

	client.AddHandler(onMessageCreate)
	if err := client.Open(); err != nil {
		slog.Error("Failed to open client", wrapme.Error(err))
		os.Exit(1)
	}

	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	slog.Info("Bot on ready!", slog.String("BotUsername", client.State.User.Username))
	<-stopSignal

	client.Close() // TODO handle error
	slog.Info("Bye bye!")
}

func onMessageCreate(s *discordgo.Session, e *discordgo.MessageCreate) {
	if e.Author.Bot {
		return
	}

	// expand link
	if messageLinkRegex.MatchString(e.Content) {
		slog.Debug("Fired!", slog.String("username", e.Author.Username))
		_extractedLink := messageLinkRegex.FindString(e.Content)
		_splitArr := strings.Split(_extractedLink, "/")
		_l := len(_splitArr)
		guildId := _splitArr[_l-3]
		if e.GuildID == guildId {
			channelId := _splitArr[_l-2]
			ch, err := s.Channel(channelId)
			if err != nil {
				slog.Error("Failed to get channel", slog.String("channelId", channelId), wrapme.Error(err))
				return
			}
			if ch.NSFW {
				return
			}
			messageId := _splitArr[_l-1]
			msg, err := s.ChannelMessage(channelId, messageId)
			if err != nil {
				// TODO change this to stats
				slog.Error("Failed to get channel message", slog.String("channelId", channelId), slog.String("messageId", messageId), wrapme.Error(err))
				return
			}
			_embed := discordgo.MessageEmbed{
				Author: &discordgo.MessageEmbedAuthor{
					IconURL: msg.Author.AvatarURL(""),
					Name:    msg.Author.Username,
				},
				Description: msg.Content,
				Footer: &discordgo.MessageEmbedFooter{
					IconURL: e.Author.AvatarURL(""),
					Text:    "From " + e.Author.GlobalName + "'s message",
				},
			}
			_, err = s.ChannelMessageSendEmbedReply(e.ChannelID, &_embed, e.Reference())
			if err != nil {
				slog.Error("Failed to reply with embed", wrapme.Error(err))
			}
		} else {
			slog.Debug("Guild Id miss matched. skipping...")
		}
	}
}
