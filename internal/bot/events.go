package bot

import (
	"go.mau.fi/whatsmeow/types/events"
)

func (b *Bot) eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		b.handleMessage(v)
	case *events.Connected:
		b.onConnected(v)
	case *events.Disconnected:
		b.onDisconnected()
	case *events.HistorySync:
		b.Log.Infof("History sync: %d conversations", len(v.Data.GetConversations()))
	}
}

func (b *Bot) handleMessage(msg *events.Message) {
	if msg.Info.IsFromMe {
		return
	}

	text := ""
	if conv := msg.Message.GetConversation(); conv != "" {
		text = conv
	} else if ext := msg.Message.GetExtendedTextMessage(); ext != nil {
		text = ext.GetText()
	}

	b.sendEvent(BotEvent{
		Type: "message",
		Content: map[string]interface{}{
			"from":     msg.Info.Sender.String(),
			"chat":     msg.Info.Chat.String(),
			"text":     text,
			"pushName": msg.Info.PushName,
			"isGroup":  msg.Info.IsGroup,
		},
	})
}
