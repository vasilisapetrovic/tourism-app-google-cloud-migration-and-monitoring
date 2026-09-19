package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"tour-service/database"
	"tour-service/handlers"
	"tour-service/proto"
	"tour-service/tracing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	database.Connect()

	shutdown := tracing.InitTracer("tour-service")
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer: %v", err)
		}
	}()

	tracer := otel.Tracer("tour-service")

	go StartNATSSubscriber()

	r := gin.Default()

	r.Use(otelgin.Middleware("tour-service"))

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api/tours")
	{
		api.POST("", handlers.CreateTour)
		api.GET("/my", handlers.GetMyTours)
		api.GET("/published", handlers.GetPublishedTours)
		api.GET("/:id", handlers.GetTourByIdNew)

		api.PUT("/:id/publish", handlers.PublishTour)
		api.PUT("/:id/archive", handlers.ArchiveTour)
		api.PUT("/:id/reactivate", handlers.ReactivateTour)
		api.PATCH("/:id/price", handlers.UpdatePrice)
		api.POST("/:id/transport-times", handlers.AddTransportTime)

		api.POST("/:id/keypoints", handlers.CreateKeypoint)
		api.GET("/:id/keypoints", handlers.GetKeypoints)
		api.PUT("/:id/keypoints/:kpId", handlers.UpdateKeypoint)
		api.DELETE("/:id/keypoints/:kpId", handlers.DeleteKeypoint)

		api.POST("/:id/reviews", handlers.CreateReview)
		api.GET("/:id/reviews", handlers.GetReviews)

		api.POST("/:id/executions", handlers.StartExecution)
	}

	executions := r.Group("/api/executions")
	{
		executions.GET("/my", handlers.GetMyExecutions)
		executions.GET("/:execId", handlers.GetExecutionById)
		executions.PUT("/:execId/position", handlers.UpdatePosition)
		executions.PUT("/:execId/abandon", handlers.AbandonExecution)
	}

	internal := r.Group("/internal")
	{
		internal.POST("/record-purchase", handlers.RecordPurchase)
		internal.GET("/check-purchase", handlers.CheckPurchase)
	}

	// ---- gRPC server (deli isti port sa HTTP/Gin serverom) ----
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(otelGrpcUnaryInterceptor(tracer)),
	)
	reflection.Register(grpcServer)
	proto.RegisterTourServiceServer(grpcServer, &TourGrpcServer{})

	mixedHandler := func(w http.ResponseWriter, req *http.Request) {
		if req.ProtoMajor == 2 && strings.HasPrefix(req.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, req)
			return
		}
		r.ServeHTTP(w, req)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
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