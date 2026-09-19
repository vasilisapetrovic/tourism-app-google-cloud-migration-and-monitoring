package main

import (
	"log"
	"net/http"
	"os"
	"payment-service/database"
	"payment-service/handlers"
	pb "payment-service/proto"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
)

func main() {
	database.Connect()
	initNATS()

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

	// ---- gRPC server (deli isti port sa HTTP/Gin serverom) ----
	grpcServer := grpc.NewServer()
	pb.RegisterPaymentServiceServer(grpcServer, &paymentGrpcServer{})

	mixedHandler := func(w http.ResponseWriter, req *http.Request) {
		if req.ProtoMajor == 2 && strings.HasPrefix(req.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, req)
			return
		}
		r.ServeHTTP(w, req)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
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