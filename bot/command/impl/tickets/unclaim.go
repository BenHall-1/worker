package tickets

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/rest"
)

type UnclaimCommand struct {
}

func (UnclaimCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "unclaim",
		Description:     translations.HelpUnclaim,
		PermissionLevel: permission.Support,
		Category:        command.Tickets,
	}
}

func (UnclaimCommand) Execute(ctx command.CommandContext) {
	// Get ticket struct
	ticket, err := dbclient.Client.Tickets.GetByChannel(ctx.ChannelId); if err != nil {
		ctx.HandleError(err)
		return
	}

	// Verify this is a ticket channel
	if ticket.UserId == 0 {
		ctx.SendEmbed(utils.Red, "Error", translations.MessageNotATicketChannel)
		ctx.ReactWithCross()
		return
	}

	// Get who claimed
	whoClaimed, err := dbclient.Client.TicketClaims.Get(ctx.GuildId, ticket.Id); if err != nil {
		ctx.HandleError(err)
		return
	}

	if whoClaimed == 0 {
		ctx.SendEmbed(utils.Red, "Error", translations.MessageNotClaimed)
		ctx.ReactWithCross()
		return
	}

	if ctx.UserPermissionLevel < permission.Admin && ctx.Author.Id != whoClaimed {
		ctx.SendEmbed(utils.Red, "Error", translations.MessageOnlyClaimerCanUnclaim)
		ctx.ReactWithCross()
		return
	}

	// Set to unclaimed in DB
	if err := dbclient.Client.TicketClaims.Delete(ctx.GuildId, ticket.Id); err != nil {
		ctx.HandleError(err)
		return
	}

	// Update channel
	data := rest.ModifyChannelData{
		PermissionOverwrites: logic.CreateOverwrites(ctx.GuildId, ticket.UserId, ctx.Worker.BotId),
	}
	if _, err := ctx.Worker.ModifyChannel(ctx.ChannelId, data); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.SendEmbed(utils.Green, "Ticket Unclaimed", translations.MessageUnclaimed)
	ctx.ReactWithCheck()
}
