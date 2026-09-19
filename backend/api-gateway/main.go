package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"

	"api-gateway/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func verifyJWT(tokenString string, secret []byte) (map[string]interface{}, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}
	message := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(message))
	expected := mac.Sum(nil)
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid signature encoding")
	}
	if !hmac.Equal(expected, sig) {
		return nil, fmt.Errorf("invalid signature")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid payload encoding")
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("invalid payload")
	}
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, fmt.Errorf("token expired")
		}
	}
	return claims, nil
}

func extractClaims(authHeader string, secret []byte) map[string]interface{} {
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil
	}
	claims, err := verifyJWT(strings.TrimPrefix(authHeader, "Bearer "), secret)
	if err != nil {
		return nil
	}
	return claims
}

func isUserBlocked(stakeholdersURL string, userID int64) bool {
	url := fmt.Sprintf("%s/internal/users/%d/blocked", stakeholdersURL, userID)
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Blocked bool `json:"blocked"`
	}
	json.Unmarshal(body, &result)
	return result.Blocked
}

func isPublicRoute(method, path string) bool {
	if (method == http.MethodPost && path == "/api/auth/login") ||
		(method == http.MethodPost && path == "/api/auth/register") {
		return true
	}
	if method == http.MethodGet {
		matched, _ := regexp.MatchString(`^/api/position/\d+$`, path)
		if matched {
			return true
		}
	}
	return false
}

func jsonError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":"%s"}`, msg)
}

func jsonUnauthorized(w http.ResponseWriter, msg string) { jsonError(w, http.StatusUnauthorized, msg) }
func jsonForbidden(w http.ResponseWriter, msg string)    { jsonError(w, http.StatusForbidden, msg) }

type corsWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (c *corsWriter) setCORS() {
	c.ResponseWriter.Header().Set("Access-Control-Allow-Origin", "*")
	c.ResponseWriter.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
	c.ResponseWriter.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func (c *corsWriter) WriteHeader(status int) {
	if !c.wroteHeader {
		c.setCORS()
		c.wroteHeader = true
	}
	c.ResponseWriter.WriteHeader(status)
}

func (c *corsWriter) Write(b []byte) (int, error) {
	if !c.wroteHeader {
		c.setCORS()
		c.wroteHeader = true
	}
	return c.ResponseWriter.Write(b)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(&corsWriter{ResponseWriter: w}, r)
	})
}

func authMiddleware(next http.Handler, jwtSecret []byte, stakeholdersURL string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isPublicRoute(r.Method, r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			jsonUnauthorized(w, "Authorization token required")
			return
		}
		claims, err := verifyJWT(strings.TrimPrefix(authHeader, "Bearer "), jwtSecret)
		if err != nil {
			jsonUnauthorized(w, "Invalid or expired token")
			return
		}
		userID := int64(claims["id"].(float64))
		if isUserBlocked(stakeholdersURL, userID) {
			jsonForbidden(w, "Account is blocked")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func newReverseProxy(target string) *httputil.ReverseProxy {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Invalid target URL %s: %v", target, err)
	}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error for %s: %v", r.URL.Path, err)
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
	}
	return proxy
}

func handleGetTourByIdGRPC(w http.ResponseWriter, r *http.Request, client proto.TourServiceClient, jwtSecret []byte) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		jsonError(w, http.StatusBadRequest, "invalid path")
		return
	}
	tourId := parts[3]

	claims := extractClaims(r.Header.Get("Authorization"), jwtSecret)
	var role string
	var userId int64
	if claims != nil {
		role, _ = claims["role"].(string)
		userId = int64(claims["id"].(float64))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.GetTourById(ctx, &proto.GetTourByIdRequest{
		Id:     tourId,
		Role:   role,
		UserId: userId,
	})
	if err != nil {
		log.Printf("gRPC GetTourById error: %v", err)
		jsonError(w, http.StatusNotFound, "tour not found")
		return
	}

	t := resp.Tour
	result := map[string]interface{}{
		"id": t.Id, "authorId": t.AuthorId, "authorUsername": t.AuthorUsername,
		"name": t.Name, "description": t.Description, "difficulty": t.Difficulty,
		"tags": t.Tags, "status": t.Status, "price": t.Price,
		"lengthKm": t.LengthKm, "createdAt": t.CreatedAt,
		"publishedAt": t.PublishedAt, "archivedAt": t.ArchivedAt,
	}
	var transportTimes []map[string]interface{}
	for _, tt := range t.TransportTimes {
		transportTimes = append(transportTimes, map[string]interface{}{
			"transportType": tt.TransportType, "minutes": tt.Minutes,
		})
	}
	result["transportTimes"] = transportTimes
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handlePublishTourGRPC(w http.ResponseWriter, r *http.Request, client proto.TourServiceClient, jwtSecret []byte) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		jsonError(w, http.StatusBadRequest, "invalid path")
		return
	}
	tourId := parts[3]

	claims := extractClaims(r.Header.Get("Authorization"), jwtSecret)
	var authorId int64
	if claims != nil {
		authorId = int64(claims["id"].(float64))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.PublishTour(ctx, &proto.PublishTourRequest{
		Id:       tourId,
		AuthorId: authorId,
	})
	if err != nil {
		log.Printf("gRPC PublishTour error: %v", err)
		jsonError(w, http.StatusInternalServerError, "failed to publish tour")
		return
	}
	if resp.Error != "" {
		jsonError(w, http.StatusBadRequest, resp.Error)
		return
	}

	t := resp.Tour
	result := map[string]interface{}{
		"id": t.Id, "authorId": t.AuthorId, "authorUsername": t.AuthorUsername,
		"name": t.Name, "description": t.Description, "difficulty": t.Difficulty,
		"tags": t.Tags, "status": t.Status, "price": t.Price,
		"lengthKm": t.LengthKm, "createdAt": t.CreatedAt,
		"publishedAt": t.PublishedAt, "archivedAt": t.ArchivedAt,
	}
	var transportTimes []map[string]interface{}
	for _, tt := range t.TransportTimes {
		transportTimes = append(transportTimes, map[string]interface{}{
			"transportType": tt.TransportType, "minutes": tt.Minutes,
		})
	}
	result["transportTimes"] = transportTimes
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleStartExecutionGRPC(w http.ResponseWriter, r *http.Request, client proto.TourServiceExecutionClient, jwtSecret []byte) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		jsonError(w, http.StatusBadRequest, "invalid path")
		return
	}
	tourId := parts[3]

	claims := extractClaims(r.Header.Get("Authorization"), jwtSecret)
	var touristId int64
	var touristUsername string
	if claims != nil {
		touristId = int64(claims["id"].(float64))
		touristUsername, _ = claims["username"].(string)
	}

	var body struct {
		StartLat float64 `json:"startLat"`
		StartLng float64 `json:"startLng"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.StartExecution(ctx, &proto.StartExecutionRequest{
		TourId:          tourId,
		TouristId:       touristId,
		TouristUsername: touristUsername,
		StartLat:        body.StartLat,
		StartLng:        body.StartLng,
	})
	if err != nil {
		log.Printf("gRPC StartExecution error: %v", err)
		jsonError(w, http.StatusInternalServerError, "failed to start execution")
		return
	}
	if resp.Error != "" {
		jsonError(w, http.StatusBadRequest, resp.Error)
		return
	}

	e := resp.Execution
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":              e.Id,
		"tourId":          e.TourId,
		"touristId":       e.TouristId,
		"touristUsername": e.TouristUsername,
		"status":          e.Status,
		"startedAt":       e.StartedAt,
		"lastActivity":    e.LastActivity,
		"currentLat":      e.CurrentLat,
		"currentLng":      e.CurrentLng,
		"completedKeypoints": e.CompletedKeypoints,
	})
}

func handleGetExecutionGRPC(w http.ResponseWriter, r *http.Request, client proto.TourServiceExecutionClient, jwtSecret []byte) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		jsonError(w, http.StatusBadRequest, "invalid path")
		return
	}
	execId := parts[3]

	claims := extractClaims(r.Header.Get("Authorization"), jwtSecret)
	var touristId int64
	if claims != nil {
		touristId = int64(claims["id"].(float64))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.GetExecution(ctx, &proto.GetExecutionRequest{
		ExecId:    execId,
		TouristId: touristId,
	})
	if err != nil {
		log.Printf("gRPC GetExecution error: %v", err)
		jsonError(w, http.StatusNotFound, "execution not found")
		return
	}
	if resp.Error != "" {
		jsonError(w, http.StatusBadRequest, resp.Error)
		return
	}

	e := resp.Execution
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":              e.Id,
		"tourId":          e.TourId,
		"touristId":       e.TouristId,
		"touristUsername": e.TouristUsername,
		"status":          e.Status,
		"startedAt":       e.StartedAt,
		"endedAt":         e.EndedAt,
		"lastActivity":    e.LastActivity,
		"currentLat":      e.CurrentLat,
		"currentLng":      e.CurrentLng,
		"completedKeypoints": e.CompletedKeypoints,
	})
}

func handleFollowGRPC(w http.ResponseWriter, r *http.Request, client proto.FollowerServiceClient, jwtSecret []byte) {
	claims := extractClaims(r.Header.Get("Authorization"), jwtSecret)
	if claims == nil {
		jsonUnauthorized(w, "invalid token")
		return
	}

	var req struct {
		FollowedId string `json:"followedId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request")
		return
	}

	followerId := fmt.Sprintf("%.0f", claims["id"].(float64))
	followerUsername, _ := claims["username"].(string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Follow(ctx, &proto.FollowRequest{
		FollowerId:       followerId,
		FollowedId:       req.FollowedId,
		FollowerUsername: followerUsername,
		FollowedUsername: "",
	})
	if err != nil {
		log.Printf("gRPC Follow error: %v", err)
		jsonError(w, http.StatusInternalServerError, "follow failed")
		return
	}
	if !resp.Success {
		jsonError(w, http.StatusBadRequest, resp.Error)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func handleUnfollowGRPC(w http.ResponseWriter, r *http.Request, client proto.FollowerServiceClient, jwtSecret []byte) {
	claims := extractClaims(r.Header.Get("Authorization"), jwtSecret)
	if claims == nil {
		jsonUnauthorized(w, "invalid token")
		return
	}

	followedId := strings.TrimPrefix(r.URL.Path, "/api/unfollow/")
	followerId := fmt.Sprintf("%.0f", claims["id"].(float64))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Unfollow(ctx, &proto.UnfollowRequest{
		FollowerId: followerId,
		FollowedId: followedId,
	})
	if err != nil {
		log.Printf("gRPC Unfollow error: %v", err)
		jsonError(w, http.StatusInternalServerError, "unfollow failed")
		return
	}
	if !resp.Success {
		jsonError(w, http.StatusBadRequest, resp.Error)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func handleGetFollowingGRPC(w http.ResponseWriter, r *http.Request, client proto.FollowerServiceClient, jwtSecret []byte) {
	claims := extractClaims(r.Header.Get("Authorization"), jwtSecret)
	if claims == nil {
		jsonUnauthorized(w, "invalid token")
		return
	}

	userId := fmt.Sprintf("%.0f", claims["id"].(float64))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.GetFollowing(ctx, &proto.GetFollowingRequest{UserId: userId})
	if err != nil {
		log.Printf("gRPC GetFollowing error: %v", err)
		jsonError(w, http.StatusInternalServerError, "failed to get following")
		return
	}

	users := resp.Users
	if users == nil {
		users = []*proto.UserProto{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func handleGetPurchasesGRPC(w http.ResponseWriter, r *http.Request, client proto.PaymentServiceClient, jwtSecret []byte) {
	claims := extractClaims(r.Header.Get("Authorization"), jwtSecret)
	if claims == nil {
		jsonUnauthorized(w, "invalid token")
		return
	}
	touristId := int64(claims["id"].(float64))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.GetPurchases(ctx, &proto.GetPurchasesRequest{TouristId: touristId})
	if err != nil {
		log.Printf("gRPC GetPurchases error: %v", err)
		jsonError(w, http.StatusInternalServerError, "failed to get purchases")
		return
	}

	tokens := resp.Tokens
	if tokens == nil {
		tokens = []*proto.TourPurchaseToken{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

func handleCheckoutGRPC(w http.ResponseWriter, r *http.Request, client proto.PaymentServiceClient, jwtSecret []byte) {
	claims := extractClaims(r.Header.Get("Authorization"), jwtSecret)
	if claims == nil {
		jsonUnauthorized(w, "invalid token")
		return
	}
	touristId := int64(claims["id"].(float64))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.Checkout(ctx, &proto.CheckoutRequest{TouristId: touristId})
	if err != nil {
		log.Printf("gRPC Checkout error: %v", err)
		jsonError(w, http.StatusInternalServerError, "checkout failed")
		return
	}
	if resp.Error != "" {
		jsonError(w, http.StatusBadRequest, resp.Error)
		return
	}

	tokens := resp.Tokens
	if tokens == nil {
		tokens = []*proto.TourPurchaseToken{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"tokens": tokens})
}

func main() {
	port := getEnv("PORT", "8080")
	jwtSecret := []byte(getEnv("JWT_SECRET", "mysupersecretkey123"))

	stakeholders := getEnv("STAKEHOLDERS_SERVICE_URL", "http://stakeholders-service:8081")
	blog := getEnv("BLOG_SERVICE_URL", "http://blog-service:8082")
	toursHTTP := getEnv("TOUR_SERVICE_URL", "http://tour-service:8083")
	follower := getEnv("FOLLOWER_SERVICE_URL", "http://follower-service:8084")
	payment := getEnv("PAYMENT_SERVICE_URL", "http://payment-service:8086")
	toursGRPC := getEnv("TOUR_SERVICE_GRPC_URL", "tour-service:9090")
	paymentGRPCAddr := getEnv("PAYMENT_SERVICE_GRPC_URL", "payment-service:9091")
	followerGRPCAddr := getEnv("FOLLOWER_SERVICE_GRPC_URL", "follower-service:9084")

	stakeholdersProxy := newReverseProxy(stakeholders)
	blogProxy := newReverseProxy(blog)
	toursProxy := newReverseProxy(toursHTTP)
	followerProxy := newReverseProxy(follower)
	paymentProxy := newReverseProxy(payment)

	conn, err := grpc.NewClient(
		toursGRPC,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
	)
	if err != nil {
		log.Fatalf("Failed to connect to tour-service gRPC: %v", err)
	}
	defer conn.Close()
	tourGrpcClient := proto.NewTourServiceClient(conn)
	tourExecGrpcClient := proto.NewTourServiceExecutionClient(conn)

	paymentGrpcConn, err := grpc.NewClient(
		paymentGRPCAddr,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
	)
	if err != nil {
		log.Fatalf("Failed to connect to payment-service gRPC: %v", err)
	}
	defer paymentGrpcConn.Close()
	paymentGrpcClient := proto.NewPaymentServiceClient(paymentGrpcConn)

	followerGrpcConn, err := grpc.NewClient(
		followerGRPCAddr,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
	)
	if err != nil {
		log.Fatalf("Failed to connect to follower-service gRPC: %v", err)
	}
	defer followerGrpcConn.Close()
	followerGrpcClient := proto.NewFollowerServiceClient(followerGrpcConn)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		log.Printf("%s %s", r.Method, path)

		switch {
		case strings.HasPrefix(path, "/api/auth/"),
			strings.HasPrefix(path, "/api/profile"),
			strings.HasPrefix(path, "/api/position"),
			strings.HasPrefix(path, "/api/user_management/"),
			strings.HasPrefix(path, "/uploads/"):
			stakeholdersProxy.ServeHTTP(w, r)

		case strings.HasPrefix(path, "/api/blogs"),
			strings.HasPrefix(path, "/api/comments"),
			strings.HasPrefix(path, "/api/likes"):
			blogProxy.ServeHTTP(w, r)

		case path == "/api/purchases" && r.Method == http.MethodGet:
			log.Printf("gRPC GetPurchases: %s", path)
			handleGetPurchasesGRPC(w, r, paymentGrpcClient, jwtSecret)

		case path == "/api/cart/checkout" && r.Method == http.MethodPost:
			log.Printf("gRPC Checkout: %s", path)
			handleCheckoutGRPC(w, r, paymentGrpcClient, jwtSecret)

		case strings.HasPrefix(path, "/api/cart"),
			strings.HasPrefix(path, "/api/purchases"):
			paymentProxy.ServeHTTP(w, r)

		case strings.HasPrefix(path, "/api/tours"),
			strings.HasPrefix(path, "/api/executions"):

			isTourById := r.Method == http.MethodGet &&
				regexp.MustCompile(`^/api/tours/[0-9a-fA-F]{24}$`).MatchString(path)
			isPublish := r.Method == http.MethodPut &&
				regexp.MustCompile(`^/api/tours/[0-9a-fA-F]{24}/publish$`).MatchString(path)
			isStartExecution := r.Method == http.MethodPost &&
				regexp.MustCompile(`^/api/tours/[0-9a-fA-F]{24}/executions$`).MatchString(path)
			isGetExecution := r.Method == http.MethodGet &&
				regexp.MustCompile(`^/api/executions/[0-9a-fA-F]{24}$`).MatchString(path)

			if isTourById {
				log.Printf("gRPC GetTourById: %s", path)
				handleGetTourByIdGRPC(w, r, tourGrpcClient, jwtSecret)
			} else if isPublish {
				log.Printf("gRPC PublishTour: %s", path)
				handlePublishTourGRPC(w, r, tourGrpcClient, jwtSecret)
			} else if isStartExecution {
				log.Printf("gRPC StartExecution: %s", path)
				handleStartExecutionGRPC(w, r, tourExecGrpcClient, jwtSecret)
			} else if isGetExecution {
				log.Printf("gRPC GetExecution: %s", path)
				handleGetExecutionGRPC(w, r, tourExecGrpcClient, jwtSecret)
			} else {
				toursProxy.ServeHTTP(w, r)
			}

		case strings.HasPrefix(path, "/api/follow") && r.Method == http.MethodPost:
			log.Printf("gRPC Follow: %s", path)
			handleFollowGRPC(w, r, followerGrpcClient, jwtSecret)

		case strings.HasPrefix(path, "/api/unfollow") && r.Method == http.MethodDelete:
			log.Printf("gRPC Unfollow: %s", path)
			handleUnfollowGRPC(w, r, followerGrpcClient, jwtSecret)

		case path == "/api/following" && r.Method == http.MethodGet:
			log.Printf("gRPC GetFollowing: %s", path)
			handleGetFollowingGRPC(w, r, followerGrpcClient, jwtSecret)

		case strings.HasPrefix(path, "/api/is-following"),
			strings.HasPrefix(path, "/api/recommendations"):
			r.URL.Path = path[4:]
			followerProxy.ServeHTTP(w, r)

		default:
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})

	handler := corsMiddleware(authMiddleware(mux, jwtSecret, stakeholders))

	log.Printf("API Gateway listening on :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Failed to start gateway: %v", err)
	}
}