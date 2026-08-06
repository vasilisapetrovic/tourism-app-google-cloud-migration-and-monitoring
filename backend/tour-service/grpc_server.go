package main

import (
	"context"
	"os"
	"strings"
	"time"
	"tour-service/database"
	"tour-service/models"
	"tour-service/proto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type TourGrpcServer struct {
	proto.UnimplementedTourServiceServer
}

func tourToProto(t models.Tour) *proto.Tour {
	p := &proto.Tour{
		Id:             t.ID.Hex(),
		AuthorId:       t.AuthorId,
		AuthorUsername: t.AuthorUsername,
		Name:           t.Name,
		Description:    t.Description,
		Difficulty:     t.Difficulty,
		Tags:           t.Tags,
		Status:         t.Status,
		Price:          t.Price,
		LengthKm:       t.LengthKm,
		CreatedAt:      t.CreatedAt.Format(time.RFC3339),
	}

	if t.Tags == nil {
		p.Tags = []string{}
	}

	if t.PublishedAt != nil {
		p.PublishedAt = t.PublishedAt.Format(time.RFC3339)
	}
	if t.ArchivedAt != nil {
		p.ArchivedAt = t.ArchivedAt.Format(time.RFC3339)
	}

	for _, tt := range t.TransportTimes {
		p.TransportTimes = append(p.TransportTimes, &proto.TransportTime{
			TransportType: tt.TransportType,
			Minutes:       int32(tt.Minutes),
		})
	}

	return p
}

func (s *TourGrpcServer) GetTourById(ctx context.Context, req *proto.GetTourByIdRequest) (*proto.GetTourByIdResponse, error) {
	objId, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tour ID")
	}

	collection := database.GetCollection("tours")
	var tour models.Tour
	err = collection.FindOne(ctx, bson.M{"_id": objId}).Decode(&tour)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "tour not found")
	}

	if tour.Status == "draft" && req.UserId != 0 && tour.AuthorId != req.UserId {
		return nil, status.Errorf(codes.PermissionDenied, "tour not available")
	}

	return &proto.GetTourByIdResponse{Tour: tourToProto(tour)}, nil
}

func (s *TourGrpcServer) PublishTour(ctx context.Context, req *proto.PublishTourRequest) (*proto.PublishTourResponse, error) {
	objId, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tour ID")
	}

	collection := database.GetCollection("tours")
	var tour models.Tour
	err = collection.FindOne(ctx, bson.M{"_id": objId}).Decode(&tour)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "tour not found")
	}
	if tour.AuthorId != req.AuthorId {
		return &proto.PublishTourResponse{Error: "not your tour"}, nil
	}
	if tour.Status != "draft" {
		return &proto.PublishTourResponse{Error: "only draft tours can be published"}, nil
	}

	if strings.TrimSpace(tour.Name) == "" ||
		strings.TrimSpace(tour.Description) == "" ||
		strings.TrimSpace(tour.Difficulty) == "" ||
		len(tour.Tags) == 0 {
		return &proto.PublishTourResponse{Error: "tour must have name, description, difficulty and at least one tag"}, nil
	}

	keypointCollection := database.GetCollection("keypoints")
	count, err := keypointCollection.CountDocuments(ctx, bson.M{"tourId": objId})
	if err != nil || count < 2 {
		return &proto.PublishTourResponse{Error: "tour must have at least 2 key points"}, nil
	}

	if len(tour.TransportTimes) == 0 {
		return &proto.PublishTourResponse{Error: "tour must have at least one transport time"}, nil
	}

	now := time.Now()
	_, err = collection.UpdateOne(ctx,
		bson.M{"_id": objId},
		bson.M{"$set": bson.M{"status": "published", "publishedAt": now}},
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish tour")
	}

	tour.Status = "published"
	tour.PublishedAt = &now
	return &proto.PublishTourResponse{Tour: tourToProto(tour)}, nil
}

func getPaymentClient() (proto.PaymentServiceGrpcClient, *grpc.ClientConn, error) {
	addr := os.Getenv("PAYMENT_GRPC_ADDR")
	if addr == "" {
		addr = "payment-service:9091"
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return proto.NewPaymentServiceGrpcClient(conn), conn, nil
}

func (s *TourGrpcServer) StartExecution(ctx context.Context, req *proto.StartExecutionRequest) (*proto.StartExecutionResponse, error) {
	tourObjId, err := primitive.ObjectIDFromHex(req.TourId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tour ID")
	}

	tourCol := database.GetCollection("tours")
	var tour models.Tour
	if err := tourCol.FindOne(ctx, bson.M{"_id": tourObjId}).Decode(&tour); err != nil {
		return nil, status.Errorf(codes.NotFound, "tour not found")
	}
	if tour.Status != "published" && tour.Status != "archived" {
		return &proto.StartExecutionResponse{Error: "tour must be published or archived"}, nil
	}

	paymentClient, conn, err := getPaymentClient()
	if err != nil {
		return &proto.StartExecutionResponse{Error: "payment service unavailable"}, nil
	}
	defer conn.Close()

	tokenResp, err := paymentClient.ValidateToken(ctx, &proto.PaymentValidateTokenRequest{
		TouristId: req.TouristId,
		TourId:    req.TourId,
	})
	if err != nil || !tokenResp.Valid {
		msg := "tour not purchased"
		if tokenResp != nil && tokenResp.Error != "" {
			msg = tokenResp.Error
		}
		return &proto.StartExecutionResponse{Error: msg}, nil
	}

	execCol := database.GetCollection("executions")
	now := time.Now()

	execCol.UpdateMany(ctx, bson.M{
		"touristId": req.TouristId,
		"tourId":    tourObjId,
		"status":    "active",
	}, bson.M{"$set": bson.M{"status": "abandoned", "endedAt": now, "lastActivity": now}})

	execution := models.TourExecution{
		ID:                 primitive.NewObjectID(),
		TourID:             tourObjId,
		TouristId:          req.TouristId,
		TouristUsername:    req.TouristUsername,
		Status:             "active",
		StartedAt:          now,
		LastActivity:       now,
		CurrentLat:         req.StartLat,
		CurrentLng:         req.StartLng,
		CompletedKeypoints: []models.CompletedKeypoint{},
	}

	if _, err := execCol.InsertOne(ctx, execution); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create execution")
	}

	return &proto.StartExecutionResponse{Execution: executionToProto(execution)}, nil
}

func (s *TourGrpcServer) GetExecution(ctx context.Context, req *proto.GetExecutionRequest) (*proto.GetExecutionResponse, error) {
	execObjId, err := primitive.ObjectIDFromHex(req.ExecId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid execution ID")
	}

	execCol := database.GetCollection("executions")
	var execution models.TourExecution
	if err := execCol.FindOne(ctx, bson.M{"_id": execObjId}).Decode(&execution); err != nil {
		return nil, status.Errorf(codes.NotFound, "execution not found")
	}
	if execution.TouristId != req.TouristId {
		return nil, status.Errorf(codes.PermissionDenied, "not your execution")
	}

	return &proto.GetExecutionResponse{Execution: executionToProto(execution)}, nil
}

func executionToProto(e models.TourExecution) *proto.ExecutionProto {
	p := &proto.ExecutionProto{
		Id:              e.ID.Hex(),
		TourId:          e.TourID.Hex(),
		TouristId:       e.TouristId,
		TouristUsername: e.TouristUsername,
		Status:          e.Status,
		StartedAt:       e.StartedAt.Format(time.RFC3339),
		LastActivity:    e.LastActivity.Format(time.RFC3339),
		CurrentLat:      e.CurrentLat,
		CurrentLng:      e.CurrentLng,
	}
	if e.EndedAt != nil {
		p.EndedAt = e.EndedAt.Format(time.RFC3339)
	}
	for _, ck := range e.CompletedKeypoints {
		p.CompletedKeypoints = append(p.CompletedKeypoints, &proto.CompletedKeypointProto{
			KeypointId:  ck.KeypointID.Hex(),
			Name:        ck.Name,
			CompletedAt: ck.CompletedAt.Format(time.RFC3339),
		})
	}
	return p
}