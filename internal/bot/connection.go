package bot

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

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
