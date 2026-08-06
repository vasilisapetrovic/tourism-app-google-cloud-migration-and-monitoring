package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"payment-service/database"
	"payment-service/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func tourServiceURL() string {
	if u := os.Getenv("TOUR_SERVICE_URL"); u != "" {
		return u
	}
	return "http://tour-service:8083"
}

func getTourStatus(tourId string) string {
	resp, err := http.Get(fmt.Sprintf("%s/api/tours/%s", tourServiceURL(), tourId))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	var info struct {
		Status string `json:"status"`
	}
	json.NewDecoder(resp.Body).Decode(&info)
	return info.Status
}

func base64Decode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

func extractUserIdFromToken(authHeader string) int64 {
	token := strings.Replace(authHeader, "Bearer ", "", 1)
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return 0
	}
	payload, _ := base64Decode(parts[1])
	var claims map[string]interface{}
	json.Unmarshal(payload, &claims)
	id, _ := claims["id"].(float64)
	return int64(id)
}

func extractRoleFromToken(authHeader string) string {
	token := strings.Replace(authHeader, "Bearer ", "", 1)
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return ""
	}
	payload, _ := base64Decode(parts[1])
	var claims map[string]interface{}
	json.Unmarshal(payload, &claims)
	role, _ := claims["role"].(string)
	return role
}

func getOrCreateCart(touristId int64) (*models.ShoppingCart, error) {
	collection := database.GetCollection("carts")
	var cart models.ShoppingCart
	err := collection.FindOne(context.Background(), bson.M{"touristId": touristId}).Decode(&cart)
	if err == mongo.ErrNoDocuments {
		cart = models.ShoppingCart{
			ID:         primitive.NewObjectID(),
			TouristId:  touristId,
			Items:      []models.OrderItem{},
			TotalPrice: 0,
		}
		_, err = collection.InsertOne(context.Background(), cart)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return &cart, nil
}

func GetCart(c *gin.Context) {
	touristId := extractUserIdFromToken(c.GetHeader("Authorization"))
	cart, err := getOrCreateCart(touristId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cart"})
		return
	}
	c.JSON(http.StatusOK, cart)
}

func AddToCart(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if extractRoleFromToken(authHeader) != "tourist" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only tourists can add to cart"})
		return
	}
	touristId := extractUserIdFromToken(authHeader)

	var body struct {
		TourId   string  `json:"tourId"`
		TourName string  `json:"tourName"`
		Price    float64 `json:"price"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.TourId == "" || body.TourName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tourId and tourName are required"})
		return
	}

	cart, err := getOrCreateCart(touristId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cart"})
		return
	}

	for _, item := range cart.Items {
		if item.TourId == body.TourId {
			c.JSON(http.StatusConflict, gin.H{"error": "Tour already in cart"})
			return
		}
	}

	switch s := getTourStatus(body.TourId); s {
	case "published":
		// ok
	case "archived":
		c.JSON(http.StatusBadRequest, gin.H{"error": "This tour has been archived and is no longer available for purchase. If you already own it, you can still start it from My Purchases."})
		return
	case "draft":
		c.JSON(http.StatusBadRequest, gin.H{"error": "This tour has not been published yet and cannot be purchased."})
		return
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tour is not available for purchase."})
		return
	}

	collection := database.GetCollection("carts")
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"touristId": touristId},
		bson.M{
			"$push": bson.M{"items": models.OrderItem{
				TourId:   body.TourId,
				TourName: body.TourName,
				Price:    body.Price,
			}},
			"$inc": bson.M{"totalPrice": body.Price},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to cart"})
		return
	}

	updated, _ := getOrCreateCart(touristId)
	c.JSON(http.StatusOK, updated)
}

func RemoveFromCart(c *gin.Context) {
	touristId := extractUserIdFromToken(c.GetHeader("Authorization"))
	tourId := c.Param("tourId")

	cart, err := getOrCreateCart(touristId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cart"})
		return
	}

	var itemPrice float64
	found := false
	for _, item := range cart.Items {
		if item.TourId == tourId {
			itemPrice = item.Price
			found = true
			break
		}
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found in cart"})
		return
	}

	collection := database.GetCollection("carts")
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"touristId": touristId},
		bson.M{
			"$pull": bson.M{"items": bson.M{"tourId": tourId}},
			"$inc":  bson.M{"totalPrice": -itemPrice},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove from cart"})
		return
	}

	updated, _ := getOrCreateCart(touristId)
	c.JSON(http.StatusOK, updated)
}

func Checkout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if extractRoleFromToken(authHeader) != "tourist" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only tourists can checkout"})
		return
	}
	touristId := extractUserIdFromToken(authHeader)

	cart, err := getOrCreateCart(touristId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cart"})
		return
	}
	if len(cart.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cart is empty"})
		return
	}

	tokensCol := database.GetCollection("purchase_tokens")
	var created []models.TourPurchaseToken

	for _, item := range cart.Items {
		if s := getTourStatus(item.TourId); s != "published" {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Tour '%s' is not available for purchase", item.TourName)})
			return
		}

		var existing models.TourPurchaseToken
		lookupErr := tokensCol.FindOne(context.Background(), bson.M{
			"touristId": touristId,
			"tourId":    item.TourId,
		}).Decode(&existing)
		if lookupErr == nil {
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
		tokensCol.InsertOne(context.Background(), token)
		created = append(created, token)
	}

	cartsCol := database.GetCollection("carts")
	cartsCol.UpdateOne(
		context.Background(),
		bson.M{"touristId": touristId},
		bson.M{"$set": bson.M{"items": []models.OrderItem{}, "totalPrice": 0.0}},
	)

	if created == nil {
		created = []models.TourPurchaseToken{}
	}

	// Notify tour service about purchases
	tourURL := tourServiceURL()
	if len(created) > 0 {
		tourIds := []string{}
		for _, t := range created {
			tourIds = append(tourIds, t.TourId)
		}
		notifyBody, _ := json.Marshal(map[string]interface{}{
			"touristId": touristId,
			"tourIds":   tourIds,
		})
		http.Post(tourURL+"/internal/record-purchase", "application/json", bytes.NewReader(notifyBody))
	}

	c.JSON(http.StatusOK, gin.H{"tokens": created})
}

func GetPurchases(c *gin.Context) {
	touristId := extractUserIdFromToken(c.GetHeader("Authorization"))
	col := database.GetCollection("purchase_tokens")
	cursor, err := col.Find(context.Background(), bson.M{"touristId": touristId})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch purchases"})
		return
	}
	defer cursor.Close(context.Background())

	var tokens []models.TourPurchaseToken
	if err = cursor.All(context.Background(), &tokens); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse purchases"})
		return
	}
	if tokens == nil {
		tokens = []models.TourPurchaseToken{}
	}
	c.JSON(http.StatusOK, tokens)
}