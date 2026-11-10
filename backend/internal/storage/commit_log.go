package storage

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/pfrancistechs-cel/minibroker/backend/internal/domain"
)

var (
	ErrPartitionNotFound = errors.New("partition not found")
	ErrOffsetOutOfBounds = errors.New("offset out of bounds")
)

// PartitionLog represents an append-only sequence of messages for a single partition.
type PartitionLog struct {
	mu            sync.RWMutex
	id            int
	messages      []domain.Message
	highWatermark uint64
}

// NewPartitionLog initializes an in-memory commit log partition.
func NewPartitionLog(id int) *PartitionLog {
	return &PartitionLog{
		id:            id,
		messages:      make([]domain.Message, 0, 1024),
		highWatermark: 0,
	}
}

// Append writes an immutable message to the partition log and increments the high watermark.
func (p *PartitionLog) Append(msg domain.Message) uint64 {
	p.mu.Lock()
	defer p.mu.Unlock()

	offset := p.highWatermark
	msg.Offset = offset
	msg.Partition = p.id

	p.messages = append(p.messages, msg)
	atomic.AddUint64(&p.highWatermark, 1)

	return offset
}

// ReadFrom reads up to `limit` messages beginning at `offset`.
func (p *PartitionLog) ReadFrom(offset uint64, limit int) []domain.Message {
	p.mu.RLock()
	defer p.mu.RUnlock()

	total := uint64(len(p.messages))
	if offset >= total {
		return []domain.Message{}
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	end := offset + uint64(limit)
	if end > total {
		end = total
	}

	res := make([]domain.Message, end-offset)
	copy(res, p.messages[offset:end])
	return res
}

// HighWatermark returns the next offset to be written.
func (p *PartitionLog) HighWatermark() uint64 {
	return atomic.LoadUint64(&p.highWatermark)
}

// TotalMessages returns the count of messages in this partition.
func (p *PartitionLog) TotalMessages() uint64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return uint64(len(p.messages))
}

// TopicLog manages a collection of partition logs for a single topic.
type TopicLog struct {
	mu         sync.RWMutex
	name       string
	partitions []*PartitionLog
}

// NewTopicLog creates a new topic with the specified partition count.
func NewTopicLog(name string, partitionCount int) *TopicLog {
	if partitionCount <= 0 {
		partitionCount = 1
	}
	parts := make([]*PartitionLog, partitionCount)
	for i := 0; i < partitionCount; i++ {
		parts[i] = NewPartitionLog(i)
	}

	return &TopicLog{
		name:       name,
		partitions: parts,
	}
}

// GetPartition retrieves a partition log by ID.
func (t *TopicLog) GetPartition(id int) (*PartitionLog, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if id < 0 || id >= len(t.partitions) {
		return nil, ErrPartitionNotFound
	}
	return t.partitions[id], nil
}

// PartitionsCount returns the number of partitions.
func (t *TopicLog) PartitionsCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.partitions)
}

// TotalMessages counts messages across all partitions.
func (t *TopicLog) TotalMessages() uint64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var total uint64
	for _, p := range t.partitions {
		total += p.TotalMessages()
	}
	return total
}

// PartitionsInfo summarizes the state of all partitions in this topic.
func (t *TopicLog) PartitionsInfo() []domain.PartitionInfo {
	t.mu.RLock()
	defer t.mu.RUnlock()

	res := make([]domain.PartitionInfo, len(t.partitions))
	for i, p := range t.partitions {
		res[i] = domain.PartitionInfo{
			ID:            i,
			HighWatermark: p.HighWatermark(),
			TotalMessages: p.TotalMessages(),
		}
	}
	return res
}
