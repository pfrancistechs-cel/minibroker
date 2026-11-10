package broker

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pfrancistechs-cel/minibroker/backend/internal/domain"
	"github.com/pfrancistechs-cel/minibroker/backend/internal/storage"
)

var (
	ErrTopicNotFound = errors.New("topic not found")
	ErrTopicExists   = errors.New("topic already exists")
)

// Broker represents the core message broker coordinator.
type Broker struct {
	mu            sync.RWMutex
	topics        map[string]*storage.TopicLog
	roundRobin    map[string]*uint64
	db            *storage.DB
	broadcaster   *Broadcaster
	startTime     time.Time
	totalMessages uint64
	lastSecCount  uint64
	msgPerSec     float64
}

// NewBroker initializes the broker engine and restores topics from storage.
func NewBroker(db *storage.DB) (*Broker, error) {
	b := &Broker{
		topics:      make(map[string]*storage.TopicLog),
		roundRobin:  make(map[string]*uint64),
		db:          db,
		broadcaster: NewBroadcaster(),
		startTime:   time.Now(),
	}

	// Restore persisted topics or seed defaults
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	persistedTopics, err := db.GetTopics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed loading topics from db: %w", err)
	}

	if len(persistedTopics) == 0 {
		log.Println("[Broker] No topics detected in database. Initializing default topics...")
		defaults := map[string]int{
			"orders.created":       3,
			"payments.processed":   3,
			"notifications.email": 2,
		}
		for name, parts := range defaults {
			if err := b.CreateTopic(ctx, name, parts); err != nil {
				return nil, err
			}
		}

		// Seed initial demonstration messages so dashboard is alive on boot
		b.seedDemoMessages()
	} else {
		for name, parts := range persistedTopics {
			b.topics[name] = storage.NewTopicLog(name, parts)
			var zero uint64
			b.roundRobin[name] = &zero
		}
		log.Printf("[Broker] Restored %d topics from database\n", len(persistedTopics))
	}

	// Background ticker to calculate throughput (msg/sec)
	go b.throughputLoop()

	return b, nil
}

// throughputLoop updates rolling messages-per-second metric every second.
func (b *Broker) throughputLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var previousTotal uint64
	for range ticker.C {
		current := atomic.LoadUint64(&b.totalMessages)
		diff := current - previousTotal
		previousTotal = current

		b.mu.Lock()
		b.msgPerSec = float64(diff)
		b.mu.Unlock()
	}
}

// CreateTopic provisions a new topic and persists metadata.
func (b *Broker) CreateTopic(ctx context.Context, name string, partitions int) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.topics[name]; exists {
		return ErrTopicExists
	}

	if partitions <= 0 {
		partitions = 1
	}

	b.topics[name] = storage.NewTopicLog(name, partitions)
	var zero uint64
	b.roundRobin[name] = &zero

	if b.db != nil {
		if err := b.db.SaveTopic(ctx, name, partitions); err != nil {
			log.Printf("[Broker] Warning: failed saving topic %s to db: %v\n", name, err)
		}
	}

	log.Printf("[Broker] Topic created: %s (%d partitions)\n", name, partitions)
	return nil
}

// GetTopics lists all registered topics with summary metrics.
func (b *Broker) GetTopics() []domain.TopicDetail {
	b.mu.RLock()
	defer b.mu.RUnlock()

	res := make([]domain.TopicDetail, 0, len(b.topics))
	for name, topicLog := range b.topics {
		res = append(res, domain.TopicDetail{
			Topic: domain.Topic{
				Name:       name,
				Partitions: topicLog.PartitionsCount(),
				CreatedAt:  b.startTime,
			},
			PartitionsInfo: topicLog.PartitionsInfo(),
			TotalMessages:  topicLog.TotalMessages(),
		})
	}
	return res
}

// GetTopicDetail returns single topic status.
func (b *Broker) GetTopicDetail(name string) (*domain.TopicDetail, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	topicLog, ok := b.topics[name]
	if !ok {
		return nil, ErrTopicNotFound
	}

	return &domain.TopicDetail{
		Topic: domain.Topic{
			Name:       name,
			Partitions: topicLog.PartitionsCount(),
			CreatedAt:  b.startTime,
		},
		PartitionsInfo: topicLog.PartitionsInfo(),
		TotalMessages:  topicLog.TotalMessages(),
	}, nil
}

// Publish writes a message to the target topic using hash or round-robin partition strategy.
func (b *Broker) Publish(ctx context.Context, topic string, input domain.PublishInput) (*domain.PublishResult, error) {
	b.mu.RLock()
	topicLog, ok := b.topics[topic]
	rrPtr := b.roundRobin[topic]
	b.mu.RUnlock()

	if !ok {
		return nil, ErrTopicNotFound
	}

	numPartitions := topicLog.PartitionsCount()
	var partitionID int

	if input.Key != "" {
		// Consistent Hash Partitioning (FNV-1a)
		h := fnv.New32a()
		h.Write([]byte(input.Key))
		partitionID = int(h.Sum32() % uint32(numPartitions))
	} else {
		// Round-Robin Partitioning
		next := atomic.AddUint64(rrPtr, 1)
		partitionID = int(next % uint64(numPartitions))
	}

	partitionLog, err := topicLog.GetPartition(partitionID)
	if err != nil {
		return nil, err
	}

	msg := domain.Message{
		Topic:     topic,
		Partition: partitionID,
		Key:       input.Key,
		Value:     input.Value,
		Headers:   input.Headers,
		Timestamp: time.Now().UTC(),
	}

	offset := partitionLog.Append(msg)
	msg.Offset = offset

	atomic.AddUint64(&b.totalMessages, 1)

	// Broadcast to active SSE listeners
	b.broadcaster.Broadcast(topic, msg)

	// Async audit record to PostgreSQL
	if b.db != nil {
		go b.db.AuditMessage(context.Background(), topic, partitionID, offset, input.Key, input.Value)
	}

	return &domain.PublishResult{
		Topic:     topic,
		Partition: partitionID,
		Offset:    offset,
	}, nil
}

// ReadMessages fetches messages starting from a given offset.
func (b *Broker) ReadMessages(topic string, partition int, offset uint64, limit int) ([]domain.Message, error) {
	b.mu.RLock()
	topicLog, ok := b.topics[topic]
	b.mu.RUnlock()

	if !ok {
		return nil, ErrTopicNotFound
	}

	partLog, err := topicLog.GetPartition(partition)
	if err != nil {
		return nil, err
	}

	return partLog.ReadFrom(offset, limit), nil
}

// CommitOffset registers committed position for a consumer group.
func (b *Broker) CommitOffset(ctx context.Context, group, topic string, partition int, offset uint64) error {
	if b.db == nil {
		return nil
	}
	return b.db.CommitOffset(ctx, group, topic, partition, offset)
}

// GetConsumerGroups aggregates offsets and computes lag per partition.
func (b *Broker) GetConsumerGroups(ctx context.Context) ([]domain.ConsumerGroupSummary, error) {
	if b.db == nil {
		return []domain.ConsumerGroupSummary{}, nil
	}

	rawOffsets, err := b.db.GetConsumerOffsets(ctx)
	if err != nil {
		return nil, err
	}

	groupsMap := make(map[string]map[string][]domain.ConsumerOffset)

	for _, ro := range rawOffsets {
		if _, ok := groupsMap[ro.Group]; !ok {
			groupsMap[ro.Group] = make(map[string][]domain.ConsumerOffset)
		}

		// Fetch high watermark to compute lag
		var hw uint64
		b.mu.RLock()
		if t, ok := b.topics[ro.Topic]; ok {
			if p, err := t.GetPartition(ro.Partition); err == nil {
				hw = p.HighWatermark()
			}
		}
		b.mu.RUnlock()

		lag := uint64(0)
		if hw > ro.Offset {
			lag = hw - ro.Offset
		}

		co := domain.ConsumerOffset{
			GroupName:     ro.Group,
			Topic:         ro.Topic,
			Partition:     ro.Partition,
			Offset:        ro.Offset,
			HighWatermark: hw,
			Lag:           lag,
			UpdatedAt:     ro.UpdatedAt,
		}

		groupsMap[ro.Group][ro.Topic] = append(groupsMap[ro.Group][ro.Topic], co)
	}

	var res []domain.ConsumerGroupSummary
	for grp, topics := range groupsMap {
		for top, offsets := range topics {
			var totalLag uint64
			for _, o := range offsets {
				totalLag += o.Lag
			}
			res = append(res, domain.ConsumerGroupSummary{
				GroupName: grp,
				Topic:     top,
				Offsets:   offsets,
				TotalLag:  totalLag,
			})
		}
	}

	return res, nil
}

// GetMetrics aggregates real-time broker statistics.
func (b *Broker) GetMetrics() domain.ClusterMetrics {
	b.mu.RLock()
	defer b.mu.RUnlock()

	activeTopics := len(b.topics)
	streams := b.broadcaster.ActiveSubscribers()
	uptime := int64(time.Since(b.startTime).Seconds())

	return domain.ClusterMetrics{
		TotalMessages:  atomic.LoadUint64(&b.totalMessages),
		MessagesPerSec: b.msgPerSec,
		ActiveTopics:   activeTopics,
		ActiveGroups:   3, // Standard seeded consumers
		ActiveStreams:  streams,
		UptimeSeconds:  uptime,
	}
}

// Subscribe returns an SSE listener channel for topic messages.
func (b *Broker) Subscribe(topic string) chan domain.Message {
	return b.broadcaster.Subscribe(topic)
}

// Unsubscribe releases an SSE listener channel.
func (b *Broker) Unsubscribe(topic string, ch chan domain.Message) {
	b.broadcaster.Unsubscribe(topic, ch)
}

func (b *Broker) seedDemoMessages() {
	sampleMsgs := []struct {
		topic string
		key   string
		val   string
	}{
		{"orders.created", "cust_101", `{"order_id":"ord_88921","customer_id":"cust_101","amount":1299.00,"currency":"USD","items":["MacBook Pro 16"]}`},
		{"orders.created", "cust_102", `{"order_id":"ord_88922","customer_id":"cust_102","amount":249.00,"currency":"USD","items":["AirPods Pro"]}`},
		{"payments.processed", "pay_701", `{"payment_id":"pay_701","order_id":"ord_88921","status":"CONFIRMED","provider":"stripe"}`},
		{"notifications.email", "usr_301", `{"to":"alex@example.com","template":"order_confirmation","order_id":"ord_88921"}`},
	}

	for _, sm := range sampleMsgs {
		_, _ = b.Publish(context.Background(), sm.topic, domain.PublishInput{
			Key:   sm.key,
			Value: sm.val,
			Headers: map[string]string{
				"env":    "production",
				"origin": "checkout-svc",
			},
		})
	}

	// Seed sample consumer group offsets
	_ = b.CommitOffset(context.Background(), "order-processing-worker", "orders.created", 0, 1)
	_ = b.CommitOffset(context.Background(), "order-processing-worker", "orders.created", 1, 0)
	_ = b.CommitOffset(context.Background(), "analytics-sync-group", "payments.processed", 0, 1)
}
