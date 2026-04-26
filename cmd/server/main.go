package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bubbletrack/server/internal/api"
	"github.com/bubbletrack/server/internal/application"
	"github.com/bubbletrack/server/internal/config"
	"github.com/bubbletrack/server/internal/domain"
	"github.com/bubbletrack/server/internal/infrastructure/adk"
	"github.com/bubbletrack/server/internal/infrastructure/ollama"
	"github.com/bubbletrack/server/internal/infrastructure/pubsub"
	"github.com/bubbletrack/server/internal/infrastructure/queue"
	"github.com/bubbletrack/server/internal/infrastructure/repository"
	"github.com/bubbletrack/server/internal/logger"
	"github.com/bubbletrack/server/internal/websocket"
	pb "github.com/qdrant/go-client/qdrant"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.Load()
	log := logger.New(cfg.Server.Env)
	slog.SetDefault(log)

	log.Info("starting bubble-track-server", "env", cfg.Server.Env, "port", cfg.Server.Port)

	pool, err := repository.NewPostgresPool(ctx, cfg.Postgres)
	if err != nil {
		log.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := repository.RunMigrations(ctx, pool); err != nil {
		log.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	log.Info("database migrations complete")

	if err := repository.AddNodeMetricsMigrations(ctx, pool); err != nil {
		log.Error("failed to run node_metrics migrations", "error", err)
		os.Exit(1)
	}

	personRepo := repository.NewPostgresPersonRepository(pool)
	relationshipRepo := repository.NewPostgresRelationshipRepository(pool)
	interactionRepo := repository.NewPostgresInteractionRepository(pool)
	graphRepo := repository.NewPostgresGraphRepository(pool)
	graphEngine := repository.NewGraphEngine(pool)
	chatRepo := repository.NewChatMessageRepository(pool)
	stateRepo := repository.NewPersonStateRepository(pool)
	userRepo := repository.NewPostgresUserRepository(pool)
	refreshTokenRepo := repository.NewPostgresRefreshTokenRepository(pool)

	wsHub := websocket.NewHub()
	go wsHub.Run(ctx)

	wsUpgrader := websocket.NewUpgrader(wsHub)

	qdrantClient, err := connectQdrant(*cfg)
	if err != nil {
		log.Error("failed to connect to qdrant", "error", err)
		os.Exit(1)
	}
	defer qdrantClient.Close()

	qdrantRepo := repository.NewQdrantMemoryRepository(qdrantClient, cfg.Qdrant)
	if err := qdrantRepo.EnsureCollection(ctx); err != nil {
		log.Error("failed to ensure qdrant collection", "error", err)
		os.Exit(1)
	}
	log.Info("qdrant connected")

	useOllama := os.Getenv("USE_OLLAMA") == "true"

	var embedder domain.Embedder
	var btAgent interface {
		ProcessInteraction(ctx context.Context, userID, text string) (*domain.AnalysisResult, error)
	}

	if useOllama {
		ollamaClient, err := ollama.NewClientWithEmbedding(cfg.Ollama.URL(), cfg.Ollama.Model, cfg.Ollama.EmbeddingModel)
		if err != nil {
			log.Error("failed to create ollama client", "error", err)
			os.Exit(1)
		}
		defer ollamaClient.Close()
		embedder = ollamaClient
		log.Info("ollama client initialized", "model", cfg.Ollama.Model, "embedding", cfg.Ollama.EmbeddingModel)

		ollamaAgent := adk.NewOllamaAgent(cfg.Ollama.URL(), cfg.Ollama.Model, log)

		ollamaAgent.RegisterToolHandler("search_memory", func(args map[string]any) (map[string]any, error) {
			limit := 5
			if v, ok := args["limit"].(float64); ok {
				limit = int(v)
			}
			query, _ := args["query"].(string)
			embedding, err := embedder.Embed(ctx, query)
			if err != nil {
				return nil, err
			}
			results, err := qdrantRepo.Search(ctx, cfg.Server.DefaultUserID, embedding, limit)
			if err != nil {
				return nil, err
			}
			items := make([]map[string]any, 0, len(results))
			for _, r := range results {
				rawText := ""
				if v, ok := r.Metadata["raw_text"].(string); ok {
					rawText = v
				}
				items = append(items, map[string]any{"content": rawText, "score": r.Score})
			}
			return map[string]any{"memories": items, "count": len(items)}, nil
		})

		ollamaAgent.RegisterToolHandler("store_memory", func(args map[string]any) (map[string]any, error) {
			content, _ := args["content"].(string)
			embedding, err := embedder.Embed(ctx, content)
			if err != nil {
				return nil, err
			}
			metadata := map[string]any{"raw_text": content, "timestamp": time.Now().UTC().Format(time.RFC3339)}
			if people, ok := args["people"].([]interface{}); ok {
				names := make([]string, 0, len(people))
				for _, p := range people {
					if s, ok := p.(string); ok {
						names = append(names, s)
					}
				}
				peopleJSON, _ := json.Marshal(names)
				metadata["people"] = string(peopleJSON)
			}
			if err := qdrantRepo.Store(ctx, cfg.Server.DefaultUserID, fmt.Sprintf("tool-%d", time.Now().UnixNano()), embedding, metadata); err != nil {
				return nil, err
			}
			return map[string]any{"status": "stored"}, nil
		})

		ollamaAgent.RegisterToolHandler("update_graph", func(args map[string]any) (map[string]any, error) {
			sourceName, _ := args["source_person"].(string)
			targetName, _ := args["target_person"].(string)
			qualityStr, _ := args["quality"].(string)
			strength := 0.5
			if v, ok := args["strength"].(float64); ok {
				strength = v
			}
			label, _ := args["label"].(string)

			if sourceName == "" || targetName == "" {
				return map[string]any{"status": "error", "message": "source and target names required"}, nil
			}

			quality := domain.QualityNeutral
			switch qualityStr {
			case "nourishing":
				quality = domain.QualityNourishing
			case "draining":
				quality = domain.QualityDraining
			case "conflicted":
				quality = domain.QualityConflicted
			}

			sourcePerson, err := personRepo.GetOrCreateByName(ctx, sourceName)
			if err != nil {
				log.Warn("update_graph: failed to get/create source", "name", sourceName, "error", err)
				return map[string]any{"status": "error", "message": err.Error()}, nil
			}
			targetPerson, err := personRepo.GetOrCreateByName(ctx, targetName)
			if err != nil {
				log.Warn("update_graph: failed to get/create target", "name", targetName, "error", err)
				return map[string]any{"status": "error", "message": err.Error()}, nil
			}

			rel := &domain.Relacionamento{
				ID:               uuid.New().String(),
				SourcePersonID:   sourcePerson.ID,
				TargetPersonID:   targetPerson.ID,
				Quality:          quality,
				Strength:         strength,
				Label:            label,
				ReciprocityIndex: 0.5,
			}
			if err := relationshipRepo.Upsert(ctx, rel); err != nil {
				log.Warn("update_graph: failed to upsert relationship", "error", err)
				return map[string]any{"status": "error", "message": err.Error()}, nil
			}

			return map[string]any{"status": "recorded", "source": sourceName, "target": targetName, "strength": strength}, nil
		})

		ollamaAgent.RegisterToolHandler("classify_role", func(args map[string]any) (map[string]any, error) {
			return map[string]any{"status": "classified", "person": args["person"], "role": args["role"]}, nil
		})

		ollamaAgent.RegisterToolHandler("set_user_preferences", func(args map[string]any) (map[string]any, error) {
			return map[string]any{"status": "saved", "philosophical_lens": args["philosophical_lens"]}, nil
		})

		ollamaAgent.RegisterToolHandler("record_emotional_state", func(args map[string]any) (map[string]any, error) {
			return map[string]any{"status": "recorded", "person": args["person_name"], "mood": args["mood"]}, nil
		})

		ollamaAgent.RegisterToolHandler("update_relationship_protocol", func(args map[string]any) (map[string]any, error) {
			return map[string]any{"status": "recorded", "source": args["source_person"], "target": args["target_person"], "protocol": args["protocol"]}, nil
		})

		btAgent = ollamaAgent
		log.Info("ollama agent initialized with tool calling", "model", cfg.Ollama.Model)
	}

	if cfg.GenAI.APIKey != "" {
		tools, err := adk.CreateBubbleTrackTools(qdrantRepo, embedder, personRepo, relationshipRepo, log)
		if err != nil {
			log.Error("failed to create tools", "error", err)
			os.Exit(1)
		}

		adkAgent, err := adk.NewBubbleTrackAgent(ctx, cfg.GenAI.APIKey, cfg.GenAI.Model, tools, log)
		if err != nil {
			log.Error("failed to create ADK agent", "error", err)
			os.Exit(1)
		}
		btAgent = adkAgent
		log.Info("ADK agent initialized with Gemini", "model", cfg.GenAI.Model)
	}

	if btAgent == nil {
		log.Error("no AI client available (need either GEMINI_API_KEY or USE_OLLAMA=true)")
		os.Exit(1)
	}

	notifier := pubsub.NewRedisNotifier(cfg.Redis, log)
	defer notifier.Close()

	asynqClient, err := queue.NewClient(cfg.Redis)
	if err != nil {
		log.Error("failed to create asynq client", "error", err)
		os.Exit(1)
	}
	defer asynqClient.Close()

	enqueuer := queue.NewEnqueuer(asynqClient)

	analyzeUC := application.NewAnalyzeUseCase(
		btAgent, interactionRepo, personRepo, relationshipRepo,
		qdrantRepo, embedder, notifier, enqueuer,
		application.NewChatMessageRepoAdapter(chatRepo),
		wsHub,
		stateRepo,
		log,
	)
	getGraphUC := application.NewGetGraphUseCase(*graphRepo)
	graphAnalysisEngine := application.NewGraphAnalysisEngine(graphEngine, personRepo, relationshipRepo, log)
	classifyUC := application.NewClassificationEngine()
	aggregationUC := application.NewAggregationEngine(graphEngine)

	authUC := application.NewAuthUseCase(userRepo, refreshTokenRepo, cfg.Server.JWTSecret)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.HTTPErrorHandler = api.ErrorHandler()
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogMethod:   true,
		LogError:    true,
		LogLatency:  true,
		LogRemoteIP: true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log.Info("http_request",
				"method", v.Method,
				"uri", v.URI,
				"status", v.Status,
				"latency", v.Latency,
				"remote_ip", v.RemoteIP,
				"error", v.Error,
			)
			return nil
		},
	}))
	e.Use(api.JWTMiddleware(authUC))

	handler := api.NewHandler(analyzeUC, getGraphUC, graphAnalysisEngine, graphEngine, classifyUC, aggregationUC, notifier, embedder, qdrantRepo, chatRepo, wsHub, stateRepo, log)
	authHandler := api.NewAuthHandler(authUC)
	authHandler.RegisterRoutes(e)
	handler.RegisterRoutes(e)
	wsUpgrader.RegisterRoutes(e)

	go startWorker(ctx, *cfg, analyzeUC, log)

	go func() {
		addr := ":" + cfg.Server.Port
		log.Info("http server listening", "addr", addr)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown error", "error", err)
	}

	log.Info("server stopped")
}

func startWorker(ctx context.Context, cfg config.Config, analyzeUC *application.AnalyzeUseCase, log *slog.Logger) {
	srv := queue.NewServer(cfg.Redis, log)
	mux := queue.NewServeMux()
	application.RegisterWorker(mux, analyzeUC, log)

	log.Info("worker started")
	if err := srv.Run(mux); err != nil {
		log.Error("worker error", "error", err)
	}
	_ = ctx
}

func connectQdrant(cfg config.Config) (*pb.Client, error) {
	return pb.NewClient(&pb.Config{
		Host:   cfg.Qdrant.Host,
		Port:   cfg.Qdrant.Port,
		APIKey: cfg.Qdrant.APIKey,
		UseTLS: cfg.Qdrant.UseTLS,
	})
}
