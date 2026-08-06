package main

import (
	"log"
	"net"
	"os"
	"payment-service/database"
	"payment-service/handlers"
	pb "payment-service/proto"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	database.Connect()
	initNATS()

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9091"
	}
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("gRPC listen failed: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterPaymentServiceServer(grpcServer, &paymentGrpcServer{})
	go func() {
		log.Printf("Payment gRPC server listening on :%s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// ---- HTTP server (Gin) ----
	r := gin.Default()

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

	cart := r.Group("/api/cart")
	{
		cart.GET("", handlers.GetCart)
		cart.POST("/items", handlers.AddToCart)
		cart.DELETE("/items/:tourId", handlers.RemoveFromCart)
		cart.POST("/checkout", handlers.Checkout)
	}

	r.GET("/api/purchases", handlers.GetPurchases)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}

	log.Printf("Payment HTTP server listening on :%s", port)
	r.Run(":" + port)
}
