package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

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

func isUserBlocked(stakeholdersURL string, userID int64) bool {
	url := fmt.Sprintf("%s/internal/users/%d/blocked", stakeholdersURL, userID)
	resp, err := httpClient.Get(url)
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

func forwardRequest(w http.ResponseWriter, r *http.Request, targetBase string, path string) {
	targetURL := targetBase + path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}
	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		log.Printf("Forward error creating request for %s: %v", path, err)
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	if auth := r.Header.Get("Authorization"); auth != "" {
		proxyReq.Header.Set("Authorization", auth)
	}
	if ct := r.Header.Get("Content-Type"); ct != "" {
		proxyReq.Header.Set("Content-Type", ct)
	}
	proxyReq.ContentLength = r.ContentLength
	resp, err := httpClient.Do(proxyReq)
	if err != nil {
		log.Printf("Forward error for %s: %v", path, err)
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	port := getEnv("PORT", "8080")
	jwtSecret := []byte(getEnv("JWT_SECRET", "mysupersecretkey123"))

	stakeholders := getEnv("STAKEHOLDERS_SERVICE_URL", "http://stakeholders-service:8081")
	blog := getEnv("BLOG_SERVICE_URL", "http://blog-service:8082")
	toursHTTP := getEnv("TOUR_SERVICE_URL", "http://tour-service:8083")
	follower := getEnv("FOLLOWER_SERVICE_URL", "http://follower-service:8084")
	payment := getEnv("PAYMENT_SERVICE_URL", "http://payment-service:8086")

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
			forwardRequest(w, r, stakeholders, path)

		case strings.HasPrefix(path, "/api/blogs"),
			strings.HasPrefix(path, "/api/comments"),
			strings.HasPrefix(path, "/api/likes"):
			forwardRequest(w, r, blog, path)

		case strings.HasPrefix(path, "/api/cart"),
			strings.HasPrefix(path, "/api/purchases"):
			forwardRequest(w, r, payment, path)

		case strings.HasPrefix(path, "/api/tours"),
			strings.HasPrefix(path, "/api/executions"):
			forwardRequest(w, r, toursHTTP, path)

		case strings.HasPrefix(path, "/api/follow"),
    	strings.HasPrefix(path, "/api/unfollow"),
    	strings.HasPrefix(path, "/api/following"):
    	forwardRequest(w, r, follower, path[4:])

		case strings.HasPrefix(path, "/api/is-following"),
			strings.HasPrefix(path, "/api/recommendations"):
			forwardRequest(w, r, follower, path[4:])

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