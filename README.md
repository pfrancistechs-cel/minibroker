# MiniBroker 📬 ⚡

> **A lightweight, high-throughput distributed message broker with consumer groups, topics, partitions, and commit-log persistence built in Go & Next.js.**

MiniBroker provides a modern Kafka/NATS-lite pub-sub model featuring durable write-ahead logging, zero-copy message delivery over TCP, consumer group offset tracking, and a live web telemetry dashboard.

---

## 🏗️ Architecture

```text
                     ┌───────────────────────────┐
                     │ MiniBroker Next.js Console│
                     │ (Topics, Lags & Telemetry)│
                     └─────────────┬─────────────┘
                                   │ HTTP/WebSocket Metrics
                                   ▼
Producers ───▶ ┌──────────────────────────────────────────────┐ ───▶ Consumers
               │              MiniBroker Core                 │
               │  • Topic & Partition Routing (Ring Buffer)   │
               │  • Consumer Group Offset Management          │
               │  • Append-Only Commit Log Persistence (WAL)  │
               └──────────────────────────────────────────────┘
```

---

## 📂 Project Structure

```text
minibroker/
├── backend/                  # Go message broker core
│   ├── cmd/
│   │   ├── broker/           # Broker TCP & API daemon
│   │   └── cli/              # Producer/Consumer CLI client
│   ├── pkg/
│   │   ├── api/              # Management & metrics REST API
│   │   ├── broker/           # Topics, partitions, consumer groups
│   │   ├── protocol/         # Framing & wire protocol encoder/decoder
│   │   └── storage/          # Append-only segment files & index
│   └── go.mod
│
├── frontend/                 # Next.js broker dashboard
│   ├── public/
│   └── src/
│       ├── app/
│       ├── components/
│       ├── hooks/
│       ├── lib/
│       └── types/
│
├── .gitignore
└── README.md
```

---

## 📜 License
MIT © pfrancistechs-cel
