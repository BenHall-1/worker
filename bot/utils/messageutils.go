package utils

import (
	"fmt"
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/rest"
	"time"
)

type SentMessage struct {
	Worker  *worker.Context
	Message *message.Message
}

// guildId is only used to get the language
func SendEmbed(worker *worker.Context, channelId, guildId uint64, colour Colour, title string, messageType translations.MessageId, fields []embed.EmbedField, deleteAfter int, isPremium bool, format ...interface{}) {
	content := fmt.Sprintf(i18n.GetMessageFromGuild(guildId, messageType), format...)
	_, _ = SendEmbedWithResponse(worker, channelId, colour, title, content, fields, deleteAfter, isPremium)
}

func SendEmbedRaw(worker *worker.Context, channel uint64, colour Colour, title, content string, fields []embed.EmbedField, deleteAfter int, isPremium bool) {
	_, _ = SendEmbedWithResponse(worker, channel, colour, title, content, fields, deleteAfter, isPremium)
}

func SendEmbedWithResponse(worker *worker.Context, channel uint64, colour Colour, title, content string, fields []embed.EmbedField, deleteAfter int, isPremium bool) (message.Message, error) {
	msgEmbed := embed.NewEmbed().
		SetColor(int(colour)).
		SetTitle(title).
		SetDescription(content)

	for _, field := range fields {
		msgEmbed.AddField(field.Name, field.Value, field.Inline)
	}

	if !isPremium {
		self, _ := worker.Self()
		msgEmbed.SetFooter("Powered by ticketsbot.net", self.AvatarUrl(256))
	}

	// Explicitly ignore error because it's usually a 403 (missing permissions)
	msg, err := worker.CreateMessageComplex(channel, rest.CreateMessageData{
		Embed: msgEmbed,
	})

	if err != nil {
		sentry.LogWithContext(err, errorcontext.WorkerErrorContext{
			Channel: channel,
		})

		return msg, err
	}

	if deleteAfter > 0 {
		DeleteAfter(SentMessage{worker, &msg}, deleteAfter)
	}

	return msg, err
}

func DeleteAfter(msg SentMessage, secs int) {
	go func() {
		time.Sleep(time.Duration(secs) * time.Second)

		// Fix a panic
		if msg.Message != nil && msg.Worker != nil {
			// Explicitly ignore error, pretty much always a 404
			_ = msg.Worker.DeleteMessage(msg.Message.ChannelId, msg.Message.Id)
		}
	}()
}

func ReactWithCheck(worker *worker.Context, channelId, messageId uint64) {
	if err := worker.CreateReaction(channelId, messageId, "✅"); err != nil {
		sentry.LogWithContext(err, errorcontext.WorkerErrorContext{
			Channel: channelId,
		})
	}
}

func ReactWithCross(worker *worker.Context, channelId, messageId uint64) {
	if err := worker.CreateReaction(channelId, messageId, "❌"); err != nil {
		sentry.LogWithContext(err, errorcontext.WorkerErrorContext{
			Channel: channelId,
		})
	}
}

func PadDiscriminator(discrim uint16) string {
	return fmt.Sprintf("%04d", discrim)
}
