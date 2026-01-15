package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	// Stream names
	StreamBusinessOS = "BUSINESSOS_SYNC"
	StreamOSA        = "OSA_EVENTS"

	// Subject prefixes
	SubjectBusinessOSPrefix = "businessos"
	SubjectOSAPrefix        = "osa"

	// Consumer configuration
	DefaultAckWait     = 30 * time.Second
	DefaultMaxDeliver  = 5
	DefaultMaxAckPending = 1000

	// Connection configuration
	DefaultReconnectWait = 2 * time.Second
	DefaultMaxReconnects = -1 // Infinite
	DefaultPingInterval  = 20 * time.Second
	DefaultMaxPingsOut   = 3
)

// NATSClient manages NATS JetStream connections for bidirectional sync
type NATSClient struct {
	conn     *nats.Conn
	js       jetstream.JetStream
	logger   *slog.Logger
	url      string
	mu       sync.RWMutex
	closed   bool
	closeCh  chan struct{}
	subjects map[string]*subscription // Track active subscriptions
}

// subscription tracks an active consumer subscription
type subscription struct {
	consumer jetstream.Consumer
	ctx      jetstream.ConsumeContext
	cancel   context.CancelFunc
}

// NewNATSClient creates a new NATS client with JetStream support
func NewNATSClient(url string, logger *slog.Logger) (*NATSClient, error) {
	if url == "" {
		return nil, fmt.Errorf("NATS URL cannot be empty")
	}

	if logger == nil {
		logger = slog.Default()
	}

	client := &NATSClient{
		url:      url,
		logger:   logger,
		closeCh:  make(chan struct{}),
		subjects: make(map[string]*subscription),
	}

	if err := client.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return client, nil
}

// connect establishes connection to NATS and initializes JetStream
func (c *NATSClient) connect() error {
	c.logger.Info("connecting to NATS", "url", c.url)

	opts := []nats.Option{
		nats.Name("BusinessOS-Sync-Client"),
		nats.ReconnectWait(DefaultReconnectWait),
		nats.MaxReconnects(DefaultMaxReconnects),
		nats.PingInterval(DefaultPingInterval),
		nats.MaxPingsOutstanding(DefaultMaxPingsOut),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				c.logger.Warn("NATS disconnected", "error", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			c.logger.Info("NATS reconnected", "url", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			c.logger.Info("NATS connection closed")
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			c.logger.Error("NATS error", "error", err)
		}),
	}

	conn, err := nats.Connect(c.url, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	c.conn = conn

	// Initialize JetStream context
	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create JetStream context: %w", err)
	}

	c.js = js

	c.logger.Info("connected to NATS", "url", conn.ConnectedUrl())
	return nil
}

// CreateStreams creates the BUSINESSOS_SYNC and OSA_EVENTS streams
func (c *NATSClient) CreateStreams(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}

	// Create BusinessOS Sync Stream
	businessOSConfig := jetstream.StreamConfig{
		Name:        StreamBusinessOS,
		Description: "BusinessOS to OSA sync events",
		Subjects: []string{
			fmt.Sprintf("%s.users.>", SubjectBusinessOSPrefix),
			fmt.Sprintf("%s.workspaces.>", SubjectBusinessOSPrefix),
			fmt.Sprintf("%s.apps.>", SubjectBusinessOSPrefix),
		},
		Storage:   jetstream.FileStorage,
		Retention: jetstream.LimitsPolicy,
		MaxAge:    7 * 24 * time.Hour, // 7 days retention
		MaxBytes:  1024 * 1024 * 1024, // 1GB
		Discard:   jetstream.DiscardOld,
		Replicas:  1,
	}

	if err := c.createOrUpdateStream(ctx, businessOSConfig); err != nil {
		return fmt.Errorf("failed to create BusinessOS stream: %w", err)
	}

	// Create OSA Events Stream
	osaConfig := jetstream.StreamConfig{
		Name:        StreamOSA,
		Description: "OSA to BusinessOS events",
		Subjects: []string{
			fmt.Sprintf("%s.builds.>", SubjectOSAPrefix),
			fmt.Sprintf("%s.apps.>", SubjectOSAPrefix),
			fmt.Sprintf("%s.workspaces.>", SubjectOSAPrefix),
		},
		Storage:   jetstream.FileStorage,
		Retention: jetstream.LimitsPolicy,
		MaxAge:    7 * 24 * time.Hour, // 7 days retention
		MaxBytes:  1024 * 1024 * 1024, // 1GB
		Discard:   jetstream.DiscardOld,
		Replicas:  1,
	}

	if err := c.createOrUpdateStream(ctx, osaConfig); err != nil {
		return fmt.Errorf("failed to create OSA stream: %w", err)
	}

	c.logger.Info("streams created successfully",
		"businessos_stream", StreamBusinessOS,
		"osa_stream", StreamOSA)

	return nil
}

// createOrUpdateStream creates a stream or updates it if it already exists
func (c *NATSClient) createOrUpdateStream(ctx context.Context, config jetstream.StreamConfig) error {
	stream, err := c.js.CreateStream(ctx, config)
	if err != nil {
		// If stream already exists, try to update it
		if err == jetstream.ErrStreamNameAlreadyInUse {
			stream, err = c.js.UpdateStream(ctx, config)
			if err != nil {
				return fmt.Errorf("failed to update stream %s: %w", config.Name, err)
			}
			c.logger.Info("stream updated", "name", config.Name)
		} else {
			return fmt.Errorf("failed to create stream %s: %w", config.Name, err)
		}
	} else {
		c.logger.Info("stream created", "name", config.Name)
	}

	// Verify stream info
	info, err := stream.Info(ctx)
	if err != nil {
		return fmt.Errorf("failed to get stream info: %w", err)
	}

	c.logger.Info("stream info",
		"name", info.Config.Name,
		"subjects", info.Config.Subjects,
		"messages", info.State.Msgs,
		"bytes", info.State.Bytes)

	return nil
}

// Publish publishes a message to a subject with idempotency
func (c *NATSClient) Publish(ctx context.Context, subject string, data interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}

	// Marshal data to JSON
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Generate idempotency key (UUIDv7 for time-ordering)
	msgID := uuid.New().String()

	// Publish with message ID for idempotency
	msg := nats.NewMsg(subject)
	msg.Data = payload
	msg.Header.Set(nats.MsgIdHdr, msgID)

	_, err = c.js.PublishMsg(ctx, msg)
	if err != nil {
		c.logger.Error("failed to publish message",
			"subject", subject,
			"msg_id", msgID,
			"error", err)
		return fmt.Errorf("failed to publish: %w", err)
	}

	c.logger.Debug("message published",
		"subject", subject,
		"msg_id", msgID,
		"size_bytes", len(payload))

	return nil
}

// Subscribe creates a durable consumer and subscribes to messages
func (c *NATSClient) Subscribe(ctx context.Context, subject string, handler MessageHandler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}

	// Check if already subscribed
	if _, exists := c.subjects[subject]; exists {
		return fmt.Errorf("already subscribed to subject: %s", subject)
	}

	// Determine stream based on subject prefix
	streamName := c.getStreamForSubject(subject)
	if streamName == "" {
		return fmt.Errorf("no stream found for subject: %s", subject)
	}

	// Generate consumer name from subject (replace . with _)
	consumerName := fmt.Sprintf("consumer_%s", subject)

	// Create or get consumer
	consumer, err := c.js.CreateOrUpdateConsumer(ctx, streamName, jetstream.ConsumerConfig{
		Name:          consumerName,
		Durable:       consumerName,
		Description:   fmt.Sprintf("Consumer for %s", subject),
		FilterSubject: subject,
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       DefaultAckWait,
		MaxDeliver:    DefaultMaxDeliver,
		MaxAckPending: DefaultMaxAckPending,
	})
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	// Create context for this subscription
	_, cancel := context.WithCancel(context.Background())

	// Start consuming messages
	consumeCtx, err := consumer.Consume(func(msg jetstream.Msg) {
		if err := c.handleMessage(msg, handler); err != nil {
			c.logger.Error("failed to handle message",
				"subject", msg.Subject(),
				"error", err)
			// Negative ack to retry
			msg.Nak()
		} else {
			// Acknowledge successful processing
			msg.Ack()
		}
	})
	if err != nil {
		cancel()
		return fmt.Errorf("failed to consume: %w", err)
	}

	// Store subscription
	c.subjects[subject] = &subscription{
		consumer: consumer,
		ctx:      consumeCtx,
		cancel:   cancel,
	}

	c.logger.Info("subscribed to subject",
		"subject", subject,
		"stream", streamName,
		"consumer", consumerName)

	return nil
}

// handleMessage processes an individual message
func (c *NATSClient) handleMessage(msg jetstream.Msg, handler MessageHandler) error {
	metadata, err := msg.Metadata()
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}

	// Extract message details
	msgID := msg.Headers().Get(nats.MsgIdHdr)

	c.logger.Debug("processing message",
		"subject", msg.Subject(),
		"msg_id", msgID,
		"stream", metadata.Stream,
		"sequence", metadata.Sequence.Stream,
		"pending", metadata.NumPending,
		"delivered", metadata.NumDelivered)

	// Create message context
	msgCtx := &Message{
		Subject:      msg.Subject(),
		Data:         msg.Data(),
		Headers:      msg.Headers(),
		MessageID:    msgID,
		Sequence:     metadata.Sequence.Stream,
		Stream:       metadata.Stream,
		NumDelivered: int(metadata.NumDelivered),
		Timestamp:    metadata.Timestamp,
	}

	// Call handler
	if err := handler(msgCtx); err != nil {
		return err
	}

	return nil
}

// Unsubscribe stops consuming from a subject
func (c *NATSClient) Unsubscribe(subject string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	sub, exists := c.subjects[subject]
	if !exists {
		return fmt.Errorf("not subscribed to subject: %s", subject)
	}

	// Stop consuming
	sub.ctx.Stop()

	// Cancel context
	sub.cancel()

	// Remove from map
	delete(c.subjects, subject)

	c.logger.Info("unsubscribed from subject", "subject", subject)
	return nil
}

// getStreamForSubject determines which stream a subject belongs to
func (c *NATSClient) getStreamForSubject(subject string) string {
	if len(subject) >= len(SubjectBusinessOSPrefix) &&
		subject[:len(SubjectBusinessOSPrefix)] == SubjectBusinessOSPrefix {
		return StreamBusinessOS
	}
	if len(subject) >= len(SubjectOSAPrefix) &&
		subject[:len(SubjectOSAPrefix)] == SubjectOSAPrefix {
		return StreamOSA
	}
	return ""
}

// Close gracefully shuts down the client
func (c *NATSClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	close(c.closeCh)

	// Stop all subscriptions
	for subject, sub := range c.subjects {
		sub.ctx.Stop()
		sub.cancel()
		c.logger.Info("stopped subscription", "subject", subject)
	}

	// Close NATS connection
	if c.conn != nil {
		c.conn.Close()
		c.logger.Info("NATS connection closed")
	}

	return nil
}

// IsConnected returns whether the client is connected
func (c *NATSClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return false
	}

	return c.conn != nil && c.conn.IsConnected()
}

// Stats returns connection statistics
func (c *NATSClient) Stats() *Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil {
		return &Stats{Connected: false}
	}

	stats := c.conn.Stats()

	return &Stats{
		Connected:       c.conn.IsConnected(),
		URL:             c.conn.ConnectedUrl(),
		InMsgs:          stats.InMsgs,
		OutMsgs:         stats.OutMsgs,
		InBytes:         stats.InBytes,
		OutBytes:        stats.OutBytes,
		Reconnects:      stats.Reconnects,
		ActiveSubscriptions: len(c.subjects),
	}
}

// Message represents a NATS message with metadata
type Message struct {
	Subject      string
	Data         []byte
	Headers      nats.Header
	MessageID    string
	Sequence     uint64
	Stream       string
	NumDelivered int
	Timestamp    time.Time
}

// Unmarshal decodes the message data into the provided value
func (m *Message) Unmarshal(v interface{}) error {
	return json.Unmarshal(m.Data, v)
}

// MessageHandler is called for each received message
type MessageHandler func(*Message) error

// Stats contains connection statistics
type Stats struct {
	Connected           bool
	URL                 string
	InMsgs              uint64
	OutMsgs             uint64
	InBytes             uint64
	OutBytes            uint64
	Reconnects          uint64
	ActiveSubscriptions int
}

// Subject helpers for common patterns

// SubjectUserCreated returns the subject for user created events
func SubjectUserCreated(userID uuid.UUID) string {
	return fmt.Sprintf("%s.users.created.%s", SubjectBusinessOSPrefix, userID.String())
}

// SubjectUserUpdated returns the subject for user updated events
func SubjectUserUpdated(userID uuid.UUID) string {
	return fmt.Sprintf("%s.users.updated.%s", SubjectBusinessOSPrefix, userID.String())
}

// SubjectWorkspaceCreated returns the subject for workspace created events
func SubjectWorkspaceCreated(workspaceID uuid.UUID) string {
	return fmt.Sprintf("%s.workspaces.created.%s", SubjectBusinessOSPrefix, workspaceID.String())
}

// SubjectWorkspaceUpdated returns the subject for workspace updated events
func SubjectWorkspaceUpdated(workspaceID uuid.UUID) string {
	return fmt.Sprintf("%s.workspaces.updated.%s", SubjectBusinessOSPrefix, workspaceID.String())
}

// SubjectBuildStarted returns the subject for build started events
func SubjectBuildStarted(buildID string) string {
	return fmt.Sprintf("%s.builds.started.%s", SubjectOSAPrefix, buildID)
}

// SubjectBuildProgress returns the subject for build progress events
func SubjectBuildProgress(buildID string) string {
	return fmt.Sprintf("%s.builds.progress.%s", SubjectOSAPrefix, buildID)
}

// SubjectBuildCompleted returns the subject for build completed events
func SubjectBuildCompleted(buildID string) string {
	return fmt.Sprintf("%s.builds.completed.%s", SubjectOSAPrefix, buildID)
}

// SubjectAppGenerated returns the subject for app generated events
func SubjectAppGenerated(appID uuid.UUID) string {
	return fmt.Sprintf("%s.apps.generated.%s", SubjectOSAPrefix, appID.String())
}

// Wildcard subscription subjects

// SubjectAllUsers subscribes to all user events
const SubjectAllUsers = "businessos.users.*"

// SubjectAllWorkspaces subscribes to all workspace events
const SubjectAllWorkspaces = "businessos.workspaces.*"

// SubjectAllBuilds subscribes to all build events
const SubjectAllBuilds = "osa.builds.*"

// SubjectAllApps subscribes to all app events
const SubjectAllApps = "osa.apps.*"
