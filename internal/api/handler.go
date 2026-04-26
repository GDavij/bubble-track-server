package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/bubbletrack/server/internal/application"
	"github.com/bubbletrack/server/internal/domain"
	"github.com/bubbletrack/server/internal/infrastructure/repository"
	"github.com/bubbletrack/server/internal/websocket"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	analyzeUC     *application.AnalyzeUseCase
	getGraphUC    *application.GetGraphUseCase
	graphEngineUC *application.GraphAnalysisEngine
	graphRepo     *repository.GraphEngine
	classifyUC    *application.ClassificationEngine
	aggregationUC *application.AggregationEngine
	notifier      SSENotifier
	embedder      domain.Embedder
	memoryRepo    domain.MemoryRepository
	chatRepo      *repository.ChatMessageRepository
	wsHub         *websocket.Hub
	stateRepo    *repository.PersonStateRepository
	logger        *slog.Logger
}

type SSENotifier interface {
	Subscribe(ctx context.Context, userID string) <-chan *domain.AnalysisResult
}

func NewHandler(
	analyzeUC *application.AnalyzeUseCase,
	getGraphUC *application.GetGraphUseCase,
	graphEngineUC *application.GraphAnalysisEngine,
	graphRepo *repository.GraphEngine,
	classifyUC *application.ClassificationEngine,
	aggregationUC *application.AggregationEngine,
	notifier SSENotifier,
	embedder domain.Embedder,
	memoryRepo domain.MemoryRepository,
	chatRepo *repository.ChatMessageRepository,
	wsHub *websocket.Hub,
	stateRepo *repository.PersonStateRepository,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		analyzeUC:     analyzeUC,
		getGraphUC:    getGraphUC,
		graphEngineUC: graphEngineUC,
		graphRepo:     graphRepo,
		classifyUC:    classifyUC,
		aggregationUC: aggregationUC,
		notifier:      notifier,
		embedder:      embedder,
		memoryRepo:    memoryRepo,
		chatRepo:      chatRepo,
		wsHub:         wsHub,
		stateRepo:    stateRepo,
		logger:        logger,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	api := e.Group("/api")
	api.POST("/analyze", h.Analyze)
	api.GET("/graph", h.GetGraph)
	api.GET("/graph/full", h.GetFullGraph)
	api.GET("/health", h.Health)
	api.GET("/analysis/stream", h.AnalysisStream)
	api.GET("/memories", h.GetMemories)
	api.POST("/chat", h.SendChatMessage)
	api.GET("/chat", h.GetChatMessages)
	api.GET("/states", h.GetEmotionalTimeline)
	api.GET("/states/self", h.GetSelfStates)

	people := api.Group("/people")
	people.GET("", h.ListPeople)
	people.GET("/:id", h.GetPersonDetail)
	people.GET("/:id/metrics", h.GetPersonMetrics)
	people.POST("/:id/classify", h.ClassifySocialRole)

	analysis := api.Group("/analysis")
	analysis.GET("/roles", h.GetAllRoles)
	analysis.GET("/profiles", h.GetAllProfiles)
	analysis.GET("/graph/snapshot", h.GetGraphSnapshot)

	relationships := api.Group("/relationships")
	relationships.GET("/:id/health", h.GetRelationshipHealth)
}

type analyzeRequest struct {
	Text string `json:"text"`
}

func (h *Handler) Analyze(c echo.Context) error {
	userID := c.Get("user_id").(string)
	var req analyzeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Text == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "text is required"})
	}
	resp, err := h.analyzeUC.Submit(c.Request().Context(), application.AnalyzeRequest{
		UserID:  userID,
		RawText: req.Text,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusAccepted, resp)
}

func (h *Handler) GetGraph(c echo.Context) error {
	userID := c.Get("user_id").(string)
	graph, err := h.getGraphUC.Execute(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch graph"})
	}
	return c.JSON(http.StatusOK, graph)
}

func (h *Handler) GetMemories(c echo.Context) error {
	userID := c.Get("user_id").(string)
	segment := c.QueryParam("segment")
	session := c.QueryParam("session")
	person := c.QueryParam("person")
	query := c.QueryParam("query")
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 10
	}
	userCtx := c.Request().Context()
	filter := domain.MemoryFilter{
		Segment: segment,
		Session: session,
		Person:  person,
	}
	if h.embedder != nil && h.memoryRepo != nil && query != "" {
		embedding, err := h.embedder.Embed(userCtx, query)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid query"})
		}
		results, err := h.memoryRepo.SearchFiltered(userCtx, userID, embedding, limit, filter)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "search failed"})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"memories": results})
	}
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "RAG not configured"})
}

func (h *Handler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) ListPeople(c echo.Context) error {
	userID := c.Get("user_id").(string)
	graph, err := h.getGraphUC.Execute(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, graph.Nodes)
}

func (h *Handler) GetPersonDetail(c echo.Context) error {
	userID := c.Get("user_id").(string)
	personID := c.Param("id")
	allMetrics, err := h.graphEngineUC.ComputeAllMetrics(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to compute metrics"})
	}
	m, ok := allMetrics[personID]
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "person not found"})
	}
	classEngine := application.NewClassificationEngine()
	classification := classEngine.ClassifyRole(m)
	aggEngine := application.NewAggregationEngine(h.graphRepo)
	profile := aggEngine.AggregateProfile(m, allMetrics)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"person":         m,
		"classification": classification,
		"profile":        profile,
	})
}

func (h *Handler) GetPersonMetrics(c echo.Context) error {
	userID := c.Get("user_id").(string)
	personID := c.Param("id")
	allMetrics, err := h.graphEngineUC.ComputeAllMetrics(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to compute metrics"})
	}
	m, ok := allMetrics[personID]
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "person not found"})
	}
	return c.JSON(http.StatusOK, m)
}

func (h *Handler) GetAllRoles(c echo.Context) error {
	userID := c.Get("user_id").(string)
	allMetrics, err := h.graphEngineUC.ComputeAllMetrics(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to compute metrics"})
	}
	classEngine := application.NewClassificationEngine()
	roles := make(map[string]interface{})
	for id, metrics := range allMetrics {
		roles[id] = classEngine.ClassifyRole(metrics)
	}
	return c.JSON(http.StatusOK, roles)
}

func (h *Handler) GetAllProfiles(c echo.Context) error {
	userID := c.Get("user_id").(string)
	allMetrics, err := h.graphEngineUC.ComputeAllMetrics(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to compute metrics"})
	}
	aggEngine := application.NewAggregationEngine(h.graphRepo)
	profiles := make(map[string]interface{})
	for id, metrics := range allMetrics {
		profiles[id] = aggEngine.AggregateProfile(metrics, allMetrics)
	}
	return c.JSON(http.StatusOK, profiles)
}

func (h *Handler) GetGraphSnapshot(c echo.Context) error {
	userID := c.Get("user_id").(string)
	graph, err := h.getGraphUC.Execute(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, graph.Stats)
}

func (h *Handler) GetRelationshipHealth(c echo.Context) error {
	userID := c.Get("user_id").(string)
	relID := c.Param("id")
	graph, err := h.getGraphUC.Execute(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch graph"})
	}
	for _, edge := range graph.Edges {
		if edge.Source == relID || edge.Target == relID {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"source":   edge.Source,
				"target":   edge.Target,
				"quality":  edge.Quality,
				"strength": edge.Strength,
			})
		}
	}
	return c.JSON(http.StatusNotFound, map[string]string{"error": "relationship not found"})
}

func (h *Handler) ClassifySocialRole(c echo.Context) error {
	userID := c.Get("user_id").(string)
	personID := c.Param("id")
	if personID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "person_id required"})
	}
	allMetrics, err := h.graphEngineUC.ComputeAllMetrics(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to compute metrics"})
	}
	m, ok := allMetrics[personID]
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "person not found"})
	}
	role := h.classifyUC.ClassifyRole(m)
	return c.JSON(http.StatusOK, map[string]interface{}{"person_id": personID, "role": role})
}

func (h *Handler) AnalysisStream(c echo.Context) error {
	userID := c.Get("user_id").(string)
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().WriteHeader(http.StatusOK)
	if h.notifier == nil {
		return nil
	}
	events := h.notifier.Subscribe(c.Request().Context(), userID)
	for {
		select {
		case <-c.Request().Context().Done():
			return nil
		case event := <-events:
			data, _ := json.Marshal(event)
			c.Response().Write([]byte("data: " + string(data) + "\n\n"))
			c.Response().Flush()
		}
	}
}

func (h *Handler) SendChatMessage(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req struct {
		Sender    string `json:"sender"`
		Content   string `json:"content"`
		SessionID string `json:"session_id"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if req.Content == "" {
		h.logWarn(c, "chat_send_rejected", map[string]interface{}{"reason": "empty_content", "user_id": userID})
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "content is required"})
	}

	msg := &repository.ChatMessage{
		UserID:    userID,
		Sender:    req.Sender,
		Content:   req.Content,
		IsUser:    true,
		SessionID: req.SessionID,
	}

	if err := h.chatRepo.Create(c.Request().Context(), msg); err != nil {
		h.logWarn(c, "chat_send_failed", map[string]interface{}{"user_id": userID, "error": err.Error()})
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save message"})
	}

	if h.wsHub != nil {
		wsMsg := websocket.ChatMessage{
			ID:        msg.ID.String(),
			UserID:    msg.UserID,
			Sender:    msg.Sender,
			Content:   msg.Content,
			IsUser:    msg.IsUser,
			CreatedAt: msg.CreatedAt.Format(time.RFC3339Nano),
		}
		if err := h.wsHub.BroadcastMessage(wsMsg); err != nil {
			h.logWarn(c, "chat_ws_broadcast_failed", map[string]interface{}{"user_id": userID, "error": err.Error()})
		}
	}

	var analyzeResp *application.AnalyzeResponse
	if h.analyzeUC != nil {
		resp, err := h.analyzeUC.Submit(c.Request().Context(), application.AnalyzeRequest{
			UserID:  userID,
			RawText: req.Content,
		})
		if err != nil {
			h.logWarn(c, "chat_analyze_enqueue_failed", map[string]interface{}{"user_id": userID, "error": err.Error()})
		} else {
			analyzeResp = resp
			h.logInfo(c, "chat_analyze_enqueued", map[string]interface{}{"user_id": userID, "interaction_id": resp.InteractionID, "job_id": resp.JobID})
		}
	}

	h.logInfo(c, "chat_send_ok", map[string]interface{}{"user_id": userID, "sender": req.Sender, "content_len": len(req.Content)})

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message":      msg,
		"analysis_job": analyzeResp,
	})
}

func (h *Handler) GetChatMessages(c echo.Context) error {
	userID := c.Get("user_id").(string)
	sessionID := c.QueryParam("session_id")
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 50
	}

	var messages []repository.ChatMessage
	var err error

	if sessionID != "" {
		messages, err = h.chatRepo.GetBySessionID(c.Request().Context(), sessionID, limit)
	} else {
		messages, err = h.chatRepo.GetByUserID(c.Request().Context(), userID, limit)
	}

	if err != nil {
		h.logWarn(c, "chat_get_failed", map[string]interface{}{"user_id": userID, "error": err.Error(), "session_id": sessionID})
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch messages"})
	}

	h.logInfo(c, "chat_get_ok", map[string]interface{}{"user_id": userID, "session_id": sessionID, "count": len(messages), "limit": limit})

	return c.JSON(http.StatusOK, map[string]interface{}{"messages": messages})
}

func (h *Handler) logInfo(c echo.Context, event string, fields map[string]interface{}) {
	if h == nil || h.logger == nil {
		return
	}
	kv := []interface{}{"event", event, "method", c.Request().Method, "path", c.Path(), "request_id", c.Response().Header().Get(echo.HeaderXRequestID)}
	for k, v := range fields {
		kv = append(kv, k, v)
	}
	h.logger.Info("api_event", kv...)
}

func (h *Handler) logWarn(c echo.Context, event string, fields map[string]interface{}) {
	if h == nil || h.logger == nil {
		return
	}
	kv := []interface{}{"event", event, "method", c.Request().Method, "path", c.Path(), "request_id", c.Response().Header().Get(echo.HeaderXRequestID)}
	for k, v := range fields {
		kv = append(kv, k, v)
	}
	h.logger.Warn("api_event", kv...)
}

func (h *Handler) GetEmotionalTimeline(c echo.Context) error {
	if h.stateRepo == nil {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "states not configured"})
	}
	userID := c.Get("user_id").(string)
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	states, err := h.stateRepo.GetTimeline(c.Request().Context(), userID, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"states": states, "count": len(states)})
}

func (h *Handler) GetSelfStates(c echo.Context) error {
	if h.stateRepo == nil {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "states not configured"})
	}
	userID := c.Get("user_id").(string)
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	states, err := h.stateRepo.GetSelfStates(c.Request().Context(), userID, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"states": states, "count": len(states)})
}

func (h *Handler) GetFullGraph(c echo.Context) error {
	userID := c.Get("user_id").(string)
	ctx := c.Request().Context()

	graph, err := h.getGraphUC.Execute(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch graph"})
	}

	allMetrics, _ := h.graphEngineUC.ComputeAllMetrics(ctx, userID)

	timeline := make([]interface{}, 0)
	if h.stateRepo != nil {
		states, err := h.stateRepo.GetTimeline(ctx, userID, 50)
		if err == nil && len(states) > 0 {
			for _, s := range states {
				timeline = append(timeline, s)
			}
		}
	}

	classEngine := application.NewClassificationEngine()
	roles := make(map[string]interface{})
	for id, m := range allMetrics {
		roles[id] = classEngine.ClassifyRole(m)
	}

	aggEngine := application.NewAggregationEngine(h.graphRepo)
	profiles := make(map[string]interface{})
	for id, m := range allMetrics {
		profiles[id] = aggEngine.AggregateProfile(m, allMetrics)
	}

	personStates := make(map[string]interface{})
	if h.stateRepo != nil {
		for _, node := range graph.Nodes {
			states, err := h.stateRepo.GetByPerson(ctx, userID, node.ID, 10)
			if err == nil && len(states) > 0 {
				personStates[node.ID] = states
			}
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"graph":         graph,
		"metrics":       allMetrics,
		"roles":         roles,
		"profiles":      profiles,
		"person_states":  personStates,
		"timeline":      timeline,
	})
}
