package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"payment-service/database"
	"payment-service/models"
	pb "payment-service/proto"

	nats "github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var natsConn *nats.Conn

func initNATS() {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}
	var err error
	natsConn, err = nats.Connect(url,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		fmt.Printf("Warning: NATS connect error: %v\n", err)
	} else {
		fmt.Printf("Payment service connected to NATS at %s\n", url)
	}
}

type paymentGrpcServer struct {
	pb.UnimplementedPaymentServiceServer
}

func (s *paymentGrpcServer) GetPurchases(ctx context.Context, req *pb.GetPurchasesRequest) (*pb.GetPurchasesResponse, error) {
	col := database.GetCollection("purchase_tokens")
	cursor, err := col.Find(ctx, bson.M{"touristId": req.TouristId})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch purchases")
	}
	defer cursor.Close(ctx)

	var tokens []models.TourPurchaseToken
	if err = cursor.All(ctx, &tokens); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse purchases")
	}

	var pbTokens []*pb.TourPurchaseToken
	for _, t := range tokens {
		pbTokens = append(pbTokens, &pb.TourPurchaseToken{
			Id:          t.ID.Hex(),
			TouristId:   t.TouristId,
			TourId:      t.TourId,
			TourName:    t.TourName,
			Price:       t.Price,
			PurchasedAt: t.PurchasedAt.Format(time.RFC3339),
		})
	}
	if pbTokens == nil {
		pbTokens = []*pb.TourPurchaseToken{}
	}
	return &pb.GetPurchasesResponse{Tokens: pbTokens}, nil
}

func (s *paymentGrpcServer) Checkout(ctx context.Context, req *pb.CheckoutRequest) (*pb.CheckoutResponse, error) {
	touristId := req.TouristId

	cart, err := getOrCreateCartInternal(touristId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get cart")
	}
	if len(cart.Items) == 0 {
		return &pb.CheckoutResponse{Error: "cart is empty"}, nil
	}

	tourServiceURL := os.Getenv("TOUR_SERVICE_URL")
	if tourServiceURL == "" {
		tourServiceURL = "http://tour-service:8083"
	}

	for _, item := range cart.Items {
		resp, err := http.Get(fmt.Sprintf("%s/api/tours/%s", tourServiceURL, item.TourId))
		if err != nil {
			return &pb.CheckoutResponse{Error: fmt.Sprintf("cannot validate tour %s", item.TourName)}, nil
		}
		defer resp.Body.Close()
		var tourInfo struct {
			Status string `json:"status"`
		}
		json.NewDecoder(resp.Body).Decode(&tourInfo)
		if tourInfo.Status != "published" {
			return &pb.CheckoutResponse{Error: fmt.Sprintf("tour '%s' is not available for purchase", item.TourName)}, nil
		}
	}

	tokensCol := database.GetCollection("purchase_tokens")
	var created []models.TourPurchaseToken
	var createdIds []primitive.ObjectID

	for _, item := range cart.Items {
		var existing models.TourPurchaseToken
		if err := tokensCol.FindOne(ctx, bson.M{"touristId": touristId, "tourId": item.TourId}).Decode(&existing); err == nil {
			continue
		}
		token := models.TourPurchaseToken{
			ID:          primitive.NewObjectID(),
			TouristId:   touristId,
			TourId:      item.TourId,
			TourName:    item.TourName,
			Price:       item.Price,
			PurchasedAt: time.Now(),
		}
		tokensCol.InsertOne(ctx, token)
		created = append(created, token)
		createdIds = append(createdIds, token.ID)
	}

	if len(created) > 0 {
		var tourIds []string
		for _, t := range created {
			tourIds = append(tourIds, t.TourId)
		}
		payload, _ := json.Marshal(map[string]interface{}{
			"touristId": touristId,
			"tourIds":   tourIds,
		})

		sagaOK := false
		if natsConn != nil {
			reply, natsErr := natsConn.Request("tour.record-purchase", payload, 5*time.Second)
			if natsErr == nil {
				var natsReply struct {
					Success bool `json:"success"`
				}
				json.Unmarshal(reply.Data, &natsReply)
				sagaOK = natsReply.Success
			}
		}
		if !sagaOK {
			for _, id := range createdIds {
				tokensCol.DeleteOne(ctx, bson.M{"_id": id})
			}
			return &pb.CheckoutResponse{Error: "checkout failed — please try again"}, nil
		}
	}

	// Clear cart
	cartsCol := database.GetCollection("carts")
	cartsCol.UpdateOne(ctx,
		bson.M{"touristId": touristId},
		bson.M{"$set": bson.M{"items": []models.OrderItem{}, "totalPrice": 0.0}},
	)

	var pbTokens []*pb.TourPurchaseToken
	for _, t := range created {
		pbTokens = append(pbTokens, &pb.TourPurchaseToken{
			Id:          t.ID.Hex(),
			TouristId:   t.TouristId,
			TourId:      t.TourId,
			TourName:    t.TourName,
			Price:       t.Price,
			PurchasedAt: t.PurchasedAt.Format(time.RFC3339),
		})
	}
	if pbTokens == nil {
		pbTokens = []*pb.TourPurchaseToken{}
	}
	return &pb.CheckoutResponse{Tokens: pbTokens}, nil
}

func getOrCreateCartInternal(touristId int64) (*models.ShoppingCart, error) {
	collection := database.GetCollection("carts")
	var cart models.ShoppingCart
	err := collection.FindOne(context.Background(), bson.M{"touristId": touristId}).Decode(&cart)
	if err != nil {
		cart = models.ShoppingCart{
			ID:         primitive.NewObjectID(),
			TouristId:  touristId,
			Items:      []models.OrderItem{},
			TotalPrice: 0,
		}
		_, err = collection.InsertOne(context.Background(), cart)
	}
	return &cart, err
}

func (s *paymentGrpcServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	if req.TouristId == 0 || req.TourId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "touristId and tourId required")
	}

	col := database.GetCollection("purchase_tokens")
	var result struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	err := col.FindOne(ctx, bson.M{
		"touristId": req.TouristId,
		"tourId":    req.TourId,
	}).Decode(&result)

	if err != nil {
		return &pb.ValidateTokenResponse{Valid: false, Error: "token not found"}, nil
	}

	return &pb.ValidateTokenResponse{Valid: true, TokenId: result.ID.Hex()}, nil
}
