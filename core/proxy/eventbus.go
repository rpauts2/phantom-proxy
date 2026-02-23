// Package proxy - Event Bus
package proxy

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// EventBus шина событий
type EventBus struct {
	mu       sync.RWMutex
	redis    *redis.Client
	logger   *zap.Logger
	channel  string
	handlers map[string][]EventHandler
}

// EventHandler обработчик событий
type EventHandler func(eventType string, payload map[string]interface{})

// NewEventBus создает шину событий
func NewEventBus(rdb *redis.Client, logger *zap.Logger, channel string) *EventBus {
	eb := &EventBus{
		redis:    rdb,
		logger:   logger,
		channel:  channel,
		handlers: make(map[string][]EventHandler),
	}

	// Запустить подписку
	go eb.subscribe()

	return eb
}

// Publish публикует событие
func (eb *EventBus) Publish(eventType string, payload map[string]interface{}) {
	event := map[string]interface{}{
		"type":      eventType,
		"payload":   payload,
		"timestamp": time.Now().UnixNano(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		eb.logger.Error("Failed to marshal event", zap.Error(err))
		return
	}

	if err := eb.redis.Publish(context.Background(), eb.channel, string(data)).Err(); err != nil {
		eb.logger.Error("Failed to publish event", zap.Error(err))
	}

	eb.logger.Debug("Event published",
		zap.String("type", eventType),
		zap.Any("payload", payload),
	)
}

// Subscribe подписывается на событие
func (eb *EventBus) Subscribe(eventType string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
	eb.logger.Debug("Handler subscribed", zap.String("type", eventType))
}

// Unsubscribe отписывается от события
func (eb *EventBus) Unsubscribe(eventType string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	handlers := eb.handlers[eventType]
	for i, h := range handlers {
		if &h == &handler {
			eb.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// subscribe подписка на Redis Pub/Sub
func (eb *EventBus) subscribe() {
	pubsub := eb.redis.Subscribe(context.Background(), eb.channel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	eb.logger.Info("Event bus subscribed", zap.String("channel", eb.channel))

	for msg := range ch {
		eb.handleMessage(msg.Payload)
	}
}

// handleMessage обрабатывает сообщение
func (eb *EventBus) handleMessage(payload string) {
	var event map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		eb.logger.Error("Failed to unmarshal event", zap.Error(err))
		return
	}

	eventType, ok := event["type"].(string)
	if !ok {
		return
	}

	eventPayload, ok := event["payload"].(map[string]interface{})
	if !ok {
		return
	}

	eb.mu.RLock()
	handlers := eb.handlers[eventType]
	eb.mu.RUnlock()

	for _, handler := range handlers {
		go handler(eventType, eventPayload)
	}
}

// Close закрывает шину событий
func (eb *EventBus) Close() error {
	return eb.redis.Close()
}

// GetStats возвращает статистику
func (eb *EventBus) GetStats() map[string]interface{} {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	totalHandlers := 0
	for _, handlers := range eb.handlers {
		totalHandlers += len(handlers)
	}

	return map[string]interface{}{
		"channel":        eb.channel,
		"total_handlers": totalHandlers,
		"event_types":    len(eb.handlers),
	}
}
