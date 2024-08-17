package window

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotkit/app/notify"
	"github.com/diamondburned/ningen/v3"
	"libdb.so/dissent/internal/gtkcord"
)

type loginWindow struct {
	*Window
	readyOnce sync.Once
}

func (w *loginWindow) Hook(state *gtkcord.State) {
	w.ctx = gtkcord.InjectState(w.ctx, state)

	w.Reconnecting()
	var reconnecting glib.SourceHandle

	// When the websocket closes, the screen must be changed to a busy one. The
	// websocket may close if it's disconnected unexpectedly.
	state.BindWidget(w, func(ev gateway.Event) {
		switch ev := ev.(type) {
		case *ningen.ConnectedEvent:
			slog.Info(
				"Discord gateway connected",
				"event", ev.EventType())

			// Cancel the 3s delay if we're already connected during that.
			if reconnecting != 0 {
				glib.SourceRemove(reconnecting)
				reconnecting = 0
			}

			w.Connected()

		case *ws.BackgroundErrorEvent:
			slog.Warn(
				"Discord gateway background error",
				"err", ev.Err)

		case *ws.CloseEvent:
			slog.Info(
				"Discord gateway closed",
				"err", ev.Err,
				"code", ev.Code)

		case *ningen.DisconnectedEvent:
			slog.Info(
				"Discord gateway disconnected",
				"err", ev.Err,
				"code", ev.Code)

			if ev.IsLoggedOut() {
				w.PromptLogin()
				return
			}

			// Add a 3s delay in case we have a sudden disruption that
			// immediately recovers.
			reconnecting = glib.TimeoutSecondsAdd(3, func() {
				w.Reconnecting()
				reconnecting = 0
			})

		case *gateway.ReadyEvent:
			if ev.UserSettings != nil {
				switch ev.UserSettings.Theme {
				case "dark":
					SetPreferDarkTheme(true)
				case "light":
					SetPreferDarkTheme(false)
				}
			}

		case *gateway.MessageCreateEvent:
			mentions := state.MessageMentions(&ev.Message)
			if mentions == 0 {
				return
			}

			if state.Status() == discord.DoNotDisturbStatus {
				return
			}

			avatarURL := gtkcord.InjectAvatarSize(ev.Author.AvatarURL())

			notify.Send(w.ctx, notify.Notification{
				ID: notify.HashID(ev.ChannelID),
				Title: fmt.Sprintf(
					"%s (%s)",
					state.AuthorDisplayName(ev),
					gtkcord.ChannelNameFromID(w.ctx, ev.ChannelID),
				),
				Body:  state.MessagePreview(&ev.Message),
				Icon:  notify.IconURL(w.ctx, avatarURL, notify.IconName("avatar-default-symbolic")),
				Sound: notify.MessageSound,
				Action: notify.Action{
					ActionID: "app.open-channel",
					Argument: gtkcord.NewChannelIDVariant(ev.ChannelID),
				},
			})
		}
	})
}

func (w *loginWindow) Ready(state *gtkcord.State) {
	app := w.Application()
	app.ConnectShutdown(func() {
		slog.Info("Closing Discord session...")

		if err := state.Close(); err != nil {
			slog.Error("error closing session", "err", err)
		}
	})
}

func (w *loginWindow) Reconnecting() {
	w.Stack.SetVisibleChild(w.Loading)
	w.SetTitle("Connecting")
}

func (w *loginWindow) Connected() {
	w.readyOnce.Do(func() {
		w.initChatPage()
		w.initActions()
	})
	w.Window.SwitchToChatPage()
}

func (w *loginWindow) PromptLogin() {
	w.Window.SwitchToLoginPage()
}
