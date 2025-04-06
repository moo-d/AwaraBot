package bot

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/moo-d/AwaraBot/internal/scraper"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

type Bot struct {
	Client         *whatsmeow.Client
	Log            waLog.Logger
	retryCount     int
	TikTokScraper  *scraper.TikTokScraper
	YouTubeScraper *scraper.YouTubeScraper
}

type BotEvent struct {
	Type    string                 `json:"type"`
	Content map[string]interface{} `json:"content"`
}

func NewBot(device *store.Device, logger waLog.Logger) *Bot {
	b := &Bot{
		Log:           logger,
		TikTokScraper: scraper.NewTikTokScraper(),
	}
	b.initClient(device)
	return b
}

func (b *Bot) initClient(device *store.Device) {
	store.DeviceProps = &waProto.DeviceProps{
		Os:              proto.String("WhatsApp Bot"),
		PlatformType:    waProto.DeviceProps_DESKTOP.Enum(),
		RequireFullSync: proto.Bool(true),
	}

	b.Client = whatsmeow.NewClient(device, b.Log)
	b.Client.EnableAutoReconnect = true
	b.Client.AutoReconnectErrors = 3
	device.PushName = os.Getenv("BOT_NAME")
	b.Client.AddEventHandler(b.eventHandler)
	b.YouTubeScraper = scraper.NewYouTubeScraper()
}

func (b *Bot) sendEvent(event BotEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		b.Log.Errorf("Marshal error: %v", err)
		return
	}
	fmt.Fprintln(os.Stdout, string(data))
}
