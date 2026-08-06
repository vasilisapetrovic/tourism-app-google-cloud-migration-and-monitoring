package main

import (
	"context"
	"log"
	"net"
	"os"
	"tour-service/database"
	"tour-service/handlers"
	"tour-service/proto"
	"tour-service/tracing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
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

	// ---- gRPC server ----
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9090"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", grpcPort, err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(otelGrpcUnaryInterceptor(tracer)),
	)
	reflection.Register(grpcServer)
	proto.RegisterTourServiceServer(grpcServer, &TourGrpcServer{})

	go func() {
		log.Printf("gRPC server listening on :%s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

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

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8083"
	}

	log.Printf("HTTP server listening on :%s", httpPort)
	r.Run(":" + httpPort)
}