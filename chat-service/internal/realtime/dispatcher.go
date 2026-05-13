package realtime

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
)

type Dispatcher struct {
	hub        *Hub
	fanout     Fanout
	offline    OfflineQueue
	instanceID string
}

type Fanout interface {
	Publish(ctx context.Context, channel string, payload []byte) error
	Subscribe(ctx context.Context, channel string, handler func(payload []byte)) (stop func(), err error)
}

type OfflineQueue interface {
	Enqueue(ctx context.Context, userID string, payload []byte) error
	Drain(ctx context.Context, userID string, handler func(payload []byte)) error
}

type FanoutEnvelope struct {
	Source  string          `json:"source"`
	UserIDs []string        `json:"userIds"`
	Payload json.RawMessage `json:"payload"`
}

func NewDispatcher(hub *Hub, instanceID string, fanout Fanout, offline OfflineQueue) *Dispatcher {
	if instanceID == "" {
		instanceID = randomHex(8)
	}
	return &Dispatcher{hub: hub, fanout: fanout, offline: offline, instanceID: instanceID}
}

func (d *Dispatcher) Start(ctx context.Context) {
	if d.fanout == nil {
		return
	}

	stop, err := d.fanout.Subscribe(ctx, "chat:fanout", func(payload []byte) {
		var env FanoutEnvelope
		if err := json.Unmarshal(payload, &env); err != nil {
			return
		}
		if env.Source == d.instanceID {
			return
		}

		for _, userID := range env.UserIDs {
			d.hub.SendRawToUser(userID, env.Payload)
		}
	})
	if err != nil {
		log.Printf("redis fanout subscribe error: %v", err)
		return
	}

	go func() {
		<-ctx.Done()
		stop()
	}()
}

func (d *Dispatcher) PublishToUsers(ctx context.Context, userIDs []string, payload any) {
	if len(userIDs) == 0 {
		return
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	for _, userID := range userIDs {
		if d.hub.IsUserConnected(userID) {
			d.hub.SendRawToUser(userID, data)
		} else if d.offline != nil {
			_ = d.offline.Enqueue(ctx, userID, data)
		}
	}

	if d.fanout != nil {
		envBytes, err := json.Marshal(FanoutEnvelope{Source: d.instanceID, UserIDs: userIDs, Payload: data})
		if err == nil {
			_ = d.fanout.Publish(ctx, "chat:fanout", envBytes)
		}
	}
}

func (d *Dispatcher) PublishTransientToUsers(ctx context.Context, userIDs []string, payload any) {
	if len(userIDs) == 0 {
		return
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	for _, userID := range userIDs {
		if d.hub.IsUserConnected(userID) {
			d.hub.SendRawToUser(userID, data)
		}
	}

	if d.fanout != nil {
		envBytes, err := json.Marshal(FanoutEnvelope{Source: d.instanceID, UserIDs: userIDs, Payload: data})
		if err == nil {
			_ = d.fanout.Publish(ctx, "chat:fanout", envBytes)
		}
	}
}

func (d *Dispatcher) DrainOffline(ctx context.Context, userID string) {
	if d.offline == nil {
		return
	}
	_ = d.offline.Drain(ctx, userID, func(payload []byte) {
		d.hub.SendRawToUser(userID, payload)
	})
}

func randomHex(nBytes int) string {
	b := make([]byte, nBytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
