package bot

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/moo-d/AwaraBot/internal/scraper"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func (b *Bot) startSTDINListener() {
	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				b.Log.Warnf("EOF received, stopping STDIN listener")
				return
			}
			b.Log.Errorf("Read error: %v", err)
			continue
		}

		msg := strings.TrimSpace(line)
		if msg == "" {
			continue
		}

		if strings.Contains(msg, "MESSAGE_END") {
			parts := strings.SplitN(msg, "MESSAGE_END", 2)
			b.processMessage(parts[0])

			if len(parts) > 1 && parts[1] != "" {
				msg = parts[1]
				continue
			}
		}
	}
}

func (b *Bot) processMessage(msg string) {
	switch {
	case strings.HasPrefix(msg, "DOWNLOAD_MEDIA:"):
		b.handleDownloadMedia(msg)
	case strings.HasPrefix(msg, "ENHANCE:"):
		b.handleEnhance(msg)
	case strings.HasPrefix(msg, "CHATBOT:"):
		b.handleChatbot(msg)
	case strings.HasPrefix(msg, "DOWNLOAD:"):
		b.handleDownloadCommand(msg)
	case strings.HasPrefix(msg, "SEND:"):
		b.handleSendMessage(msg)
	case strings.HasPrefix(msg, "REACT:"):
		b.handleReaction(msg)
	case strings.HasPrefix(msg, "SEND_URL_IMAGE:"), strings.HasPrefix(msg, "SEND_IMAGE:"):
		b.processMediaCommand(msg, "SEND_URL_IMAGE:", MediaImage)
		b.processMediaCommand(msg, "SEND_IMAGE:", MediaImage)
	case strings.HasPrefix(msg, "SEND_URL_VIDEO:"), strings.HasPrefix(msg, "SEND_VIDEO:"):
		b.processMediaCommand(msg, "SEND_URL_VIDEO:", MediaVideo)
		b.processMediaCommand(msg, "SEND_VIDEO:", MediaVideo)
	case strings.HasPrefix(msg, "SEND_URL_AUDIO:"), strings.HasPrefix(msg, "SEND_AUDIO:"):
		b.processMediaCommand(msg, "SEND_URL_AUDIO:", MediaAudio)
		b.processMediaCommand(msg, "SEND_AUDIO:", MediaAudio)
	}
}

func (b *Bot) handleDownloadMedia(msg string) {
	parts := strings.SplitN(strings.TrimPrefix(msg, "DOWNLOAD_MEDIA:"), "|", 3)
	if len(parts) < 3 {
		b.Log.Errorf("Invalid DOWNLOAD_MEDIA format")
		return
	}

	messageID := parts[0]
	chatJID := parts[1]
	contextType := parts[2]
	messageID = strings.TrimSuffix(messageID, "MESSAGE_END")

	go func() {
		chat, err := types.ParseJID(chatJID)
		if err != nil {
			b.Log.Errorf("Invalid chat JID: %v", err)
			fmt.Println("MEDIA_DATA:errorMESSAGE_END")
			return
		}

		var msg *waE2E.Message

		switch contextType {
		case "direct":
			msg = &waE2E.Message{
				ImageMessage: &waE2E.ImageMessage{
					URL:           proto.String(messageID),
					Mimetype:      proto.String("image/jpeg"),
					FileSHA256:    []byte("dummy-sha256"),
					FileEncSHA256: []byte("dummy-enc-sha256"),
					FileLength:    proto.Uint64(1),
				},
			}
		case "quoted":
			msg = &waE2E.Message{
				ExtendedTextMessage: &waE2E.ExtendedTextMessage{
					ContextInfo: &waE2E.ContextInfo{
						StanzaID:    proto.String(messageID),
						Participant: proto.String(chat.String()),
						QuotedMessage: &waE2E.Message{
							ImageMessage: &waE2E.ImageMessage{
								URL:           proto.String(messageID),
								Mimetype:      proto.String("image/jpeg"),
								FileSHA256:    []byte("dummy-sha256"),
								FileEncSHA256: []byte("dummy-enc-sha256"),
								FileLength:    proto.Uint64(1),
							},
						},
					},
				},
			}
		default:
			b.Log.Errorf("Unknown context type: %s", contextType)
			fmt.Println("MEDIA_DATA:errorMESSAGE_END")
			return
		}

		data, err := b.Client.DownloadAny(msg)
		if err != nil {
			b.Log.Errorf("Download error: %v", err)
			fmt.Println("MEDIA_DATA:errorMESSAGE_END")
			return
		}

		fmt.Printf("MEDIA_DATA:%sMESSAGE_END\n", base64.StdEncoding.EncodeToString(data))
	}()
}

func (b *Bot) handleEnhance(msg string) {
	parts := strings.SplitN(msg[len("ENHANCE:"):], "|", 3)
	if len(parts) < 3 {
		b.Log.Errorf("Invalid enhance format")
		return
	}

	action := parts[0]
	imageData := parts[1]
	isUrl := parts[2] == "1"

	go b.handleEnhanceRequest(action, imageData, isUrl)
}

func (b *Bot) handleChatbot(msg string) {
	parts := strings.SplitN(msg[len("CHATBOT:"):], "|", 4)
	if len(parts) < 4 {
		b.Log.Errorf("Invalid CHATBOT format")
		return
	}

	jid, err := types.ParseJID(parts[0])
	if err != nil {
		b.Log.Errorf("Failed to parse JID: %v", err)
		return
	}

	prompt := parts[1]
	model := parts[2]

	var messages []scraper.Message
	if err := json.Unmarshal([]byte(parts[3]), &messages); err != nil {
		b.Log.Errorf("Failed to unmarshal messages: %v", err)
		return
	}

	msgEvent := &events.Message{
		Info: types.MessageInfo{
			MessageSource: types.MessageSource{
				Chat:    jid,
				Sender:  jid,
				IsGroup: jid.Server == types.GroupServer,
			},
			ID:       types.MessageID("cli-" + time.Now().Format("20060102-150405")),
			PushName: "User",
		},
	}

	go b.handleGPTRequest(msgEvent, jid, prompt, model, messages)
}

func (b *Bot) handleDownloadCommand(msg string) {
	parts := strings.SplitN(msg[len("DOWNLOAD:"):], "|", 3)
	if len(parts) < 2 {
		b.Log.Errorf("Invalid download format")
		return
	}

	go b.handleDownload(parts[0], parts[1], parts[2])
}

func (b *Bot) handleSendMessage(msg string) {
	content := strings.SplitN(msg[5:], "|", 2)
	if len(content) != 2 {
		return
	}

	message := strings.ReplaceAll(content[1], "{{NL}}", "\n")
	jid, err := types.ParseJID(content[0])
	if err != nil {
		b.Log.Errorf("JID parse error: %v", err)
		return
	}

	_, err = b.Client.SendMessage(context.Background(), jid, &waProto.Message{
		Conversation: proto.String(message),
	})
	if err != nil {
		b.Log.Errorf("Send error: %v", err)
	}
}

func (b *Bot) handleReaction(msg string) {
	parts := strings.SplitN(msg[len("REACT:"):], "|", 4)
	if len(parts) < 4 {
		b.Log.Errorf("Invalid reaction format")
		return
	}

	jid, err := types.ParseJID(parts[0])
	if err != nil {
		b.Log.Errorf("Invalid JID: %v", err)
		return
	}

	messageID := parts[1]
	emoji := parts[2]
	senderjid, err := types.ParseJID(parts[3])
	if err != nil {
		b.Log.Errorf("Invalid JID: %v", err)
		return
	}

	_, err = b.Client.SendMessage(context.Background(), jid, b.Client.BuildReaction(jid, senderjid, messageID, emoji))
	if err != nil {
		b.Log.Errorf("Failed to send reaction: %v", err)
	}
}

func (b *Bot) handleDownload(service, url, format string) {
	var result interface{}
	var err error

	switch service {
	case "tiktok":
		result, err = b.TikTokScraper.DownloadVideo(url)
	case "youtube":
		if format == "mp3" {
			res, e := b.YouTubeScraper.Audio(url)
			if e == nil {
				result = map[string]interface{}{
					"status":    true,
					"url":       res.URL,
					"title":     res.Title,
					"duration":  res.Time * 60,
					"thumbnail": res.Thumbnail,
				}
			}
			err = e
		} else {
			res, e := b.YouTubeScraper.Video(url, "720")
			if e == nil {
				result = map[string]interface{}{
					"status":   true,
					"url":      res.URL,
					"title":    res.Title,
					"duration": res.Time * 60,
				}
			}
			err = e
		}
	}

	if err != nil {
		b.sendErrorResponse(err)
		return
	}

	b.sendSuccessResponse(result)
}

func (b *Bot) sendSuccessResponse(result interface{}) {
	response := map[string]interface{}{
		"type":   "download_result",
		"status": true,
		"result": result,
	}
	jsonResponse, _ := json.Marshal(response)
	fmt.Println("DOWNLOAD_RESULT:" + string(jsonResponse) + "MESSAGE_END")
	os.Stdout.Sync()
}

func (b *Bot) sendErrorResponse(err error) {
	response := map[string]interface{}{
		"type":   "download_result",
		"status": false,
		"error":  err.Error(),
	}
	jsonResponse, _ := json.Marshal(response)
	fmt.Println("DOWNLOAD_RESULT:" + string(jsonResponse) + "MESSAGE_END")
	os.Stdout.Sync()
}

func (b *Bot) handleGPTRequest(evt *events.Message, jid types.JID, prompt, model string, messages []scraper.Message) {
	if len(messages) == 0 || messages[0].Role != "system" {
		messages = append([]scraper.Message{
			{
				Role:    "system",
				Content: "Kamu adalah Alexa, asisten WhatsApp yang cerdas...",
			},
		}, messages...)
	}

	result, err := b.GPTScraper.Chat(prompt, messages, model)
	if err != nil {
		b.Log.Errorf("GPT error: %v", err)
		b.sendEvent(BotEvent{
			Type: "chatbot_error",
			Content: map[string]interface{}{
				"chat":  jid,
				"error": err.Error(),
			},
		})
		return
	}

	var jsonResponse struct {
		Cmd     string `json:"cmd"`
		Caption string `json:"caption"`
		Query   string `json:"query"`
	}

	b.sendEvent(BotEvent{
		Type: "chatbot_result",
		Content: map[string]interface{}{
			"chat":      evt.Info.Chat.String(),
			"from":      evt.Info.Sender.String(),
			"sender":    evt.Info.Sender.String(),
			"messageId": evt.Info.ID,
			"pushName":  evt.Info.PushName,
			"isGroup":   evt.Info.IsGroup,
			"message":   result.Message,
			"command":   jsonResponse.Cmd,
			"query":     jsonResponse.Query,
			"caption":   jsonResponse.Caption,
		},
	})
}

func (b *Bot) handleEnhanceRequest(action, imageData string, isUrl bool) {
	var imgBytes []byte
	var err error

	if isUrl {
		resp, err := http.Get(imageData)
		if err != nil {
			b.sendErrorResponse(err)
			return
		}
		defer resp.Body.Close()

		imgBytes, err = io.ReadAll(resp.Body)
	} else {
		imgBytes, err = base64.StdEncoding.DecodeString(imageData)
	}

	if err != nil {
		b.sendErrorResponse(err)
		return
	}

	enhanced, err := b.VyroScraper.EnhanceImage(imgBytes, action)
	if err != nil {
		b.sendErrorResponse(err)
		return
	}

	b.sendSuccessResponse(map[string]interface{}{
		"url": fmt.Sprintf("data:image/jpeg;base64,%s", base64.StdEncoding.EncodeToString(enhanced)),
	})
}
