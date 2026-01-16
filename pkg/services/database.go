package services

// import (
// 	"context"
// 	"log/slog"

// )

// // DatabaseService wraps database connection as a Service
// type DatabaseService struct {
// 	client *postgres.Client
// 	name   string
// }

// // NewDatabaseService creates a new database service wrapper
// func NewDatabaseService(name string, client *postgres.Client) *DatabaseService {
// 	return &DatabaseService{
// 		name:   name,
// 		client: client,
// 	}
// }

// func (s *DatabaseService) Start(ctx context.Context) error {
// 	slog.Info("Database service ready", "service", s.name)
// 	return nil
// }

// func (s *DatabaseService) Stop(ctx context.Context) error {
// 	slog.Info("Database service stopping", "service", s.name)
// 	s.client.Close()
// 	return nil
// }

// func (s *DatabaseService) Name() string {
// 	return s.name
// }
