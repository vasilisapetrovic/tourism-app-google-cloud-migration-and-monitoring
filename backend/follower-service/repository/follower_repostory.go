package repository

import (
	"context"
	"errors"
	"follower-service/model"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type FollowerRepository struct {
	driver neo4j.DriverWithContext
}

func NewFollowerRepository(driver neo4j.DriverWithContext) *FollowerRepository {
	return &FollowerRepository{driver: driver}
}

func (r *FollowerRepository) ensureDriver() error {
	if r == nil || r.driver == nil {
		return errors.New("neo4j driver is not initialized")
	}
	return nil
}

func (r *FollowerRepository) Follow(followerID, followedID, followerUsername, followedUsername string) error {
	if err := r.ensureDriver(); err != nil {
		return err
	}

	ctx := context.Background()
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (a:User {id: $followerID})-[:FOLLOWS]->(b:User {id: $followedID})
		 RETURN count(*) > 0 AS already_following`,
		map[string]interface{}{
			"followerID": followerID,
			"followedID": followedID,
		})
	if err != nil {
		return err
	}

	if result.Next(ctx) {
		alreadyFollowing, _ := result.Record().Get("already_following")
		if alreadyFollowing.(bool) {
			return errors.New("already following this user")
		}
	}

	_, err = session.Run(ctx,
		`MERGE (a:User {id: $followerID})
		 ON CREATE SET a.username = $followerUsername
		 MERGE (b:User {id: $followedID})
		 ON CREATE SET b.username = $followedUsername
		 MERGE (a)-[:FOLLOWS]->(b)`,
		map[string]interface{}{
			"followerID":       followerID,
			"followedID":       followedID,
			"followerUsername": followerUsername,
			"followedUsername": followedUsername,
		})
	return err
}

func (r *FollowerRepository) Unfollow(followerID, followedID string) error {
	if err := r.ensureDriver(); err != nil {
		return err
	}

	ctx := context.Background()
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.Run(ctx,
		`MATCH (a:User {id: $followerID})-[r:FOLLOWS]->(b:User {id: $followedID})
		 DELETE r`,
		map[string]interface{}{
			"followerID": followerID,
			"followedID": followedID,
		})
	return err
}

func (r *FollowerRepository) IsFollowing(followerID, followedID string) (bool, error) {
	if err := r.ensureDriver(); err != nil {
		return false, err
	}

	ctx := context.Background()
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (a:User {id: $followerID})-[:FOLLOWS]->(b:User {id: $followedID})
		 RETURN count(*) > 0 AS following`,
		map[string]interface{}{
			"followerID": followerID,
			"followedID": followedID,
		})
	if err != nil {
		return false, err
	}

	if result.Next(ctx) {
		following, _ := result.Record().Get("following")
		return following.(bool), nil
	}
	return false, nil
}

func (r *FollowerRepository) GetFollowing(userID string) ([]model.User, error) {
	if err := r.ensureDriver(); err != nil {
		return nil, err
	}

	ctx := context.Background()
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (a:User {id: $userID})-[:FOLLOWS]->(b:User)
		 RETURN DISTINCT b.id AS id, b.username AS username`,
		map[string]interface{}{"userID": userID})
	if err != nil {
		return nil, err
	}

	var users []model.User
	for result.Next(ctx) {
		id, _ := result.Record().Get("id")
		username, _ := result.Record().Get("username")
		usernameStr := ""
		if username != nil {
			usernameStr = username.(string)
		}
		users = append(users, model.User{ID: id.(string), Username: usernameStr})
	}
	return users, nil
}

func (r *FollowerRepository) GetRecommendations(userID string) ([]model.User, error) {
	if err := r.ensureDriver(); err != nil {
		return nil, err
	}

	ctx := context.Background()
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (a:User {id: $userID})-[:FOLLOWS]->(b:User)-[:FOLLOWS]->(c:User)
		 WHERE c.id <> $userID
		 AND NOT (a)-[:FOLLOWS]->(c)
		 RETURN DISTINCT c.id AS id, c.username AS username`,
		map[string]interface{}{"userID": userID})
	if err != nil {
		return nil, err
	}

	var users []model.User
	for result.Next(ctx) {
		id, _ := result.Record().Get("id")
		username, _ := result.Record().Get("username")
		usernameStr := ""
		if username != nil {
			usernameStr = username.(string)
		}
		users = append(users, model.User{ID: id.(string), Username: usernameStr})
	}
	return users, nil
}