package main

import (
	"context"
	"follower-service/database"
	"follower-service/handler"
	"follower-service/proto"
	"follower-service/repository"
	"follower-service/tracing"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	if err := database.Connect(); err != nil {
		log.Printf("Neo4j not available at startup: %v", err)
	}
	defer func() {
		if database.Driver != nil {
			database.Driver.Close(nil)
		}
	}()

	shutdown := tracing.InitTracer("follower-service")
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer: %v", err)
		}
	}()

	tracer := otel.Tracer("follower-service")

	repo := repository.NewFollowerRepository(database.Driver)

	h := handler.NewFollowerHandler(repo)

	router := mux.NewRouter()

	router.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}).Methods("GET", "OPTIONS")

	router.Handle("/follow",
		otelhttp.NewHandler(http.HandlerFunc(h.Follow), "POST /follow"),
	).Methods("POST", "OPTIONS")

	router.Handle("/unfollow",
		otelhttp.NewHandler(http.HandlerFunc(h.Unfollow), "DELETE /unfollow"),
	).Methods("DELETE", "OPTIONS")

	router.Handle("/is-following/{followerId}/{followedId}",
		otelhttp.NewHandler(http.HandlerFunc(h.IsFollowing), "GET /is-following"),
	).Methods("GET", "OPTIONS")

	router.Handle("/following/{userId}",
		otelhttp.NewHandler(http.HandlerFunc(h.GetFollowing), "GET /following"),
	).Methods("GET", "OPTIONS")

	router.Handle("/recommendations/{userId}",
		otelhttp.NewHandler(http.HandlerFunc(h.GetRecommendations), "GET /recommendations"),
	).Methods("GET", "OPTIONS")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	corsHandler := c.Handler(router)

	// ---- gRPC server (deli isti port sa HTTP serverom) ----
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(otelGrpcUnaryInterceptorFollower(tracer)),
	)
	proto.RegisterFollowerServiceServer(grpcServer, NewFollowerGrpcServer(repo))

	mixedHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
			return
		}
		corsHandler.ServeHTTP(w, r)
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: h2c.NewHandler(http.HandlerFunc(mixedHandler), &http2.Server{}),
	}

	log.Printf("Listening on :%s (HTTP + gRPC on one handler)", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
