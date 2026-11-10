# MiniBroker ⚡ 📨

> **A high-throughput distributed message broker with partitioned commit logs, consumer group offset tracking, and live SSE streaming built with Go, PostgreSQL (Neon), and Next.js 15.**

Inspired by Apache Kafka and NATS, MiniBroker delivers partitioned append-only write-ahead logs, FNV-1a consistent hash routing, consumer lag monitoring in PostgreSQL, and real-time streaming dashboards over Server-Sent Events (SSE).

---

## 🏗️ Architecture

```text
                        ┌──────────────────────────────────────────────┐
                        │          MiniBroker Visual Console           │
                        │ (Next.js 15, Live SSE Ticker, Lag Dashboard) │
                        └──────────────────────┬───────────────────────┘
                                               │
                        ┌──────────────────────┴──────────────────────┐
                        │ HTTP REST & SSE Stream (/api/topics/*/stream)│
                        ▼                                             ▼
┌──────────────────────────────────────────────────────────────────────────────────┐
│                            MiniBroker Cluster Coordinator                        │
│                                                                                  │
│   ┌─────────────────────┐   ┌───────────────────────┐   ┌────────────────────┐   │
│   │   Producer Router   │──▶│   SSE Broadcaster     │──▶│   Rolling Metrics  │   │
│   │ (FNV-1a / Round-Rob)│   │   (Channel Multiplex) │   │   (msg/s, uptime)  │   │
│   └──────────┬──────────┘   └───────────────────────┘   └────────────────────┘   │
│              │                                                                   │
│              ▼                                                                   │
│   ┌──────────────────────────────────────────────────────────────────────────┐   │
│   │                 In-Memory Append-Only Commit Log (WAL)                   │   │
│   │  Partition 0 [Offset 0, 1, 2...]  |  Partition 1 [Offset 0, 1, 2...]     │   │
│   └──────────────────────────────────┬───────────────────────────────────────┘   │
└──────────────────────────────────────┼───────────────────────────────────────────┘
                                       │
                                       ▼
┌──────────────────────────────────────────────────────────────────────────────────┐
│                         PostgreSQL (Neon / Local)                                │
│                                                                                  │
│   ┌────────────────────────────────┐       ┌─────────────────────────────────┐   │
│   │   consumer_offsets             │       │     topics & partition metadata │   │
│   │   Committed group positions    │       │     Durable cluster registry    │   │
│   └────────────────────────────────┘       └─────────────────────────────────┘   │
│   ┌────────────────────────────────┐                                             │
│   │   messages_audit               │                                             │
│   │   Async audit log persistence  │                                             │
│   └────────────────────────────────┘                                             │
└──────────────────────────────────────────────────────────────────────────────────┘
```

---

## 🚀 Key Capabilities

- **⚡ Partitioned Append-Only Log:** Each partition writes to an isolated commit log with 64-bit monotonic sequence offsets and zero-copy slice reads.
- **🎯 Intelligent Partition Routing:**
  - If a message has a `key`, MiniBroker uses **FNV-1a consistent hashing** to guarantee ordered per-entity delivery to the same partition.
  - If no `key` is supplied, MiniBroker distributes messages using atomic **round-robin** rotation.
- **👥 Consumer Groups & Lag Monitoring (Neon PostgreSQL):** Group coordinator stores committed consumer group offsets in Neon PostgreSQL and computes real-time partition lag (`high_watermark - committed_offset`).
- **📡 Native Server-Sent Events (SSE):** Real-time `/api/topics/{topic}/stream` multiplexer dispatches live messages to web consoles without polling.
- **📊 Real-Time Cluster Telemetry:** Rolling 1-second throughput tracking (messages/sec), active streams count, and partition statistics.

---

## 📂 Project Structure

```text
minibroker/
├── backend/
│   ├── cmd/
│   │   └── broker/
│   │       └── main.go           # Entrypoint, CORS, graceful shutdown
│   ├── internal/
│   │   ├── domain/
│   │   │   └── models.go         # Message, Topic, ConsumerOffset, Metrics
│   │   ├── storage/
│   │   │   ├── commit_log.go     # High-concurrency WAL partition logs
│   │   │   └── db.go             # Neon PostgreSQL migrations & offset persistence
│   │   ├── broker/
│   │   │   ├── broker.go         # Cluster coordinator, routing, demo seeds
│   │   │   └── broadcaster.go    # SSE pub/sub channel multiplexer
│   │   └── handler/
│   │       └── handler.go        # HTTP & SSE endpoints
│   ├── .env.example
│   └── go.mod
│
├── frontend/
│   ├── src/
│   │   ├── app/
│   │   │   ├── layout.tsx
│   │   │   └── page.tsx          # Real-time message streaming dashboard
│   │   ├── components/
│   │   │   ├── Header.tsx        # Cluster telemetry (msg/s, uptime, total)
│   │   │   ├── TopicList.tsx     # Topic manager & partition inspector
│   │   │   ├── MessagePublisher.tsx # JSON/Text publisher with partition key
│   │   │   ├── LiveStreamViewer.tsx # SSE auto-scroll console with P-0/P-1 tags
│   │   │   └── ConsumerGroupLag.tsx # Visual progress bars & lag monitor
│   │   ├── lib/
│   │   │   └── api.ts            # REST client & EventSource SSE subscriber
│   │   └── types/
│   │       └── index.ts          # TypeScript interfaces
│   ├── .env.example
│   └── package.json
│
├── docker-compose.yml            # Local PostgreSQL 16 container
├── .gitignore
└── README.md
```

---

## 🛠️ Quickstart & Setup

### 1. Database Setup (Choose Neon or Local Docker)

#### Option A: Neon Cloud PostgreSQL

Create a database on [Neon.tech](https://neon.tech) and copy your connection string (`sslmode=require`).

#### Option B: Local Docker PostgreSQL

```bash
docker-compose up -d
```

_(Runs PostgreSQL 16 on `localhost:5434`)_

---

### 2. Backend Setup (`/backend`)

1. Copy `.env.example` to `.env`:
   ```bash
   cd backend
   cp .env.example .env
   ```
2. Configure `DATABASE_URL` in `.env`:
   ```ini
   PORT=8082
   # For Neon:
   DATABASE_URL=postgres://neondb_owner:YOUR_PASSWORD@ep-sample-12345.us-east-2.aws.neon.tech/neondb?sslmode=require
   # For Local Docker:
   # DATABASE_URL=postgres://minibroker:minibroker_password@localhost:5434/minibroker?sslmode=disable
   ```
3. Run the Go broker daemon:
   ```bash
   go run ./cmd/broker
   ```
   > **Note:** On first startup, the database automatically initializes tables and provisions default topics (`orders.created`, `payments.processed`, `notifications.email`).

---

### 3. Frontend Setup (`/frontend`)

1. Copy `.env.example` to `.env.local`:
   ```bash
   cd ../frontend
   cp .env.example .env.local
   ```
2. Start the Next.js development server:
   ```bash
   npm run dev
   ```
3. Open **[http://localhost:3000](http://localhost:3000)** in your browser.

---

## 📡 REST & Streaming API Reference

| Method | Endpoint                                              | Description                                             |
| ------ | ----------------------------------------------------- | ------------------------------------------------------- |
| `GET`  | `/health`                                             | Cluster health check                                    |
| `GET`  | `/api/topics`                                         | List all topics with partition summaries                |
| `POST` | `/api/topics`                                         | Provision a new topic (`name`, `partitions`)            |
| `GET`  | `/api/topics/{topic}`                                 | Get topic partition watermarks                          |
| `POST` | `/api/topics/{topic}/publish`                         | Produce message (`key`, `value`, `headers`)             |
| `GET`  | `/api/topics/{topic}/partitions/{partition}/messages` | Batch consume messages (`offset`, `limit`)              |
| `GET`  | `/api/topics/{topic}/stream`                          | **Server-Sent Events (SSE)** real-time stream           |
| `GET`  | `/api/consumer-groups`                                | List consumer groups and partition lag                  |
| `POST` | `/api/consumer-groups/{group}/offsets`                | Commit consumer offset (`topic`, `partition`, `offset`) |
| `GET`  | `/api/metrics`                                        | Instantaneous cluster metrics                           |
| `GET`  | `/api/metrics/stream`                                 | **SSE Stream** of 1-second cluster telemetry            |

---

## 📜 License

MIT © pfrancistechs-cel
