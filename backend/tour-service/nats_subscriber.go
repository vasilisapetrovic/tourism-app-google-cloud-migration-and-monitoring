package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"
	"tour-service/database"
	"tour-service/handlers"

	nats "github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type natsReply struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func StartNATSSubscriber() {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL // nats://127.0.0.1:4222
	}

	nc, err := nats.Connect(url,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		log.Fatalf("NATS connect error: %v", err)
	}
	log.Printf("Connected to NATS at %s", url)

	_, err = nc.QueueSubscribe("tour.record-purchase", "tour-service-group", func(msg *nats.Msg) {
		var body struct {
			TouristId int64    `json:"touristId"`
			TourIds   []string `json:"tourIds"`
		}
		if err := json.Unmarshal(msg.Data, &body); err != nil {
			replyErr(nc, msg, "invalid payload: "+err.Error())
			return
		}

		col := database.GetCollection("tour_purchases")
		now := time.Now()
		for _, tourIdStr := range body.TourIds {
			tourObjId, err := primitive.ObjectIDFromHex(tourIdStr)
			if err != nil {
				continue
			}
			var existing handlers.TourPurchaseRecord
			if err = col.FindOne(context.Background(), bson.M{
				"touristId": body.TouristId,
				"tourId":    tourObjId,
			}).Decode(&existing); err == nil {
				continue
			}
			col.InsertOne(context.Background(), handlers.TourPurchaseRecord{
				ID:          primitive.NewObjectID(),
				TouristId:   body.TouristId,
				TourId:      tourObjId,
				PurchasedAt: now,
			})
		}

		replyOK(nc, msg)
	})
	if err != nil {
		log.Fatalf("NATS QueueSubscribe error: %v", err)
	}

	log.Println("NATS subscriptions active: tour.record-purchase")
}

func replyOK(nc *nats.Conn, msg *nats.Msg) {
	if msg.Reply == "" {
		return
	}
	data, _ := json.Marshal(natsReply{Success: true})
	nc.Publish(msg.Reply, data)
}

func replyErr(nc *nats.Conn, msg *nats.Msg, errMsg string) {
	log.Printf("NATS handler error: %s", errMsg)
	if msg.Reply == "" {
		return
	}
	data, _ := json.Marshal(natsReply{Success: false, Error: errMsg})
	nc.Publish(msg.Reply, data)
}
