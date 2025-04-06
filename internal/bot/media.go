package bot

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

type MediaType string

const (
	MediaImage MediaType = "image"
	MediaVideo MediaType = "video"
	MediaAudio MediaType = "audio"
)

func getAudioDuration(path string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("ffprobe error: %v, output: %s", err, string(output))
	}

	durationStr := strings.TrimSpace(string(output))
	return strconv.ParseFloat(durationStr, 64)
}

func (b *Bot) getAudioDuration(audioData []byte) (float64, error) {
	tmpFile, err := os.CreateTemp("", "whatsapp_audio_*.tmp")
	if err != nil {
		return 0, fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(audioData); err != nil {
		return 0, fmt.Errorf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	return getAudioDuration(tmpFile.Name())
}

func (b *Bot) uploadAndSendMedia(jid types.JID, mediaData []byte, mediaType MediaType, caption string) error {
	var waMediaType whatsmeow.MediaType
	var msg *waProto.Message

	switch mediaType {
	case MediaImage:
		waMediaType = whatsmeow.MediaImage
		msg = &waProto.Message{
			ImageMessage: &waProto.ImageMessage{
				Caption: proto.String(caption),
			},
		}
	case MediaVideo:
		waMediaType = whatsmeow.MediaVideo
		msg = &waProto.Message{
			VideoMessage: &waProto.VideoMessage{
				Caption: proto.String(caption),
			},
		}
	case MediaAudio:
		waMediaType = whatsmeow.MediaAudio
		msg = &waProto.Message{
			AudioMessage: &waProto.AudioMessage{},
		}
	default:
		return fmt.Errorf("unsupported media type")
	}

	uploaded, err := b.Client.Upload(context.Background(), mediaData, waMediaType)
	if err != nil {
		return fmt.Errorf("upload failed: %v", err)
	}

	switch mediaType {
	case MediaImage:
		imgMsg := msg.ImageMessage
		imgMsg.Mimetype = proto.String(http.DetectContentType(mediaData))
		imgMsg.URL = proto.String(uploaded.URL)
		imgMsg.DirectPath = proto.String(uploaded.DirectPath)
		imgMsg.MediaKey = uploaded.MediaKey
		imgMsg.FileEncSHA256 = uploaded.FileEncSHA256
		imgMsg.FileSHA256 = uploaded.FileSHA256
		imgMsg.FileLength = proto.Uint64(uint64(len(mediaData)))
	case MediaVideo:
		vidMsg := msg.VideoMessage
		vidMsg.Mimetype = proto.String(http.DetectContentType(mediaData))
		vidMsg.URL = proto.String(uploaded.URL)
		vidMsg.DirectPath = proto.String(uploaded.DirectPath)
		vidMsg.MediaKey = uploaded.MediaKey
		vidMsg.FileEncSHA256 = uploaded.FileEncSHA256
		vidMsg.FileSHA256 = uploaded.FileSHA256
		vidMsg.FileLength = proto.Uint64(uint64(len(mediaData)))
	case MediaAudio:
		audioMsg := msg.AudioMessage
		audioMsg.Mimetype = proto.String("audio/mpeg")
		audioMsg.URL = proto.String(uploaded.URL)
		audioMsg.DirectPath = proto.String(uploaded.DirectPath)
		audioMsg.MediaKey = uploaded.MediaKey
		audioMsg.FileEncSHA256 = uploaded.FileEncSHA256
		audioMsg.FileSHA256 = uploaded.FileSHA256
		audioMsg.FileLength = proto.Uint64(uint64(len(mediaData)))

		if duration, err := b.getAudioDuration(mediaData); err == nil {
			audioMsg.Seconds = proto.Uint32(uint32(duration + 0.5))
		}
	}

	_, err = b.Client.SendMessage(context.Background(), jid, msg)
	return err
}

func (b *Bot) processMediaCommand(msg string, prefix string, mediaType MediaType) {
	content := strings.SplitN(strings.TrimPrefix(msg, prefix), "|", 3)
	if len(content) < 2 {
		b.Log.Errorf("Invalid %s message format", mediaType)
		return
	}

	jid, err := types.ParseJID(content[0])
	if err != nil {
		b.Log.Errorf("JID parse error: %v", err)
		return
	}

	var mediaData []byte
	var caption string

	if strings.HasPrefix(prefix, "SEND_URL_") {
		url := content[1]
		if !strings.HasPrefix(url, "http") {
			b.Log.Errorf("Invalid URL scheme: %s", url)
			return
		}

		resp, err := http.Get(url)
		if err != nil {
			b.Log.Errorf("Failed to download %s: %v", mediaType, err)
			return
		}
		defer resp.Body.Close()

		mediaData, err = io.ReadAll(resp.Body)
		if err != nil {
			b.Log.Errorf("Failed to read %s data: %v", mediaType, err)
			return
		}

		if len(content) > 2 {
			caption = strings.ReplaceAll(content[2], "{{NL}}", "\n")
		}
	} else {
		var err error
		mediaData, err = base64.StdEncoding.DecodeString(content[1])
		if err != nil {
			return
		}

		if len(content) > 2 {
			caption = strings.ReplaceAll(content[2], "{{NL}}", "\n")
		}
	}

	if err := b.uploadAndSendMedia(jid, mediaData, mediaType, caption); err != nil {
		b.Log.Errorf("%s send error: %v", strings.Title(string(mediaType)), err)
	}
}
