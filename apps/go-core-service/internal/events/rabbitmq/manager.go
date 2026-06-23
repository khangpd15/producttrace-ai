package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	ErrClosed       = errors.New("rabbitmq manager is closed")
	ErrNotConnected = errors.New("rabbitmq manager is not connected")
)

type Manager struct {
	url          string
	mu           sync.RWMutex
	conn         *amqp.Connection
	ch           *amqp.Channel
	reconnecting int32 // atomic boolean

	// Reconnection options
	reconnectDelay time.Duration
	maxRetries     int

	ctx        context.Context
	cancelFunc context.CancelFunc
}

// NewManager creates a new Manager instance and initiates connection.
func NewManager(url string) (*Manager, error) {
	return NewManagerWithContext(context.Background(), url)
}

// NewManagerWithContext creates a new Manager instance using the provided context for initialization,
// and manages connection retries and closures.
func NewManagerWithContext(ctx context.Context, url string) (*Manager, error) {
	mCtx, cancel := context.WithCancel(ctx)
	m := &Manager{
		url:            url,
		reconnectDelay: 2 * time.Second,
		maxRetries:     5,
		ctx:            mCtx,
		cancelFunc:     cancel,
	}

	maxStartupRetries := 30
	startupRetryDelay := 2 * time.Second

	var err error
	for i := 0; i < maxStartupRetries; i++ {
		select {
		case <-mCtx.Done():
			cancel()
			return nil, fmt.Errorf("initial rabbitmq connection cancelled: %w", mCtx.Err())
		default:
		}

		err = m.connect()
		if err == nil {
			return m, nil
		}

		log.Printf("[RabbitMQ] Startup connection attempt %d/%d failed: %v. Retrying in %v...", i+1, maxStartupRetries, err, startupRetryDelay)

		select {
		case <-mCtx.Done():
			cancel()
			return nil, fmt.Errorf("initial rabbitmq connection cancelled during wait: %w", mCtx.Err())
		case <-time.After(startupRetryDelay):
		}
	}

	cancel()
	return nil, fmt.Errorf("initial rabbitmq connection failed after %d attempts: %w", maxStartupRetries, err)
}

// Channel returns the current active RabbitMQ channel.
// Note: Interacting with the raw channel directly bypasses the Manager's
// safety mechanisms and is not recommended for concurrent use.
func (m *Manager) Channel() *amqp.Channel {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ch
}

// IsConnected checks if both connection and channel are active and open.
func (m *Manager) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.conn != nil && !m.conn.IsClosed() && m.ch != nil && !m.ch.IsClosed()
}

// connect establishes the connection and channel, and declares the topology.
// It must be called under a write lock or during initialization.
func (m *Manager) connect() error {
	conn, err := Connect(m.url)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("open channel: %w", err)
	}

	// Put the channel in confirm mode for reliable publishing
	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("enable publisher confirms: %w", err)
	}

	// Setup topology (exchanges only)
	if err := SetupTopology(ch); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("setup topology: %w", err)
	}

	m.mu.Lock()
	m.conn = conn
	m.ch = ch
	m.mu.Unlock()

	// Set up notification channels for unexpected closure
	notifyConnClose := make(chan *amqp.Error, 1)
	notifyChanClose := make(chan *amqp.Error, 1)

	conn.NotifyClose(notifyConnClose)
	ch.NotifyClose(notifyChanClose)

	// Launch background watch closure goroutine for this specific connection/channel session
	go m.watchClosure(conn, ch, notifyConnClose, notifyChanClose)

	return nil
}

// watchClosure monitors connection/channel state and triggers reconnection on unexpected close events.
func (m *Manager) watchClosure(conn *amqp.Connection, ch *amqp.Channel, notifyConnClose chan *amqp.Error, notifyChanClose chan *amqp.Error) {
	select {
	case <-m.ctx.Done():
		return
	case err, ok := <-notifyConnClose:
		if ok {
			select {
			case <-m.ctx.Done():
				return
			default:
			}
			log.Printf("[RabbitMQ] Connection closed unexpectedly: %v. Triggering recovery...", err)
			m.reconnect()
		}
	case err, ok := <-notifyChanClose:
		if ok {
			select {
			case <-m.ctx.Done():
				return
			default:
			}
			log.Printf("[RabbitMQ] Channel closed unexpectedly: %v. Triggering recovery...", err)
			m.reconnect()
		}
	}
}

// reconnect executes the recovery loop to re-establish connection and topology.
func (m *Manager) reconnect() {
	select {
	case <-m.ctx.Done():
		return
	default:
	}

	if !atomic.CompareAndSwapInt32(&m.reconnecting, 0, 1) {
		return // Reconnection is already in progress
	}
	defer atomic.StoreInt32(&m.reconnecting, 0)

	m.mu.Lock()
	// Close existing resources to prevent leaks before creating new ones
	if m.ch != nil {
		_ = m.ch.Close()
	}
	if m.conn != nil {
		_ = m.conn.Close()
	}
	m.mu.Unlock()

	backoff := m.reconnectDelay
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-m.ctx.Done():
			return
		default:
		}

		log.Printf("[RabbitMQ] Attempting to reconnect...")
		err := m.connect()
		if err == nil {
			log.Println("[RabbitMQ] Successfully reconnected and redeclared topology")
			return
		}

		log.Printf("[RabbitMQ] Reconnection failed: %v. Retrying in %v...", err, backoff)

		select {
		case <-m.ctx.Done():
			return
		case <-time.After(backoff):
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}

// Publish sends a message to RabbitMQ with a retry mechanism and publisher confirmations.
func (m *Manager) Publish(ctx context.Context, routingKey string, body []byte) error {
	select {
	case <-m.ctx.Done():
		return ErrClosed
	default:
	}

	var err error
	backoff := 100 * time.Millisecond
	maxBackoff := 1 * time.Second

	for i := 0; i < m.maxRetries; i++ {
		err = m.publishOnce(ctx, routingKey, body)
		if err == nil {
			return nil
		}

		log.Printf("[RabbitMQ] Publish failed (attempt %d/%d): %v", i+1, m.maxRetries, err)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-m.ctx.Done():
			return ErrClosed
		case <-time.After(backoff):
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}

	return fmt.Errorf("failed to publish after %d attempts: %w", m.maxRetries, err)
}

// publishOnce performs the actual publish and blocks until confirmation is received.
func (m *Manager) publishOnce(ctx context.Context, routingKey string, body []byte) error {
	m.mu.RLock()
	ch := m.ch
	conn := m.conn
	m.mu.RUnlock()

	if conn == nil || ch == nil || conn.IsClosed() || ch.IsClosed() {
		return ErrNotConnected
	}

	// Use deferred confirmations to ensure reliable delivery (publisher confirms)
	confirm, err := ch.PublishWithDeferredConfirmWithContext(
		ctx,
		DefaultExchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // Make message persistent
			Timestamp:    time.Now().UTC(),
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("publish call: %w", err)
	}

	// Wait for acknowledgment from RabbitMQ broker
	confirmed := confirm.Wait()
	if !confirmed {
		return errors.New("message nack'd by broker")
	}

	return nil
}

// Close gracefully shuts down the Manager, closing connection and channel.
func (m *Manager) Close() error {
	m.cancelFunc()

	m.mu.Lock()
	defer m.mu.Unlock()

	var err error
	if m.ch != nil {
		if chErr := m.ch.Close(); chErr != nil {
			err = fmt.Errorf("close channel: %w", chErr)
		}
	}

	if m.conn != nil {
		if connErr := m.conn.Close(); connErr != nil {
			if err != nil {
				err = fmt.Errorf("%v; close connection: %w", err, connErr)
			} else {
				err = fmt.Errorf("close connection: %w", connErr)
			}
		}
	}

	return err
}
