package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/XoDeR/customer-support-desk-go/internal/adapter/http/v1/handler"
	"github.com/XoDeR/customer-support-desk-go/internal/adapter/http/v1/middleware"
	"github.com/XoDeR/customer-support-desk-go/internal/adapter/http/v1/router"
	"github.com/XoDeR/customer-support-desk-go/internal/adapter/repository/postgres"
	"github.com/XoDeR/customer-support-desk-go/internal/adapter/storage"
	"github.com/XoDeR/customer-support-desk-go/internal/adapter/ws"
	"github.com/XoDeR/customer-support-desk-go/internal/domain/entity"
	"github.com/XoDeR/customer-support-desk-go/internal/infrastructure/config"
	"github.com/XoDeR/customer-support-desk-go/internal/infrastructure/database"
	redisinfrastructure "github.com/XoDeR/customer-support-desk-go/internal/infrastructure/redis"
	"github.com/XoDeR/customer-support-desk-go/internal/usecase"
	jwtpkg "github.com/XoDeR/customer-support-desk-go/pkg/jwt"
	"github.com/XoDeR/customer-support-desk-go/pkg/migration"
	"github.com/XoDeR/customer-support-desk-go/pkg/uuidv7"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, e := config.Load()
	if e != nil {
		log.Fatal(e)
	}
	db, e := database.NewPostgresConnection(cfg)
	if e != nil {
		log.Fatal(e)
	}
	defer db.Close()
	if cfg.App.AutoMigrate {
		if e = migration.NewManager(db, "migrations/app").Migrate(context.Background()); e != nil {
			log.Fatal(e)
		}
	}
	r := postgres.New(db)
	tx := database.NewTransactionManager(db)
	jwt := jwtpkg.NewManager(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL, cfg.JWT.Issuer)
	auth := usecase.NewAuth(r, tx, jwt)
	bootstrap(context.Background(), auth, r, cfg.Admin.Email, cfg.Admin.Password, entity.RoleAdmin)
	bootstrap(context.Background(), auth, r, cfg.Agent.Email, cfg.Agent.Password, entity.RoleAgent)
	redisClient, e := redisinfrastructure.New(cfg)
	if e != nil {
		log.Fatal(e)
	}
	defer redisClient.Close()
	localStorage, e := storage.NewLocal(cfg.Storage.Directory)
	if e != nil {
		log.Fatal(e)
	}
	hub := ws.NewHub()
	ws.Subscribe(context.Background(), redisClient, hub)
	publisher := ws.NewPublisher(redisClient)
	h := &handler.Handler{Auth: auth, Tickets: usecase.NewTickets(r, tx, publisher), Repo: r, InternalToken: cfg.InternalToken, Storage: localStorage, Limiter: redisinfrastructure.NewLimiter(redisClient), Hub: hub}
	srv := &http.Server{Addr: fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port), Handler: router.New(h, middleware.NewAuth(jwt)), ReadTimeout: cfg.Server.ReadTimeout, WriteTimeout: cfg.Server.WriteTimeout}
	go func() {
		log.Printf("api listening on %s", srv.Addr)
		if e := srv.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
			log.Fatal(e)
		}
	}()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	_ = srv.Shutdown(context.Background())
}
func bootstrap(ctx context.Context, a *usecase.Auth, r *postgres.Repository, email, password string, role entity.Role) {
	if email == "" || password == "" {
		return
	}
	if _, e := r.GetUserByEmail(ctx, email); e == nil {
		return
	}
	h, e := usecase.HashPassword(password)
	if e == nil {
		_ = r.CreateUser(ctx, entity.User{ID: uuidv7.New(), Email: email, PasswordHash: h, Role: role, Status: "active"})
	}
}
