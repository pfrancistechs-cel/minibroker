package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps PostgreSQL connection pool for durable metadata and offset tracking.
type DB struct {
	Pool *pgxpool.Pool
}

// ConnectAndMigrate connects to PostgreSQL (Neon compatible) and runs migrations.
func ConnectAndMigrate(connStr string) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("invalid connection string: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 2
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 15 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database unreachable: %w", err)
	}
	log.Println("[Database] Connected to PostgreSQL (Neon) successfully")

	db := &DB{Pool: pool}
	if err := db.migrate(ctx); err != nil {
		return nil, fmt.Errorf("migration failure: %w", err)
	}

	return db, nil
}

func (db *DB) migrate(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS topics (
			name VARCHAR(255) PRIMARY KEY,
			partitions INT NOT NULL DEFAULT 1,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS consumer_offsets (
			group_name VARCHAR(255) NOT NULL,
			topic VARCHAR(255) NOT NULL,
			partition INT NOT NULL,
			offset_val BIGINT NOT NULL DEFAULT 0,
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			PRIMARY KEY (group_name, topic, partition)
		);`,
		`CREATE TABLE IF NOT EXISTS messages_audit (
			id BIGSERIAL PRIMARY KEY,
			topic VARCHAR(255) NOT NULL,
			partition INT NOT NULL,
			offset_val BIGINT NOT NULL,
			msg_key TEXT,
			msg_value TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE INDEX IF NOT EXISTS idx_audit_topic_partition ON messages_audit(topic, partition, offset_val);`,
	}

	for _, q := range queries {
		if _, err := db.Pool.Exec(ctx, q); err != nil {
			return fmt.Errorf("failed executing: %s: %w", q, err)
		}
	}

	log.Println("[Database] MiniBroker schema migrations verified")
	return nil
}

// SaveTopic persists a topic definition.
func (db *DB) SaveTopic(ctx context.Context, name string, partitions int) error {
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO topics (name, partitions)
		VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET partitions = EXCLUDED.partitions
	`, name, partitions)
	return err
}

// GetTopics fetches all persisted topics.
func (db *DB) GetTopics(ctx context.Context) (map[string]int, error) {
	rows, err := db.Pool.Query(ctx, "SELECT name, partitions FROM topics")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[string]int)
	for rows.Next() {
		var name string
		var parts int
		if err := rows.Scan(&name, &parts); err == nil {
			res[name] = parts
		}
	}
	return res, nil
}

// CommitOffset updates the committed offset for a consumer group on a partition.
func (db *DB) CommitOffset(ctx context.Context, group, topic string, partition int, offset uint64) error {
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO consumer_offsets (group_name, topic, partition, offset_val, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (group_name, topic, partition)
		DO UPDATE SET offset_val = EXCLUDED.offset_val, updated_at = NOW()
	`, group, topic, partition, int64(offset))
	return err
}

// GetConsumerOffsets retrieves offsets stored for consumer groups.
func (db *DB) GetConsumerOffsets(ctx context.Context) ([]struct {
	Group     string
	Topic     string
	Partition int
	Offset    uint64
	UpdatedAt time.Time
}, error) {
	rows, err := db.Pool.Query(ctx, "SELECT group_name, topic, partition, offset_val, updated_at FROM consumer_offsets")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []struct {
		Group     string
		Topic     string
		Partition int
		Offset    uint64
		UpdatedAt time.Time
	}

	for rows.Next() {
		var item struct {
			Group     string
			Topic     string
			Partition int
			Offset    uint64
			UpdatedAt time.Time
		}
		var offVal int64
		if err := rows.Scan(&item.Group, &item.Topic, &item.Partition, &offVal, &item.UpdatedAt); err == nil {
			item.Offset = uint64(offVal)
			res = append(res, item)
		}
	}
	return res, nil
}

// AuditMessage records a message into PostgreSQL audit log asynchronously.
func (db *DB) AuditMessage(ctx context.Context, topic string, partition int, offset uint64, key, value string) {
	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO messages_audit (topic, partition, offset_val, msg_key, msg_value)
		VALUES ($1, $2, $3, $4, $5)
	`, topic, partition, int64(offset), key, value)
}
