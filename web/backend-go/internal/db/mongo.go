package db

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"backendgo/internal/types"
)

type MongoLogger struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewMongoLogger(ctx context.Context) (*MongoLogger, error) {
	uri := os.Getenv("MONGO_CONNECTION_STRING")
	if uri == "" { return &MongoLogger{}, nil }
	cli, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil { return nil, err }
	name := "course_feedback"
	if os.Getenv("ENVIRONMENT") == "dev" { name = "course_feedback_dev" }
	return &MongoLogger{client: cli, db: cli.Database(name)}, nil
}

func (m *MongoLogger) LogRecommendationFeedback(ctx context.Context, log types.RecommendationFeedbackLog) error {
	if m.db == nil { return nil }
	doc := map[string]interface{}{
		"user_id": log.UserID,
		"course": log.Course,
		"liked_recommendations": log.Liked,
		"disliked_recommendations": log.Disliked,
		"skipped_recommendations": log.Skipped,
		"action": log.Action,
		"model": log.Model,
		"recommended_from": log.RecommendedFrom,
		"timestamp": time.Now(),
	}
	_, err := m.db.Collection("recommendation_feedback").InsertOne(ctx, doc)
	return err
}

func (m *MongoLogger) LogUserFeedback(ctx context.Context, log types.UserFeedbackLog) error {
	if m.db == nil { return nil }
	doc := map[string]interface{}{
		"text": log.Text,
		"faculty": log.Faculty,
		"user_id": log.UserID,
		"phrases": log.Phrases,
		"model": log.Model,
		"study_type": log.StudyType,
		"semester": log.Semester,
		"timestamp": time.Now(),
	}
	_, err := m.db.Collection("user_feedback").InsertOne(ctx, doc)
	return err
}