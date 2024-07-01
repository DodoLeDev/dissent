package channels

import (
	"log/slog"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotkit/gtkutil/cssutil"
)

var channelIndicatorCSS = cssutil.Applier("channel-item-indicator", `
	@define-color channel_item_indicator_color @theme_bg_color;

	.channel-item-indicator {
		font-family: monospace;
		font-size: 0.75em;
		font-weight: bold;

		margin-right: -12px;
		min-width: 12px; /* prevent (size >= 0) warnings */

		/* Replicate a text outline for the indicators. */
		text-shadow:
			0 0 2px @channel_item_indicator_color,
			0 0 2px @channel_item_indicator_color,
			0 0 2px @channel_item_indicator_color,
			0 0 2px @channel_item_indicator_color,
			0 0 2px @channel_item_indicator_color,
			0 0 2px @channel_item_indicator_color,
			0 0 2px @channel_item_indicator_color,
			0 0 2px @channel_item_indicator_color,
			0 0 2px @channel_item_indicator_color,
			0 0 2px @channel_item_indicator_color,
			0 0 2px @channel_item_indicator_color,
			0 0 2px @channel_item_indicator_color,
			0 0 2px @channel_item_indicator_color,
			0 0 2px @channel_item_indicator_color,
			0 0 2px @channel_item_indicator_color;
	}
`)

var channelIconImageCSS = cssutil.Applier("channel-icon-image", `
	.channel-icon-image {
		min-width: 0px;
                border-radius: 99px;
	}
`)

type ChannelIconOverrideFunc func(t discord.ChannelType) (string, bool)

var _ ChannelIconOverrideFunc = channelIconBase

func channelIconBase(chType discord.ChannelType) (string, bool) {
	switch chType {
	case discord.GuildText:
		return "channel-symbolic", true
	case discord.GuildAnnouncement:
		return "channel-broadcast-symbolic", true
	case discord.GuildPublicThread, discord.GuildPrivateThread, discord.GuildAnnouncementThread:
		return "thread-branch-symbolic", true
	case discord.GuildVoice, discord.GuildStageVoice:
		return "channel-voice-symbolic", true
	case discord.DirectMessage:
		return "person-symbolic", true
	case discord.GroupDM:
		return "group-symbolic", true
	default:
		return "channel-symbolic", false
	}
}

// ChannelIcon is a widget that displays a channel's icon.
// This does not handle DM channels, as it does not display avatars. It is only
// for displaying the channel icon.
type ChannelIcon struct {
	*gtk.Widget
	Icon *gtk.Image
}

const unknownChannelType discord.ChannelType = 9999

// NewChannelIcon creates a new ChannelIcon.
// If ch is nil, the icon will be a special unknown icon.
func NewChannelIcon(ch *discord.Channel, overrides ...ChannelIconOverrideFunc) *ChannelIcon {
	chType := unknownChannelType
	var nsfw bool
	if ch != nil {
		chType = ch.Type
		nsfw = ch.NSFW
	}

	var iconName string
	var found bool
	for _, override := range append(overrides, channelIconBase) {
		iconName, found = override(chType)
		if found {
			break
		}
	}

	if !found && chType != unknownChannelType {
		slog.Debug(
			"channel icon called with unknown channel type, using fallback icon",
			"channel_type", chType)
	}

	icon := gtk.NewImageFromIconName(iconName)
	channelIconImageCSS(icon)

	if found && !nsfw {
		return &ChannelIcon{
			Widget: &icon.Widget,
			Icon:   icon,
		}
	}

	var indicatorStr string
	if !found {
		indicatorStr = "?"
	} else {
		indicatorStr = "!"
	}

	indicator := gtk.NewLabel(indicatorStr)
	indicator.SetXAlign(1)
	indicator.SetHAlign(gtk.AlignCenter)
	indicator.SetVAlign(gtk.AlignEnd)
	indicator.SetHExpand(false)
	indicator.SetVExpand(false)
	channelIndicatorCSS(indicator)

	iconFrame := gtk.NewOverlay()
	iconFrame.SetHAlign(gtk.AlignCenter)
	iconFrame.SetVAlign(gtk.AlignCenter)
	iconFrame.SetChild(icon)
	iconFrame.AddOverlay(indicator)

	return &ChannelIcon{
		Widget: &iconFrame.Widget,
		Icon:   icon,
	}
}
