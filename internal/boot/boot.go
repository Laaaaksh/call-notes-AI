package boot

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/call-notes-ai-service/internal/config"
	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/interceptors"
	"github.com/call-notes-ai-service/internal/logger"
	"github.com/call-notes-ai-service/internal/modules/analytics"
	"github.com/call-notes-ai-service/internal/modules/extraction"
	"github.com/call-notes-ai-service/internal/modules/fieldmapper"
	"github.com/call-notes-ai-service/internal/modules/followup"
	"github.com/call-notes-ai-service/internal/modules/health"
	"github.com/call-notes-ai-service/internal/modules/prediction"
	"github.com/call-notes-ai-service/internal/modules/reasoning"
	"github.com/call-notes-ai-service/internal/modules/salesforce"
	"github.com/call-notes-ai-service/internal/modules/sentiment"
	"github.com/call-notes-ai-service/internal/modules/session"
	"github.com/call-notes-ai-service/internal/modules/transcription"
	"github.com/call-notes-ai-service/internal/modules/triage"
	"github.com/call-notes-ai-service/internal/services/deepgram"
	"github.com/call-notes-ai-service/internal/services/llm"
	"github.com/call-notes-ai-service/internal/services/sfdc"
	"github.com/call-notes-ai-service/internal/tracing"
	ws "github.com/call-notes-ai-service/internal/websocket"
	"github.com/call-notes-ai-service/pkg/database"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

// App holds all application dependencies
type App struct {
	Config     *config.Config
	Database   *database.Database
	Redis      *redis.Client
	Modules    *Modules
	Services   *Services
	WSHub      *ws.Hub
	Tracer     *tracing.Tracer
	MainServer *http.Server
	OpsServer  *http.Server
}

// Modules holds all application modules
type Modules struct {
	Health        health.IModule
	Session       session.IModule
	Extraction    extraction.IModule
	Prediction    prediction.IModule
	Sentiment     sentiment.IModule
	Triage        triage.IModule
	FollowUp      followup.IModule
	Analytics     analytics.IModule
	FieldMapper   fieldmapper.IModule
	Transcription transcription.IModule
	Salesforce    salesforce.IModule
	Reasoning     reasoning.IModule
}

// Services holds external service clients
type Services struct {
	Deepgram deepgram.IClient
	LLM      llm.IClient
	SFDC     sfdc.IClient
}

// Initialize creates and configures the application
func Initialize(ctx context.Context) (*App, error) {
	app := &App{}

	if err := app.loadConfig(); err != nil {
		return nil, err
	}
	if err := app.initLogger(); err != nil {
		return nil, err
	}

	app.logStartup()

	if err := app.initTracing(ctx); err != nil {
		return nil, err
	}
	if err := app.initDatabase(ctx); err != nil {
		return nil, err
	}

	app.initRedis(ctx)
	app.initServices()
	app.initWSHub()
	app.initModules(ctx)
	app.setupServers()

	return app, nil
}

func (a *App) loadConfig() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	a.Config = cfg
	return nil
}

func (a *App) initLogger() error {
	return logger.Initialize(a.Config.Logging.Level, a.Config.Logging.Format)
}

func (a *App) logStartup() {
	logger.Info(constants.LogMsgStartingService,
		constants.LogFieldName, a.Config.App.Name,
		constants.LogFieldEnv, a.Config.App.Env,
		constants.LogFieldPort, a.Config.App.Port,
		constants.LogFieldOpsPort, a.Config.App.OpsPort,
	)
}

func (a *App) initTracing(ctx context.Context) error {
	tracer, err := tracing.Initialize(ctx, &a.Config.Tracing, a.Config.App.Name)
	if err != nil {
		logger.Warn(constants.LogMsgTracerInitFailed, constants.LogKeyError, err)
		return nil
	}
	a.Tracer = tracer
	return nil
}

func (a *App) initDatabase(ctx context.Context) error {
	dbCfg := buildDatabaseConfig(&a.Config.Database)

	maxRetries := a.Config.Database.Retry.MaxRetries
	if !a.Config.Database.Retry.Enabled {
		maxRetries = 1
	}

	db, err := database.InitializeWithRetry(ctx, dbCfg, maxRetries, time.Second)
	if err != nil {
		return err
	}

	a.Database = db
	return nil
}

func buildDatabaseConfig(cfg *config.DatabaseConfig) *database.DatabaseConfig {
	return &database.DatabaseConfig{
		Host:            cfg.Host,
		Port:            cfg.Port,
		User:            cfg.User,
		Password:        cfg.Password,
		Name:            cfg.Name,
		SSLMode:         cfg.SSLMode,
		MaxConnections:  cfg.MaxConnections,
		MinConnections:  cfg.MinConnections,
		MaxConnLifetime: cfg.GetMaxConnLifetime(),
		MaxConnIdleTime: cfg.GetMaxConnIdleTime(),
	}
}

func (a *App) initRedis(ctx context.Context) {
	a.Redis = redis.NewClient(&redis.Options{
		Addr:     a.Config.Redis.Addr,
		Password: a.Config.Redis.Password,
		DB:       a.Config.Redis.DB,
	})

	if err := a.Redis.Ping(ctx).Err(); err != nil {
		logger.Warn(constants.LogMsgRedisConnFailed, constants.LogKeyError, err)
		a.Redis = nil
		return
	}
	logger.Info(constants.LogMsgRedisConnected, constants.LogFieldAddr, a.Config.Redis.Addr)
}

func (a *App) initServices() {
	a.Services = &Services{
		Deepgram: deepgram.NewClient(&a.Config.Deepgram),
		LLM:      llm.NewClient(&a.Config.LLM),
		SFDC:     sfdc.NewClient(&a.Config.Salesforce),
	}
}

func (a *App) initWSHub() {
	a.WSHub = ws.NewHub()
	go a.WSHub.Run()
}

func (a *App) initModules(ctx context.Context) {
	pool := a.Database.GetIPool()

	a.Modules = &Modules{
		Health:        health.NewModule(ctx, a.Database),
		Session:       session.NewModule(ctx, pool, a.Redis),
		Extraction:    extraction.NewModule(ctx, a.Services.LLM),
		Prediction:    prediction.NewModule(pool),
		Sentiment:     sentiment.NewModule(pool),
		Triage:        triage.NewModule(pool),
		FollowUp:      followup.NewModule(pool),
		Analytics:     analytics.NewModule(pool),
		FieldMapper:   fieldmapper.NewModule(ctx),
		Transcription: transcription.NewModule(ctx),
		Salesforce:    salesforce.NewModule(ctx, a.Services.SFDC),
		Reasoning:     reasoning.NewModule(ctx, a.Services.LLM),
	}
}

func (a *App) setupServers() {
	mainRouter := a.createMainRouter()
	opsRouter := a.createOpsRouter()

	a.MainServer = &http.Server{Addr: a.Config.App.Port, Handler: mainRouter}
	a.OpsServer = &http.Server{Addr: a.Config.App.OpsPort, Handler: opsRouter}
}

func (a *App) createMainRouter() chi.Router {
	r := chi.NewRouter()

	middlewareCfg := interceptors.MiddlewareConfig{
		RateLimit: a.Config.RateLimit,
		Tracing:   a.Config.Tracing,
	}
	for _, mw := range interceptors.GetChiMiddlewareWithFullConfig(middlewareCfg) {
		r.Use(mw)
	}

	r.Route(constants.APIVersionPrefix, func(r chi.Router) {
		a.Modules.Session.GetHandler().RegisterRoutes(r)
		a.Modules.Prediction.GetHandler().RegisterRoutes(r)
		a.Modules.Triage.GetHandler().RegisterRoutes(r)
		a.Modules.FollowUp.GetHandler().RegisterRoutes(r)
		a.Modules.Analytics.GetHandler().RegisterRoutes(r)
	})

	return r
}

func (a *App) createOpsRouter() chi.Router {
	r := chi.NewRouter()
	healthHandler := health.NewHTTPHandler(a.Modules.Health.GetCore())
	healthHandler.RegisterRoutes(r)
	r.Handle(constants.RouteMetrics, promhttp.Handler())
	return r
}

// Start begins listening on both servers
func (a *App) Start() {
	go a.startMainServer()
	go a.startOpsServer()
}

func (a *App) startMainServer() {
	logger.Info(constants.LogMsgMainServerStarting, constants.LogFieldAddr, a.MainServer.Addr)
	if err := a.MainServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal(constants.LogMsgMainServerFailed, constants.LogKeyError, err)
	}
}

func (a *App) startOpsServer() {
	logger.Info(constants.LogMsgOpsServerStarting, constants.LogFieldAddr, a.OpsServer.Addr)
	if err := a.OpsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal(constants.LogMsgOpsServerFailed, constants.LogKeyError, err)
	}
}

// Shutdown gracefully stops the application
func (a *App) Shutdown(ctx context.Context) {
	logger.Info(constants.LogMsgShutdownSignalReceived)

	a.Modules.Health.GetCore().MarkUnhealthy()

	delay := a.Config.App.ShutdownDelay
	logger.Info(constants.LogMsgWaitingForShutdownDelay, constants.LogFieldDelaySeconds, delay)
	time.Sleep(time.Duration(delay) * time.Second)

	shutdownCtx, cancel := context.WithTimeout(ctx, time.Duration(a.Config.App.ShutdownTimeout)*time.Second)
	defer cancel()

	a.shutdownServers(shutdownCtx)
	a.shutdownTracer(shutdownCtx)
	a.shutdownConnections()

	logger.Info(constants.LogMsgServiceStopped)
}

func (a *App) shutdownServers(ctx context.Context) {
	if err := a.MainServer.Shutdown(ctx); err != nil {
		logger.Error(constants.LogMsgMainServerShutdownErr, constants.LogKeyError, err)
	}
	if err := a.OpsServer.Shutdown(ctx); err != nil {
		logger.Error(constants.LogMsgOpsServerShutdownErr, constants.LogKeyError, err)
	}
}

func (a *App) shutdownTracer(ctx context.Context) {
	if a.Tracer != nil {
		a.Tracer.Shutdown(ctx)
	}
}

func (a *App) shutdownConnections() {
	if a.Redis != nil {
		_ = a.Redis.Close()
	}
	a.Database.Close()
}
