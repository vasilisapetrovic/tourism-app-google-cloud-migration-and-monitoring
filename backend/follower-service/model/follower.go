package model

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type FollowRequest struct {
	FollowerID string `json:"followerId"`
	FollowedID string `json:"followedId"`
	FollowerUsername string `json:"followerUsername"`
    FollowedUsername string `json:"followedUsername"`
}

type RecommendationResponse struct {
	Users []User `json:"users"`
}