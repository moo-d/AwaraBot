package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

type Bot struct {
	Client     *whatsmeow.Client
	Log        waLog.Logger
	retryCount int
}

type BotEvent struct {
	Type    string                 `json:"type"`
	Content map[string]interface{} `json:"content"`
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
}

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

func (b *Bot) sendEvent(event BotEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		b.Log.Errorf("Marshal error: %v", err)
		return
	}

	output := os.Stderr
	if event.Type == "message" {
		output = os.Stdout
	}
	fmt.Fprintln(output, string(data))
}

func (b *Bot) startSTDINListener() {
	reader := bufio.NewReader(os.Stdin)
	var buffer string

	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			b.Log.Errorf("Read error: %v", err)
			continue
		}

		buffer += text
		if !strings.Contains(buffer, "MESSAGE_END") {
			continue
		}

		parts := strings.SplitN(buffer, "MESSAGE_END", 2)
		msg := strings.TrimSpace(parts[0])
		buffer = parts[1]

		if !strings.HasPrefix(msg, "SEND:") {
			continue
		}

		content := strings.SplitN(msg[5:], "|", 2)
		if len(content) != 2 {
			continue
		}

		message := strings.ReplaceAll(content[1], "{{NL}}", "\n")
		jid, err := types.ParseJID(content[0])
		if err != nil {
			b.Log.Errorf("JID parse error: %v", err)
			continue
		}

		_, err = b.Client.SendMessage(context.Background(), jid, &waProto.Message{
			Conversation: proto.String(message),
		})
		if err != nil {
			b.Log.Errorf("Send error: %v", err)
		}
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

func (b *Bot) onConnected(evt *events.Connected) {
	b.retryCount = 0
	b.Log.Infof("Connected successfully")

	if b.Client.Store.PushName == "" {
		if name := os.Getenv("BOT_NAME"); name != "" {
			b.Client.Store.PushName = name
		} else {
			b.Client.Store.PushName = "Awara"
		}
	}

	go func() {
		time.Sleep(3 * time.Second)
		if err := b.Client.SendPresence(types.PresenceAvailable); err != nil {
			b.Log.Errorf("Presence error: %v", err)
		} else {
			b.Log.Infof("Presence set")
		}
	}()
}

func (b *Bot) onDisconnected() {
	b.Log.Warnf("Disconnected, reconnecting...")
	time.Sleep(5 * time.Second)

	if b.retryCount >= 5 {
		b.Log.Errorf("Max reconnect attempts")
		return
	}

	b.retryCount++
	if err := b.Client.Connect(); err != nil {
		b.Log.Errorf("Reconnect error: %v", err)
	}
}

func (b *Bot) connectWithQR() error {
	qrChan, _ := b.Client.GetQRChannel(context.Background())
	if err := b.Client.Connect(); err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}

	for evt := range qrChan {
		switch evt.Event {
		case "code":
			b.sendEvent(BotEvent{
				Type: "qr",
				Content: map[string]interface{}{
					"code":    evt.Code,
					"message": "Scan QR code with your phone",
				},
			})
		case "success":
			return nil
		case "timeout":
			return fmt.Errorf("QR expired")
		}
	}
	return nil
}

func (b *Bot) Run() {
	var err error
	if b.Client.Store.ID == nil {
		err = b.connectWithQR()
	} else {
		err = b.Client.Connect()
	}

	if err != nil {
		b.Log.Errorf("Connection failed: %v", err)
		return
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go b.startSTDINListener()
	<-sigCh
	b.Client.Disconnect()
}

func main() {
	waLog.Stdout("NETWORK", "DEBUG", true)
	waLog.Stdout("DATABASE", "INFO", true)

	if err := godotenv.Load(); err != nil {
		log.Printf("Env load warning: %v", err)
	}

	container, err := sqlstore.New("sqlite3", "file:bot.db?_foreign_keys=on&_journal_mode=WAL&_timeout=5000", nil)
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}

	device, err := container.GetFirstDevice()
	if err != nil {
		log.Fatalf("Device error: %v", err)
	}

	bot := &Bot{Log: waLog.Stdout("BOT", "INFO", true)}
	bot.initClient(device)

	log.Println("Starting bot...")
	bot.Run()
}
