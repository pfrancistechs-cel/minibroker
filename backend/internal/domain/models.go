package domain

import (
	"time"
)

// Message represents an immutable record within a topic partition commit log.
type Message struct {
	Offset    uint64            `json:"offset"`
	Partition int               `json:"partition"`
	Topic     string            `json:"topic"`
	Key       string            `json:"key,omitempty"`
	Value     string            `json:"value"`
	Headers   map[string]string `json:"headers,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// Topic holds metadata and partition configuration for a message queue.
type Topic struct {
	Name       string    `json:"name"`
	Partitions int       `json:"partitions"`
	CreatedAt  time.Time `json:"created_at"`
}

// PartitionInfo describes partition state and high watermarks.
type PartitionInfo struct {
	ID            int    `json:"id"`
	HighWatermark uint64 `json:"high_watermark"`
	TotalMessages uint64 `json:"total_messages"`
}

// TopicDetail provides comprehensive topic state including partition statistics.
type TopicDetail struct {
	Topic
	PartitionsInfo []PartitionInfo `json:"partitions_info"`
	TotalMessages  uint64          `json:"total_messages"`
}

// ConsumerOffset tracks the committed read position of a consumer group.
type ConsumerOffset struct {
	GroupName     string    `json:"group_name"`
	Topic         string    `json:"topic"`
	Partition     int       `json:"partition"`
	Offset        uint64    `json:"offset"`
	HighWatermark uint64    `json:"high_watermark"`
	Lag           uint64    `json:"lag"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ConsumerGroupSummary groups partition offsets for a single consumer group.
type ConsumerGroupSummary struct {
	GroupName string           `json:"group_name"`
	Topic     string           `json:"topic"`
	Offsets   []ConsumerOffset `json:"offsets"`
	TotalLag  uint64           `json:"total_lag"`
}

// PublishInput is the payload sent by producers.
type PublishInput struct {
	Key     string            `json:"key,omitempty"`
	Value   string            `json:"value"`
	Headers map[string]string `json:"headers,omitempty"`
}

// PublishResult returns the partition and offset assigned to the published message.
type PublishResult struct {
	Topic     string `json:"topic"`
	Partition int    `json:"partition"`
	Offset    uint64 `json:"offset"`
}

// ClusterMetrics reports live throughput and storage telemetry.
type ClusterMetrics struct {
	TotalMessages uint64  `json:"total_messages"`
	MessagesPerSec float64 `json:"messages_per_sec"`
	ActiveTopics  int     `json:"active_topics"`
	ActiveGroups  int     `json:"active_groups"`
	ActiveStreams int     `json:"active_streams"`
	UptimeSeconds int64   `json:"uptime_seconds"`
}

