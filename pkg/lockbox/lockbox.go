package lockbox

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	ycsdk "github.com/yandex-cloud/go-sdk"
	lockboxpayload "github.com/yandex-cloud/go-sdk/gen/lockboxpayload"
	lockbox1 "github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"

	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/models"
)

var (
	client         *lockboxpayload.PayloadServiceClient
	clientOnce     sync.Once
	payloadCache   *models.LockboxPayload
	cacheExpiry    time.Time
	cacheMutex     sync.RWMutex
	secretID       string
)

// ClientAdapter implements LockboxClient using the global functions
type ClientAdapter struct{}

// NewClientAdapter creates a new ClientAdapter
func NewClientAdapter() *ClientAdapter {
	return &ClientAdapter{}
}

// GetUserTokens retrieves access and refresh tokens for a specific user
func (c *ClientAdapter) GetUserTokens(ctx context.Context, reviewerLogin string) (*models.UserTokens, error) {
	return GetUserTokens(ctx, reviewerLogin)
}

// StoreUserTokens stores new tokens for a user
func (c *ClientAdapter) StoreUserTokens(ctx context.Context, reviewerLogin, accessToken, refreshToken string) error {
	return StoreUserTokens(ctx, reviewerLogin, accessToken, refreshToken)
}

// DeleteUserTokens removes tokens for a user
func (c *ClientAdapter) DeleteUserTokens(ctx context.Context, reviewerLogin string) error {
	return DeleteUserTokens(ctx, reviewerLogin)
}

// InitClient initializes the Lockbox client
func InitClient(ctx context.Context) (*lockboxpayload.PayloadServiceClient, error) {
	var initErr error
	clientOnce.Do(func() {
		secretID = os.Getenv("LOCKBOX_SECRET_ID")
		if secretID == "" {
			initErr = fmt.Errorf("LOCKBOX_SECRET_ID environment variable not set")
			return
		}

		// Use service account credentials
		credentials := ycsdk.InstanceServiceAccount()
		sdk, err := ycsdk.Build(ctx, ycsdk.Config{
			Credentials: credentials,
		})
		if err != nil {
			initErr = fmt.Errorf("failed to create Yandex Cloud SDK: %w", err)
			return
		}

		client = sdk.LockboxPayload().Payload()
	})

	return client, initErr
}

// GetClient returns the Lockbox client (initialized or panics)
func GetClient(ctx context.Context) (*lockboxpayload.PayloadServiceClient, error) {
	if client == nil {
		return InitClient(ctx)
	}
	return client, nil
}

// GetUserTokens retrieves access and refresh tokens for a specific user from Lockbox
func GetUserTokens(ctx context.Context, reviewerLogin string) (*models.UserTokens, error) {
	pl, err := getPayload(ctx)
	if err != nil {
		return nil, err
	}

	tokens, ok := pl.Users[reviewerLogin]
	if !ok {
		return nil, fmt.Errorf("tokens not found for user: %s", reviewerLogin)
	}

	return &tokens, nil
}

// StoreUserTokens stores new tokens for a user in Lockbox
// Note: This is a placeholder implementation. The Yandex Cloud Lockbox Go SDK
// doesn't currently support writing secrets programmatically.
// For production, you'll need to use one of:
// 1. The yc command-line tool
// 2. The Yandex Cloud REST API directly
// 3. A separate write service
func StoreUserTokens(ctx context.Context, reviewerLogin, accessToken, refreshToken string) error {
	// Get current payload
	pl, err := getPayload(ctx)
	if err != nil {
		return err
	}

	// Update tokens for user
	if pl.Users == nil {
		pl.Users = make(map[string]models.UserTokens)
	}
	pl.Users[reviewerLogin] = models.UserTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	// Serialize to JSON
	data, err := json.Marshal(pl)
	if err != nil {
		return fmt.Errorf("failed to marshal Lockbox payload: %w", err)
	}

	// TODO: Implement actual write to Lockbox
	// For now, store the JSON data that would be written
	_ = data

	// Invalidate cache
	cacheMutex.Lock()
	payloadCache = nil
	cacheMutex.Unlock()

	return fmt.Errorf("StoreUserTokens: writing to Lockbox is not yet implemented via SDK. " +
		"Please use 'yc lockbox payload add' or the REST API to update the secret. " +
		"Payload: %s", string(data))
}

// getPayload fetches and caches the Lockbox payload
func getPayload(ctx context.Context) (*models.LockboxPayload, error) {
	cacheMutex.RLock()
	if payloadCache != nil && time.Now().Before(cacheExpiry) {
		defer cacheMutex.RUnlock()
		return payloadCache, nil
	}
	cacheMutex.RUnlock()

	lbClient, err := GetClient(ctx)
	if err != nil {
		return nil, err
	}

	pl, err := getRawPayload(ctx, lbClient)
	if err != nil {
		return nil, err
	}

	// Cache for 5 minutes
	cacheMutex.Lock()
	payloadCache = pl
	cacheExpiry = time.Now().Add(5 * time.Minute)
	cacheMutex.Unlock()

	return pl, nil
}

// getRawPayload fetches the payload from Lockbox without caching
func getRawPayload(ctx context.Context, lbClient *lockboxpayload.PayloadServiceClient) (*models.LockboxPayload, error) {
	resp, err := lbClient.Get(ctx, &lockbox1.GetPayloadRequest{
		SecretId: secretID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Lockbox payload: %w", err)
	}

	var pl models.LockboxPayload

	// Find the entry containing the user tokens
	for _, entry := range resp.Entries {
		key := entry.GetKey()
		if key == "users" || key == "payload" || key == "tokens" || key == "" {
			var data []byte
			if entry.GetTextValue() != "" {
				data = []byte(entry.GetTextValue())
			}
			if len(data) > 0 {
				err = json.Unmarshal(data, &pl)
				if err != nil {
					return nil, fmt.Errorf("failed to unmarshal Lockbox payload: %w", err)
				}
				break
			}
		}
	}

	return &pl, nil
}

// DeleteUserTokens removes tokens for a user from Lockbox
// Note: This is a placeholder implementation. See StoreUserTokens for details.
func DeleteUserTokens(ctx context.Context, reviewerLogin string) error {
	// Get current payload
	pl, err := getPayload(ctx)
	if err != nil {
		return err
	}

	// Remove user tokens
	if pl.Users != nil {
		delete(pl.Users, reviewerLogin)
	}

	// Invalidate cache
	cacheMutex.Lock()
	payloadCache = nil
	cacheMutex.Unlock()

	return fmt.Errorf("DeleteUserTokens: writing to Lockbox is not yet implemented via SDK. " +
		"Please use 'yc lockbox payload add' or the REST API to update the secret")
}

// InvalidateCache clears the payload cache
func InvalidateCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	payloadCache = nil
	cacheExpiry = time.Time{}
}

// GetSecretID returns the configured Lockbox secret ID
func GetSecretID() string {
	return secretID
}

// SetPayloadCache sets the payload cache (useful for testing or manual updates)
func SetPayloadCache(pl *models.LockboxPayload, ttl time.Duration) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	payloadCache = pl
	cacheExpiry = time.Now().Add(ttl)
}
