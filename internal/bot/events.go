package bot

import (
	waProto "go.mau.fi/whatsmeow/binary/proto"
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

	isImage := msg.Message.GetImageMessage() != nil
	var quotedMsg *waProto.Message
	var isQuotedImage bool
	var quotedMsgID string
	var quotedParticipant string

	if extMsg := msg.Message.GetExtendedTextMessage(); extMsg != nil {
		if ctxInfo := extMsg.GetContextInfo(); ctxInfo != nil {
			quotedMsg = ctxInfo.GetQuotedMessage()
			quotedMsgID = ctxInfo.GetStanzaId()
			quotedParticipant = ctxInfo.GetParticipant()
			isQuotedImage = quotedMsg.GetImageMessage() != nil
		}
	}

	eventContent := map[string]interface{}{
		"from":          msg.Info.Sender.String(),
		"chat":          msg.Info.Chat.String(),
		"text":          text,
		"pushName":      msg.Info.PushName,
		"isGroup":       msg.Info.IsGroup,
		"messageId":     msg.Info.ID,
		"isImage":       isImage,
		"isQuotedImage": isQuotedImage,
	}

	if isQuotedImage {
		eventContent["quotedMessage"] = map[string]interface{}{
			"messageId": quotedMsgID,
			"from":      quotedParticipant,
			"isImage":   true,
		}
	}

	b.sendEvent(BotEvent{
		Type:    "message",
		Content: eventContent,
	})
}
