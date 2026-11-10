package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pfrancistechs-cel/minibroker/backend/internal/broker"
	"github.com/pfrancistechs-cel/minibroker/backend/internal/domain"
)

// Handler serves HTTP & SSE endpoints for MiniBroker.
type Handler struct {
	broker *broker.Broker
}

// NewHandler initializes a new handler.
func NewHandler(b *broker.Broker) *Handler {
	return &Handler{broker: b}
}

// RegisterRoutes attaches endpoints to the provided ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.HandleHealth)
	mux.HandleFunc("GET /api/topics", h.HandleListTopics)
	mux.HandleFunc("POST /api/topics", h.HandleCreateTopic)
	mux.HandleFunc("GET /api/topics/{topic}", h.HandleGetTopic)
	mux.HandleFunc("POST /api/topics/{topic}/publish", h.HandlePublish)
	mux.HandleFunc("GET /api/topics/{topic}/partitions/{partition}/messages", h.HandleReadMessages)
	mux.HandleFunc("GET /api/topics/{topic}/stream", h.HandleTopicSSEStream)
	mux.HandleFunc("GET /api/consumer-groups", h.HandleGetConsumerGroups)
	mux.HandleFunc("POST /api/consumer-groups/{group}/offsets", h.HandleCommitOffset)
	mux.HandleFunc("GET /api/metrics", h.HandleGetMetrics)
	mux.HandleFunc("GET /api/metrics/stream", h.HandleMetricsSSEStream)
}

func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"cluster": "minibroker-core",
		"engine":  "in-memory-wal+neon",
	})
}

func (h *Handler) HandleListTopics(w http.ResponseWriter, r *http.Request) {
	topics := h.broker.GetTopics()
	respondJSON(w, http.StatusOK, topics)
}

func (h *Handler) HandleCreateTopic(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name       string `json:"name"`
		Partitions int    `json:"partitions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payload: " + err.Error()})
		return
	}

	if input.Name == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "topic name is required"})
		return
	}

	if input.Partitions <= 0 {
		input.Partitions = 3
	}

	if err := h.broker.CreateTopic(r.Context(), input.Name, input.Partitions); err != nil {
		if err == broker.ErrTopicExists {
			respondJSON(w, http.StatusConflict, map[string]string{"error": "topic already exists"})
			return
		}
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	detail, _ := h.broker.GetTopicDetail(input.Name)
	respondJSON(w, http.StatusCreated, detail)
}

func (h *Handler) HandleGetTopic(w http.ResponseWriter, r *http.Request) {
	topic := r.PathValue("topic")
	detail, err := h.broker.GetTopicDetail(topic)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "topic not found"})
		return
	}
	respondJSON(w, http.StatusOK, detail)
}

func (h *Handler) HandlePublish(w http.ResponseWriter, r *http.Request) {
	topic := r.PathValue("topic")
	var input domain.PublishInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body: " + err.Error()})
		return
	}

	if input.Value == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "value is required"})
		return
	}

	res, err := h.broker.Publish(r.Context(), topic, input)
	if err != nil {
		if err == broker.ErrTopicNotFound {
			respondJSON(w, http.StatusNotFound, map[string]string{"error": "topic not found"})
			return
		}
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusCreated, res)
}

func (h *Handler) HandleReadMessages(w http.ResponseWriter, r *http.Request) {
	topic := r.PathValue("topic")
	partStr := r.PathValue("partition")
	partition, err := strconv.Atoi(partStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid partition number"})
		return
	}

	offset := uint64(0)
	if offStr := r.URL.Query().Get("offset"); offStr != "" {
		if o, err := strconv.ParseUint(offStr, 10, 64); err == nil {
			offset = o
		}
	}

	limit := 50
	if limStr := r.URL.Query().Get("limit"); limStr != "" {
		if l, err := strconv.Atoi(limStr); err == nil && l > 0 {
			limit = l
		}
	}

	msgs, err := h.broker.ReadMessages(topic, partition, offset, limit)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, msgs)
}

func (h *Handler) HandleTopicSSEStream(w http.ResponseWriter, r *http.Request) {
	topic := r.PathValue("topic")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := h.broker.Subscribe(topic)
	defer h.broker.Unsubscribe(topic, ch)

	// Send initial connected ping
	fmt.Fprintf(w, "event: connected\ndata: {\"topic\":\"%s\",\"status\":\"streaming\"}\n\n", topic)
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			dataBytes, _ := json.Marshal(msg)
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", string(dataBytes))
			flusher.Flush()
		}
	}
}

func (h *Handler) HandleGetConsumerGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := h.broker.GetConsumerGroups(r.Context())
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	respondJSON(w, http.StatusOK, groups)
}

func (h *Handler) HandleCommitOffset(w http.ResponseWriter, r *http.Request) {
	group := r.PathValue("group")
	var input struct {
		Topic     string `json:"topic"`
		Partition int    `json:"partition"`
		Offset    uint64 `json:"offset"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payload: " + err.Error()})
		return
	}

	if err := h.broker.CommitOffset(r.Context(), group, input.Topic, input.Partition, input.Offset); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"status": "committed", "group": group, "offset": input.Offset})
}

func (h *Handler) HandleGetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := h.broker.GetMetrics()
	respondJSON(w, http.StatusOK, metrics)
}

func (h *Handler) HandleMetricsSSEStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			m := h.broker.GetMetrics()
			dataBytes, _ := json.Marshal(m)
			fmt.Fprintf(w, "data: %s\n\n", string(dataBytes))
			flusher.Flush()
		}
	}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
