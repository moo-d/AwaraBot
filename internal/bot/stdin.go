package bot

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
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
	case strings.HasPrefix(msg, "DOWNLOAD:"):
		parts := strings.SplitN(msg[len("DOWNLOAD:"):], "|", 3)
		if len(parts) < 2 {
			b.Log.Errorf("Invalid download format")
			return
		}

		go b.handleDownload(parts[0], parts[1], parts[2])

	case strings.HasPrefix(msg, "SEND:"):
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
