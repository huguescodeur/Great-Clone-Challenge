// @title           Insta-Like API
// @version         1.0
// @description     API REST pour une application de type Instagram (posts, likes, commentaires, follow).
// @termsOfService  http://swagger.io/terms/

// @contact.name   Hugues Codeur
// @contact.email  huguescodeur@gmail.com

// @license.name  MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Entrez le token JWT avec le préfixe "Bearer ". Exemple: "Bearer eyJhbGci..."

package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/huguescodeur/insta-like/docs"
	"github.com/huguescodeur/insta-like/internal/app"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// config.LoadConfig()

	// dsn := os.Getenv("DATABASE_URL")
	const DATABASE_URL = "postgres://postgres:64062639@localhost:5432/insta"

	db, err := sql.Open("pgx", DATABASE_URL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("database connection established")

	myApp := app.Init(db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      myApp.Routes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	slog.Info("Server start", "addr", "http://localhost:8080/api/v1")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}

}
