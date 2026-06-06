package app

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/danzt/daas/api/internal/domain/notification"
)

// MultiChannelNotifier fans out a single notification to multiple adapters
// concurrently. Each adapter runs in a separate goroutine; errors are logged
// but never returned — delivery is best-effort across all channels.
//
// Callers set msg.Channel to any value (or leave it empty). MultiChannelNotifier
// overrides the Channel field per-adapter before calling Send, so each adapter
// receives exactly its registered channel name.
type MultiChannelNotifier struct {
	adapters []channelAdapter
}

type channelAdapter struct {
	channel string
	svc     notification.NotificationService
}

// NewMultiChannelNotifier creates a notifier that fans out to the given
// adapters. Pairs are passed as alternating (service, channel) values:
//
//	NewMultiChannelNotifier(emailSvc, "email", whatsappSvc, "whatsapp")
//
// Non-conforming pairs (wrong type, missing channel) are silently skipped.
func NewMultiChannelNotifier(pairs ...any) *MultiChannelNotifier {
	n := &MultiChannelNotifier{}
	for i := 0; i+1 < len(pairs); i += 2 {
		svc, ok1 := pairs[i].(notification.NotificationService)
		ch, ok2 := pairs[i+1].(string)
		if ok1 && ok2 {
			n.adapters = append(n.adapters, channelAdapter{channel: ch, svc: svc})
		}
	}
	return n
}

// Send fans out msg to every registered adapter concurrently.
// Each adapter receives a copy of msg with Channel set to its registered
// channel name. All failures are logged as warnings — Send always returns nil.
func (n *MultiChannelNotifier) Send(ctx context.Context, msg notification.Message) error {
	var wg sync.WaitGroup
	for _, a := range n.adapters {
		wg.Add(1)
		go func(ca channelAdapter) {
			defer wg.Done()
			m := msg
			m.Channel = ca.channel
			if err := ca.svc.Send(ctx, m); err != nil {
				log.Warn().
					Err(err).
					Str("channel", ca.channel).
					Str("to", msg.To).
					Msg("multi-channel: delivery failed")
			}
		}(a)
	}
	wg.Wait()
	return nil
}
