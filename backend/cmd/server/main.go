package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/aankitroy/oauth-sample/backend/internal/auth"
	"github.com/aankitroy/oauth-sample/backend/internal/handlers"
	"github.com/aankitroy/oauth-sample/backend/internal/rbac"
	"github.com/aankitroy/oauth-sample/backend/internal/session"
	_ "github.com/lib/pq" // Postgres driver
	"github.com/rs/cors"
	gofx "go.uber.org/fx"
)

func main() {
	app := gofx.New(
		gofx.Provide(
			provideDB,
			session.NewManager,
			provideOIDCConfig,
			rbac.NewRBACStore,
			provideServer,
		),
		gofx.Invoke(registerHandlers),
	)

	app.Run()
}

func provideDB() *sql.DB {
	// Connect to Postgres
	db, err := sql.Open("postgres", "postgres://aankitroy:xyzpassword@localhost:5432/mydatabase?sslmode=disable")
	if err != nil {
		log.Fatal("Cannot connect to Postgres: ", err)
	}
	return db
}

func provideOIDCConfig() *auth.OIDCConfig {
	return &auth.OIDCConfig{
		TokenURL:    "https://uat-login.sadhguru.org/oidc/token",
		ClientID:    "miracle-of-mind-admin-nicxl1ln7y",
		RedirectURI: "http://localhost:3000/auth/callback",
		UserInfoURL: "https://uat-login.sadhguru.org/oidc/userinfo",
	}
}

func provideServer(
	oidcConfig *auth.OIDCConfig,
	mgr *session.Manager,
	store *rbac.RBACStore,
) *handlers.Server {
	return &handlers.Server{
		OIDCConfig: oidcConfig,
		SessionMgr: mgr,
		RBACStore:  store,
	}
}

func registerHandlers(lc gofx.Lifecycle, s *handlers.Server) {
	// Create a new ServeMux and register your endpoints.
	mux := http.NewServeMux()
	mux.HandleFunc("/token-exchange", s.TokenExchangeHandler)
	mux.HandleFunc("/protected", s.ProtectedHandler)
	mux.HandleFunc("/logout", s.LogoutHandler)

	// Set up CORS options to allow your Next.js frontend.
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // exact match of your frontend origin
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Wrap the mux with the CORS handler.
	handler := c.Handler(mux)

	srv := &http.Server{
		Addr:    ":8081",
		Handler: handler,
	}
	lc.Append(gofx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Println("Starting server on :8081")
			go srv.ListenAndServe()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Close()
		},
	})
}
