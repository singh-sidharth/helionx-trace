package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/singh-sidharth/helionx-trace/internal/api"
	"github.com/singh-sidharth/helionx-trace/internal/service"
	"github.com/singh-sidharth/helionx-trace/internal/store"
)

func main() {
	_ = godotenv.Load()

	storeBackend := getEnv("STORE_BACKEND", "memory")

	var eventStore store.EventStore

	switch storeBackend {
	case "memory":
		eventStore = store.NewMemoryStore()
		log.Println("using InMemoryStore")
	case "postgres":
		db, err := openPostgres()
		if err != nil {
			log.Fatalf("failed to connect to postgres: %v", err)
		}
		defer db.Close()

		eventStore = store.NewPostgresStore(db)
		log.Println("using PostgresStore")
	default:
		log.Fatalf("unsupported STORE_BACKEND: %s", storeBackend)
	}

	timelineService := service.NewTimelineService(eventStore)
	handler := api.NewHandler(eventStore, timelineService)

	mux := http.NewServeMux()
	handler.Register(mux)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           loggingMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("helionx event debugger running on :8080")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s took=%s", r.Method, r.URL.Path, time.Since(start))
	})
}

func openPostgres() (*sql.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	name := getEnv("DB_NAME", "helionx")
	user := getEnv("DB_USER", "helionx")
	password := getEnv("DB_PASSWORD", "helionx")
	sslmode := getEnv("DB_SSLMODE", "disable")
	connectTimeout := getEnv("DB_CONNECT_TIMEOUT", "5")

	dsn := "host=" + host + " port=" + port + " dbname=" + name + " user=" + user + " password=" + password + " sslmode=" + sslmode + " connect_timeout=" + connectTimeout

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// connection pool tuning
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	return db, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
