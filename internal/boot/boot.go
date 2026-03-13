package boot

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/call-notes-ai-service/internal/config"
	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/logger"
	"github.com/call-notes-ai-service/internal/modules/analytics"
	"github.com/call-notes-ai-service/internal/modules/extraction"
	"github.com/call-notes-ai-service/internal/modules/followup"
	"github.com/call-notes-ai-service/internal/modules/health"
	"github.com/call-notes-ai-service/internal/modules/prediction"
	"github.com/call-notes-ai-service/internal/modules/sentiment"
	"github.com/call-notes-ai-service/internal/modules/session"
	"github.com/call-notes-ai-service/internal/modules/triage"
	"github.com/call-notes-ai-service/internal/services/deepgram"
	"github.com/call-notes-ai-service/internal/services/llm"
	"github.com/call-notes-ai-service/internal/services/sfdc"
	ws "github.com/call-notes-ai-service/internal/websocket"
	"github.com/call-notes-ai-service/pkg/database"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

type App struct {
	Config     *config.Config
	Database   *database.Database
	Redis      *redis.Client
	Modules    *Modules
	Services   *Services
	WSHub      *ws.Hub
	MainServer *http.Server
	OpsServer  *http.Server
}

type Modules struct {
	Health     health.IModule
	Session    session.IModule
	Extraction extraction.IModule
	Prediction prediction.IModule
	Sentiment  sentiment.IModule
	Triage     triage.IModule
	FollowUp   followup.IModule
	Analytics  analytics.IModule
}

type Services struct {
	Deepgram deepgram.IClient
	LLM      llm.IClient
	SFDC     sfdc.IClient
}

func Initialize(ctx context.Context) (*App, error) {
	app := &App{}

	if err := app.loadConfig(); err != nil {
		return nil, err
	}
	if err := app.initLogger(); err != nil {
		return nil, err
	}

	app.logStartup()

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

func (a *App) initDatabase(ctx context.Context) error {
	dbCfg := &database.DatabaseConfig{
		Host:            a.Config.Database.Host,
		Port:            a.Config.Database.Port,
		User:            a.Config.Database.User,
		Password:        a.Config.Database.Password,
		Name:            a.Config.Database.Name,
		SSLMode:         a.Config.Database.SSLMode,
		MaxConnections:  a.Config.Database.MaxConnections,
		MinConnections:  a.Config.Database.MinConnections,
		MaxConnLifetime: a.Config.Database.GetMaxConnLifetime(),
		MaxConnIdleTime: a.Config.Database.GetMaxConnIdleTime(),
	}

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

func (a *App) initRedis(ctx context.Context) {
	a.Redis = redis.NewClient(&redis.Options{
		Addr:     a.Config.Redis.Addr,
		Password: a.Config.Redis.Password,
		DB:       a.Config.Redis.DB,
	})

	if err := a.Redis.Ping(ctx).Err(); err != nil {
		logger.Warn(constants.LogMsgRedisConnFailed, constants.LogKeyError, err)
		a.Redis = nil
	} else {
		logger.Info(constants.LogMsgRedisConnected, constants.LogFieldAddr, a.Config.Redis.Addr)
	}
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
	pool := a.Database.GetPool()

	healthModule := health.NewModule(ctx, a.Database)
	extractionModule := extraction.NewModule(ctx, a.Services.LLM)
	sessionModule := session.NewModule(ctx, pool, a.Redis)
	predictionModule := prediction.NewModule(pool)
	sentimentModule := sentiment.NewModule(pool)
	triageModule := triage.NewModule(pool)
	followupModule := followup.NewModule(pool)
	analyticsModule := analytics.NewModule(pool)

	a.Modules = &Modules{
		Health:     healthModule,
		Session:    sessionModule,
		Extraction: extractionModule,
		Prediction: predictionModule,
		Sentiment:  sentimentModule,
		Triage:     triageModule,
		FollowUp:   followupModule,
		Analytics:  analyticsModule,
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
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

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

func (a *App) Start() {
	go func() {
		logger.Info(constants.LogMsgMainServerStarting, constants.LogFieldAddr, a.MainServer.Addr)
		if err := a.MainServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(constants.LogMsgMainServerFailed, constants.LogKeyError, err)
		}
	}()
	go func() {
		logger.Info(constants.LogMsgOpsServerStarting, constants.LogFieldAddr, a.OpsServer.Addr)
		if err := a.OpsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(constants.LogMsgOpsServerFailed, constants.LogKeyError, err)
		}
	}()
}

func (a *App) Shutdown(ctx context.Context) {
	logger.Info(constants.LogMsgShutdownSignalReceived)

	a.Modules.Health.GetCore().MarkUnhealthy()

	delay := a.Config.App.ShutdownDelay
	logger.Info(constants.LogMsgWaitingForShutdownDelay, constants.LogFieldDelaySeconds, delay)
	time.Sleep(time.Duration(delay) * time.Second)

	shutdownCtx, cancel := context.WithTimeout(ctx, time.Duration(a.Config.App.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := a.MainServer.Shutdown(shutdownCtx); err != nil {
		logger.Error(constants.LogMsgMainServerShutdownErr, constants.LogKeyError, err)
	}
	if err := a.OpsServer.Shutdown(shutdownCtx); err != nil {
		logger.Error(constants.LogMsgOpsServerShutdownErr, constants.LogKeyError, err)
	}

	if a.Redis != nil {
		_ = a.Redis.Close()
	}
	a.Database.Close()

	logger.Info(constants.LogMsgServiceStopped)
}
