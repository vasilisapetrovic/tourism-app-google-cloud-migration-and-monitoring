package main

import (
	"context"
	"follower-service/database"
	"follower-service/handler"
	"follower-service/proto"
	"follower-service/repository"
	"follower-service/tracing"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
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

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9084"
	}
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("gRPC listen failed: %v", err)
	}
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(otelGrpcUnaryInterceptorFollower(tracer)),
	)
	proto.RegisterFollowerServiceServer(grpcServer, NewFollowerGrpcServer(repo))
	go func() {
		log.Printf("Follower gRPC server listening on :%s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

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

	log.Printf("Follower HTTP server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, corsHandler); err != nil {
		log.Fatalf("Follower HTTP server failed: %v", err)
	}
}