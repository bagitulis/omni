package google

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// AuthService handles Google API authentication
type AuthService struct {
	mu           sync.RWMutex
	credentials  []byte
	clients      map[string]*sheets.Service // tenantID -> sheets client
	driveClients map[string]*drive.Service  // tenantID -> drive client
}

// NewAuthService creates a new Google auth service
func NewAuthService(credentials []byte) *AuthService {
	return &AuthService{
		credentials:  credentials,
		clients:      make(map[string]*sheets.Service),
		driveClients: make(map[string]*drive.Service),
	}
}

// GetClient returns authenticated Sheets client for tenant
func (s *AuthService) GetClient(ctx context.Context, tenantID string) (*sheets.Service, error) {
	// Check if credentials are available
	if len(s.credentials) == 0 {
		return nil, fmt.Errorf("no service account credentials configured")
	}

	s.mu.RLock()
	if client, ok := s.clients[tenantID]; ok {
		s.mu.RUnlock()
		return client, nil
	}
	s.mu.RUnlock()

	// Create new client
	client, err := s.createClient(ctx)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.clients[tenantID] = client
	s.mu.Unlock()

	return client, nil
}

// createClient creates a new authenticated Sheets client
func (s *AuthService) createClient(ctx context.Context) (*sheets.Service, error) {
	if len(s.credentials) == 0 {
		return nil, fmt.Errorf("no service account credentials configured")
	}

	config, err := google.JWTConfigFromJSON(s.credentials, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	client := config.Client(ctx)
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("create sheets service: %w", err)
	}

	return srv, nil
}

// CreateOAuthClient creates client using OAuth tokens
func (s *AuthService) CreateOAuthClient(ctx context.Context, token *oauth2.Token) (*sheets.Service, error) {
	config := &oauth2.Config{
		Scopes: []string{sheets.SpreadsheetsScope},
	}

	client := config.Client(ctx, token)
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("create sheets service: %w", err)
	}

	return srv, nil
}

// RefreshToken refreshes OAuth token if expired
func (s *AuthService) RefreshToken(ctx context.Context, config *oauth2.Config, token *oauth2.Token) (*oauth2.Token, error) {
	tokenSource := config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("refresh token: %w", err)
	}
	return newToken, nil
}

// InvalidateClient removes cached client for tenant
func (s *AuthService) InvalidateClient(tenantID string) {
	s.mu.Lock()
	delete(s.clients, tenantID)
	delete(s.driveClients, tenantID)
	s.mu.Unlock()
}

// GetDriveClient returns authenticated Drive client for tenant
func (s *AuthService) GetDriveClient(ctx context.Context, tenantID string) (*drive.Service, error) {
	// Check if credentials are available
	if len(s.credentials) == 0 {
		return nil, fmt.Errorf("no service account credentials configured")
	}

	s.mu.RLock()
	if client, ok := s.driveClients[tenantID]; ok {
		s.mu.RUnlock()
		return client, nil
	}
	s.mu.RUnlock()

	// Create new drive client
	config, err := google.JWTConfigFromJSON(s.credentials, drive.DriveReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	httpClient := config.Client(ctx)
	srv, err := drive.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create drive service: %w", err)
	}

	s.mu.Lock()
	s.driveClients[tenantID] = srv
	s.mu.Unlock()

	return srv, nil
}
