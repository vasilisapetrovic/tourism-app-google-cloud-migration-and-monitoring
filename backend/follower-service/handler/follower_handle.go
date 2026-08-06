package handler

import (
	"encoding/json"
	"follower-service/model"
	"follower-service/repository"
	"net/http"

	"github.com/gorilla/mux"
)

type FollowerHandler struct {
	repo *repository.FollowerRepository
}

func NewFollowerHandler(repo *repository.FollowerRepository) *FollowerHandler {
	return &FollowerHandler{repo: repo}
}

func (h *FollowerHandler) Follow(w http.ResponseWriter, r *http.Request) {
	var req model.FollowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.repo.Follow(req.FollowerID, req.FollowedID, req.FollowerUsername, req.FollowedUsername); err != nil {
		if err.Error() == "already following this user" {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, "Follow failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *FollowerHandler) Unfollow(w http.ResponseWriter, r *http.Request) {
	var req model.FollowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.repo.Unfollow(req.FollowerID, req.FollowedID); err != nil {
		http.Error(w, "Failed to unfollow: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *FollowerHandler) IsFollowing(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	followerID := vars["followerId"]
	followedID := vars["followedId"]

	following, err := h.repo.IsFollowing(followerID, followedID)
	if err != nil {
		http.Error(w, "Failed to check: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"following": following})
}

func (h *FollowerHandler) GetFollowing(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	users, err := h.repo.GetFollowing(userID)
	if err != nil {
		http.Error(w, "Failed to get following: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *FollowerHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	users, err := h.repo.GetRecommendations(userID)
	if err != nil {
		http.Error(w, "Failed to get recommendations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}