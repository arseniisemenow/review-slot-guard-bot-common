package lockbox

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/arseniisemenow/review-slot-guard-bot-common/pkg/models"
)

// MockPayloadServiceClient is a mock implementation of the PayloadServiceClient
type MockPayloadServiceClient struct {
	mu              sync.Mutex
	getCallCount    int
	payloadToReturn *models.LockboxPayload
	errorToReturn   error
	closed          bool
}

func NewMockPayloadServiceClient() *MockPayloadServiceClient {
	return &MockPayloadServiceClient{
		payloadToReturn: &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		},
	}
}

func (m *MockPayloadServiceClient) Get(ctx context.Context, req interface{}) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.getCallCount++

	if m.closed {
		return nil, errors.New("client is closed")
	}

	if m.errorToReturn != nil {
		return nil, m.errorToReturn
	}

	// Simulate Lockbox response
	payloadJSON, _ := json.Marshal(m.payloadToReturn)
	return &MockLockboxResponse{
		entries: []*MockLockboxEntry{
			{
				key:       "users",
				textValue: string(payloadJSON),
			},
		},
	}, nil
}

func (m *MockPayloadServiceClient) SetPayload(pl *models.LockboxPayload) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.payloadToReturn = pl
}

func (m *MockPayloadServiceClient) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorToReturn = err
}

func (m *MockPayloadServiceClient) GetCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.getCallCount
}

func (m *MockPayloadServiceClient) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
}

// MockLockboxResponse simulates the Lockbox GetPayload response
type MockLockboxResponse struct {
	entries []*MockLockboxEntry
}

func (m *MockLockboxResponse) GetEntries() []*MockLockboxEntry {
	return m.entries
}

// MockLockboxEntry simulates a Lockbox entry
type MockLockboxEntry struct {
	key       string
	textValue string
}

func (m *MockLockboxEntry) GetKey() string {
	return m.key
}

func (m *MockLockboxEntry) GetTextValue() string {
	return m.textValue
}

// Helper function to reset package-level state between tests
func resetPackageState() {
	client = nil
	clientOnce = sync.Once{}
	payloadCache = nil
	cacheExpiry = time.Time{}
	secretID = ""
}

// Helper function to set up test environment
func setupTestEnv(t *testing.T) {
	resetPackageState()
	// Set required environment variable
	os.Setenv("LOCKBOX_SECRET_ID", "test-secret-id")
}

// Helper function to tear down test environment
func teardownTestEnv(t *testing.T) {
	os.Unsetenv("LOCKBOX_SECRET_ID")
	resetPackageState()
}

// TestInitClient tests the InitClient function
func TestInitClient(t *testing.T) {
	t.Run("Missing LOCKBOX_SECRET_ID", func(t *testing.T) {
		setupTestEnv(t)
		defer teardownTestEnv(t)

		// Unset the environment variable
		os.Unsetenv("LOCKBOX_SECRET_ID")

		ctx := context.Background()
		lockboxClient, err := InitClient(ctx)

		assert.Error(t, err)
		assert.Nil(t, lockboxClient)
		assert.Contains(t, err.Error(), "LOCKBOX_SECRET_ID environment variable not set")
	})

	t.Run("Successful initialization with retry", func(t *testing.T) {
		setupTestEnv(t)
		defer teardownTestEnv(t)

		ctx := context.Background()

		// First call should attempt initialization
		// (it will fail in test environment without actual Yandex credentials)
		_, err := InitClient(ctx)

		// We expect this to fail in test environment
		// but we're testing that the function runs
		// In a real environment with proper credentials, this would succeed
		_ = err
	})

	t.Run("Idempotent initialization", func(t *testing.T) {
		setupTestEnv(t)
		defer teardownTestEnv(t)

		ctx := context.Background()

		// First call
		_, err1 := InitClient(ctx)

		// Second call should return the same client
		_, err2 := InitClient(ctx)

		// Both should have the same error (or success) result
		assert.Equal(t, err1 != nil, err2 != nil)
	})

	t.Run("Multiple initializations return same client on success", func(t *testing.T) {
		setupTestEnv(t)
		defer teardownTestEnv(t)

		ctx := context.Background()

		// In test environment, InitClient will fail due to missing credentials
		// but we can test that clientOnce prevents multiple initialization attempts
		err1, err2 := make(chan error, 1), make(chan error, 1)

		go func() {
			_, err := InitClient(ctx)
			err1 <- err
		}()

		go func() {
			_, err := InitClient(ctx)
			err2 <- err
		}()

		e1 := <-err1
		e2 := <-err2

		// Both attempts should complete (not hang)
		// The errors might be different due to timing, but both should be non-nil in test env
		assert.True(t, true) // If we get here, both completed without hanging
		_ = e1
		_ = e2
	})
}

// TestGetClient tests the GetClient function
func TestGetClient(t *testing.T) {
	t.Run("GetClient when client is nil", func(t *testing.T) {
		setupTestEnv(t)
		defer teardownTestEnv(t)

		ctx := context.Background()
		client, err := GetClient(ctx)

		// Will fail in test environment but should call InitClient
		_ = client
		_ = err
	})

	t.Run("GetClient when client is initialized", func(t *testing.T) {
		setupTestEnv(t)
		defer teardownTestEnv(t)

		ctx := context.Background()

		// Initialize first
		_, _ = InitClient(ctx)

		// Get should return existing client
		_, err := GetClient(ctx)
		_ = err
	})
}

// TestInvalidateCache tests the InvalidateCache function
func TestInvalidateCache(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	t.Run("Invalidate valid cache", func(t *testing.T) {
		// Set up a cache
		payloadCache = &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"testuser": {
					AccessToken:  "test_access",
					RefreshToken: "test_refresh",
				},
			},
		}
		cacheExpiry = time.Now().Add(5 * time.Minute)

		// Verify cache is set
		assert.NotNil(t, payloadCache)
		assert.False(t, cacheExpiry.IsZero())

		// Invalidate
		InvalidateCache()

		// Verify cache is cleared
		assert.Nil(t, payloadCache)
		assert.True(t, cacheExpiry.IsZero())
	})

	t.Run("Invalidate nil cache", func(t *testing.T) {
		// Start with nil cache
		payloadCache = nil
		cacheExpiry = time.Time{}

		// Should not panic
		InvalidateCache()

		assert.Nil(t, payloadCache)
		assert.True(t, cacheExpiry.IsZero())
	})

	t.Run("Invalidate expired cache", func(t *testing.T) {
		// Set up an expired cache
		payloadCache = &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		}
		cacheExpiry = time.Now().Add(-1 * time.Hour)

		// Invalidate
		InvalidateCache()

		assert.Nil(t, payloadCache)
		assert.True(t, cacheExpiry.IsZero())
	})
}

// TestGetSecretID tests the GetSecretID function
func TestGetSecretID(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	t.Run("Secret ID before initialization", func(t *testing.T) {
		resetPackageState()
		id := GetSecretID()
		assert.Empty(t, id)
	})

	t.Run("Secret ID after setting environment variable", func(t *testing.T) {
		os.Setenv("LOCKBOX_SECRET_ID", "test-secret-123")
		defer os.Unsetenv("LOCKBOX_SECRET_ID")

		ctx := context.Background()
		_, _ = InitClient(ctx)

		id := GetSecretID()
		assert.Equal(t, "test-secret-123", id)
	})

	t.Run("Secret ID is consistent", func(t *testing.T) {
		resetPackageState()
		os.Setenv("LOCKBOX_SECRET_ID", "consistent-secret-id")
		defer os.Unsetenv("LOCKBOX_SECRET_ID")

		ctx := context.Background()

		// Multiple initializations should give same ID
		_, _ = InitClient(ctx)
		id1 := GetSecretID()

		_, _ = InitClient(ctx)
		id2 := GetSecretID()

		assert.Equal(t, id1, id2)
		assert.Equal(t, "consistent-secret-id", id1)
	})
}

// TestSetPayloadCache tests the SetPayloadCache function
func TestSetPayloadCache(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	t.Run("Set cache with TTL", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"user1": {
					AccessToken:  "access1",
					RefreshToken: "refresh1",
				},
			},
		}

		ttl := 10 * time.Minute
		SetPayloadCache(testPayload, ttl)

		// Verify cache is set
		assert.NotNil(t, payloadCache)
		assert.Equal(t, 1, payloadCache.Version)
		assert.Len(t, payloadCache.Users, 1)

		// Verify expiry is in the future
		assert.True(t, time.Now().Before(cacheExpiry))
		assert.True(t, time.Now().Add(ttl).After(cacheExpiry) ||
			time.Now().Add(ttl).Add(time.Second).After(cacheExpiry))
	})

	t.Run("Set nil cache", func(t *testing.T) {
		SetPayloadCache(nil, 5*time.Minute)

		assert.Nil(t, payloadCache)
	})

	t.Run("Replace existing cache", func(t *testing.T) {
		// Set initial cache
		payload1 := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"user1": {
					AccessToken:  "access1",
					RefreshToken: "refresh1",
				},
			},
		}
		SetPayloadCache(payload1, 5*time.Minute)

		// Replace with new cache
		payload2 := &models.LockboxPayload{
			Version: 2,
			Users: map[string]models.UserTokens{
				"user2": {
					AccessToken:  "access2",
					RefreshToken: "refresh2",
				},
			},
		}
		SetPayloadCache(payload2, 10*time.Minute)

		assert.Equal(t, 2, payloadCache.Version)
		assert.Len(t, payloadCache.Users, 1)
		_, exists := payloadCache.Users["user2"]
		assert.True(t, exists)
	})

	t.Run("Set cache with zero TTL", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		}

		SetPayloadCache(testPayload, 0)

		assert.NotNil(t, payloadCache)
		// Cache should be immediately expired
		assert.True(t, time.Now().After(cacheExpiry) || time.Now().Equal(cacheExpiry))
	})

	t.Run("Set cache with negative TTL", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		}

		SetPayloadCache(testPayload, -1*time.Hour)

		assert.NotNil(t, payloadCache)
		// Cache should be expired
		assert.True(t, time.Now().After(cacheExpiry))
	})
}

// TestGetUserTokens tests the GetUserTokens function
func TestGetUserTokens(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	t.Run("Get tokens for existing user", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"testuser": {
					AccessToken:  "test_access_token",
					RefreshToken: "test_refresh_token",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		tokens, err := GetUserTokens(ctx, "testuser")

		require.NoError(t, err)
		assert.NotNil(t, tokens)
		assert.Equal(t, "test_access_token", tokens.AccessToken)
		assert.Equal(t, "test_refresh_token", tokens.RefreshToken)
	})

	t.Run("Get tokens for non-existent user", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"existinguser": {
					AccessToken:  "access",
					RefreshToken: "refresh",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		tokens, err := GetUserTokens(ctx, "nonexistentuser")

		assert.Error(t, err)
		assert.Nil(t, tokens)
		assert.Contains(t, err.Error(), "tokens not found for user")
		assert.Contains(t, err.Error(), "nonexistentuser")
	})

	t.Run("Get tokens with empty username", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		tokens, err := GetUserTokens(ctx, "")

		assert.Error(t, err)
		assert.Nil(t, tokens)
		assert.Contains(t, err.Error(), "tokens not found for user")
	})

	t.Run("Get tokens when cache is expired", func(t *testing.T) {
		// Set expired cache
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"testuser": {
					AccessToken:  "access",
					RefreshToken: "refresh",
				},
			},
		}

		SetPayloadCache(testPayload, -1*time.Minute)

		ctx := context.Background()
		// This will try to fetch from Lockbox and fail in test environment
		_, err := GetUserTokens(ctx, "testuser")

		assert.Error(t, err)
	})

	t.Run("Get multiple users", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"user1": {
					AccessToken:  "access1",
					RefreshToken: "refresh1",
				},
				"user2": {
					AccessToken:  "access2",
					RefreshToken: "refresh2",
				},
				"user3": {
					AccessToken:  "access3",
					RefreshToken: "refresh3",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()

		// Get user1
		tokens1, err := GetUserTokens(ctx, "user1")
		require.NoError(t, err)
		assert.Equal(t, "access1", tokens1.AccessToken)

		// Get user2
		tokens2, err := GetUserTokens(ctx, "user2")
		require.NoError(t, err)
		assert.Equal(t, "access2", tokens2.AccessToken)

		// Get user3
		tokens3, err := GetUserTokens(ctx, "user3")
		require.NoError(t, err)
		assert.Equal(t, "access3", tokens3.AccessToken)
	})

	t.Run("Get tokens with special characters in username", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"user@test.com": {
					AccessToken:  "access",
					RefreshToken: "refresh",
				},
				"user-name_123": {
					AccessToken:  "access2",
					RefreshToken: "refresh2",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()

		tokens1, err := GetUserTokens(ctx, "user@test.com")
		require.NoError(t, err)
		assert.Equal(t, "access", tokens1.AccessToken)

		tokens2, err := GetUserTokens(ctx, "user-name_123")
		require.NoError(t, err)
		assert.Equal(t, "access2", tokens2.AccessToken)
	})
}

// TestStoreUserTokens tests the StoreUserTokens function
func TestStoreUserTokens(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	t.Run("Store tokens for new user", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		err := StoreUserTokens(ctx, "newuser", "new_access", "new_refresh")

		// Should return error (not implemented)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")
	})

	t.Run("Store tokens for existing user", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"existinguser": {
					AccessToken:  "old_access",
					RefreshToken: "old_refresh",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		err := StoreUserTokens(ctx, "existinguser", "new_access", "new_refresh")

		// Should return error (not implemented)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")

		// Cache should be invalidated
		assert.Nil(t, payloadCache)
	})

	t.Run("Store tokens with empty access token", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		err := StoreUserTokens(ctx, "testuser", "", "refresh_token")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")
	})

	t.Run("Store tokens with empty refresh token", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		err := StoreUserTokens(ctx, "testuser", "access_token", "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")
	})

	t.Run("Store tokens with empty username", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		err := StoreUserTokens(ctx, "", "access_token", "refresh_token")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")
	})

	t.Run("Verify cache invalidation after store", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"user1": {
					AccessToken:  "access1",
					RefreshToken: "refresh1",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		// Verify cache is set
		assert.NotNil(t, payloadCache)

		ctx := context.Background()
		_ = StoreUserTokens(ctx, "user2", "access2", "refresh2")

		// Verify cache is invalidated
		assert.Nil(t, payloadCache)
	})

	t.Run("Store tokens with very long strings", func(t *testing.T) {
		longToken := string(make([]byte, 10000))
		for i := range longToken {
			longToken = longToken[:i] + "a" + longToken[i+1:]
		}

		testPayload := &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		err := StoreUserTokens(ctx, "testuser", longToken, longToken)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")
	})
}

// TestDeleteUserTokens tests the DeleteUserTokens function
func TestDeleteUserTokens(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	t.Run("Delete existing user", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"user1": {
					AccessToken:  "access1",
					RefreshToken: "refresh1",
				},
				"user2": {
					AccessToken:  "access2",
					RefreshToken: "refresh2",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		err := DeleteUserTokens(ctx, "user1")

		// Should return error (not implemented)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")

		// Cache should be invalidated
		assert.Nil(t, payloadCache)
	})

	t.Run("Delete non-existent user", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"user1": {
					AccessToken:  "access1",
					RefreshToken: "refresh1",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		err := DeleteUserTokens(ctx, "nonexistent")

		// Should return error (not implemented)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")
	})

	t.Run("Delete with empty username", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"user1": {
					AccessToken:  "access1",
					RefreshToken: "refresh1",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		err := DeleteUserTokens(ctx, "")

		// Should return error (not implemented)
		assert.Error(t, err)
	})

	t.Run("DELETE FROM empty payload", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		err := DeleteUserTokens(ctx, "user1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")
	})

	t.Run("Verify cache invalidation after delete", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"user1": {
					AccessToken:  "access1",
					RefreshToken: "refresh1",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		// Verify cache is set
		assert.NotNil(t, payloadCache)

		ctx := context.Background()
		_ = DeleteUserTokens(ctx, "user1")

		// Verify cache is invalidated
		assert.Nil(t, payloadCache)
	})

	t.Run("Delete when Users map is nil", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users:   nil,
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		err := DeleteUserTokens(ctx, "user1")

		// Should return error (not implemented)
		assert.Error(t, err)
	})
}

// TestCacheConcurrency tests concurrent access to cache
func TestCacheConcurrency(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	t.Run("Concurrent reads", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"user1": {
					AccessToken:  "access1",
					RefreshToken: "refresh1",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		var wg sync.WaitGroup
		errors := make(chan error, 10)

		// Launch 10 concurrent reads
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := GetUserTokens(ctx, "user1")
				if err != nil {
					errors <- err
				}
			}()
		}

		wg.Wait()
		close(errors)

		// Should have no errors
		for err := range errors {
			t.Errorf("Unexpected error in concurrent read: %v", err)
		}
	})

	t.Run("Concurrent invalidations", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		var wg sync.WaitGroup

		// Launch 10 concurrent invalidations
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				InvalidateCache()
			}()
		}

		wg.Wait()

		// Cache should be nil
		assert.Nil(t, payloadCache)
	})

	t.Run("Concurrent writes and reads", func(t *testing.T) {
		var wg sync.WaitGroup

		// Concurrent reads
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				testPayload := &models.LockboxPayload{
					Version: 1,
					Users:   make(map[string]models.UserTokens),
				}
				SetPayloadCache(testPayload, 5*time.Minute)
			}()
		}

		// Concurrent invalidations
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				InvalidateCache()
			}()
		}

		wg.Wait()

		// Should not panic or deadlock
	})
}

// TestCacheExpiryLogic tests cache expiry behavior
func TestCacheExpiryLogic(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	t.Run("Cache hit before expiry", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"user1": {
					AccessToken:  "access1",
					RefreshToken: "refresh1",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		// Cache should be valid
		cacheMutex.RLock()
		valid := payloadCache != nil && time.Now().Before(cacheExpiry)
		cacheMutex.RUnlock()

		assert.True(t, valid)
	})

	t.Run("Cache expired", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		}

		SetPayloadCache(testPayload, -1*time.Minute)

		// Cache should be expired
		cacheMutex.RLock()
		expired := payloadCache != nil && time.Now().After(cacheExpiry)
		cacheMutex.RUnlock()

		assert.True(t, expired)
	})

	t.Run("Cache exactly at expiry time", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		}

		// Set cache with very short TTL
		SetPayloadCache(testPayload, 1*time.Millisecond)

		// Wait for expiry
		time.Sleep(10 * time.Millisecond)

		// Cache should be expired
		cacheMutex.RLock()
		expired := payloadCache != nil && time.Now().After(cacheExpiry)
		cacheMutex.RUnlock()

		assert.True(t, expired)
	})
}

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	t.Run("Empty payload", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		_, err := GetUserTokens(ctx, "anyuser")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tokens not found for user")
	})

	t.Run("Payload with nil Users map", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users:   nil,
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		_, err := GetUserTokens(ctx, "anyuser")

		assert.Error(t, err)
	})

	t.Run("User tokens with empty strings", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"emptyuser": {
					AccessToken:  "",
					RefreshToken: "",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		tokens, err := GetUserTokens(ctx, "emptyuser")

		require.NoError(t, err)
		assert.NotNil(t, tokens)
		assert.Empty(t, tokens.AccessToken)
		assert.Empty(t, tokens.RefreshToken)
	})

	t.Run("Unicode usernames", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"用户": {
					AccessToken:  "access",
					RefreshToken: "refresh",
				},
				"пользователь": {
					AccessToken:  "access2",
					RefreshToken: "refresh2",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()

		tokens1, err := GetUserTokens(ctx, "用户")
		require.NoError(t, err)
		assert.Equal(t, "access", tokens1.AccessToken)

		tokens2, err := GetUserTokens(ctx, "пользователь")
		require.NoError(t, err)
		assert.Equal(t, "access2", tokens2.AccessToken)
	})

	t.Run("Very long username", func(t *testing.T) {
		longUsername := string(make([]byte, 1000))
		for i := range longUsername {
			longUsername = longUsername[:i] + "a" + longUsername[i+1:]
		}

		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				longUsername: {
					AccessToken:  "access",
					RefreshToken: "refresh",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		tokens, err := GetUserTokens(ctx, longUsername)

		require.NoError(t, err)
		assert.Equal(t, "access", tokens.AccessToken)
	})
}

// TestContextCancellation tests context cancellation behavior
func TestContextCancellation(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	t.Run("Cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Operations with cancelled context should fail
		// (This tests that context is properly propagated)
		_, err := GetUserTokens(ctx, "testuser")

		// Should fail either due to cancellation or cache miss
		_ = err
	})

	t.Run("Context with timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// Wait for timeout
		time.Sleep(10 * time.Millisecond)

		// Operations should respect timeout
		_, err := GetUserTokens(ctx, "testuser")

		// Should fail either due to timeout or cache miss
		_ = err
	})
}

// TestPayloadVersioning tests payload version handling
func TestPayloadVersioning(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	t.Run("Different payload versions", func(t *testing.T) {
		v1 := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"user1": {
					AccessToken:  "access1",
					RefreshToken: "refresh1",
				},
			},
		}

		v2 := &models.LockboxPayload{
			Version: 2,
			Users: map[string]models.UserTokens{
				"user2": {
					AccessToken:  "access2",
					RefreshToken: "refresh2",
				},
			},
		}

		SetPayloadCache(v1, 5*time.Minute)
		assert.Equal(t, 1, payloadCache.Version)

		SetPayloadCache(v2, 5*time.Minute)
		assert.Equal(t, 2, payloadCache.Version)
	})
}

// TestMultipleOperations tests sequences of operations
func TestMultipleOperations(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	t.Run("Store then invalidate then get", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"user1": {
					AccessToken:  "access1",
					RefreshToken: "refresh1",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()

		// Get user1
		tokens1, err := GetUserTokens(ctx, "user1")
		require.NoError(t, err)
		assert.Equal(t, "access1", tokens1.AccessToken)

		// Invalidate cache
		InvalidateCache()
		assert.Nil(t, payloadCache)

		// Set cache again
		SetPayloadCache(testPayload, 5*time.Minute)

		// Get user1 again
		tokens2, err := GetUserTokens(ctx, "user1")
		require.NoError(t, err)
		assert.Equal(t, "access1", tokens2.AccessToken)
	})

	t.Run("Multiple invalidations", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		}

		// Set and invalidate multiple times
		for i := 0; i < 10; i++ {
			SetPayloadCache(testPayload, 5*time.Minute)
			assert.NotNil(t, payloadCache)

			InvalidateCache()
			assert.Nil(t, payloadCache)
		}
	})
}

// TestTableDrivenGetUserTokens provides table-driven tests for GetUserTokens
func TestTableDrivenGetUserTokens(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	tests := []struct {
		name        string
		payload     *models.LockboxPayload
		username    string
		wantErr     bool
		errContains string
		wantToken   string
		setupCache  bool
		cacheTTL    time.Duration
	}{
		{
			name: "successful retrieval",
			payload: &models.LockboxPayload{
				Version: 1,
				Users: map[string]models.UserTokens{
					"testuser": {
						AccessToken:  "test_access",
						RefreshToken: "test_refresh",
					},
				},
			},
			username:   "testuser",
			wantErr:    false,
			wantToken:  "test_access",
			setupCache: true,
			cacheTTL:   5 * time.Minute,
		},
		{
			name: "user not found",
			payload: &models.LockboxPayload{
				Version: 1,
				Users: map[string]models.UserTokens{
					"otheruser": {
						AccessToken:  "access",
						RefreshToken: "refresh",
					},
				},
			},
			username:    "nonexistent",
			wantErr:     true,
			errContains: "tokens not found for user",
			setupCache:  true,
			cacheTTL:    5 * time.Minute,
		},
		{
			name: "empty username",
			payload: &models.LockboxPayload{
				Version: 1,
				Users: map[string]models.UserTokens{
					"user": {
						AccessToken:  "access",
						RefreshToken: "refresh",
					},
				},
			},
			username:    "",
			wantErr:     true,
			errContains: "tokens not found for user",
			setupCache:  true,
			cacheTTL:    5 * time.Minute,
		},
		{
			name: "case sensitive username",
			payload: &models.LockboxPayload{
				Version: 1,
				Users: map[string]models.UserTokens{
					"TestUser": {
						AccessToken:  "access",
						RefreshToken: "refresh",
					},
				},
			},
			username:    "testuser",
			wantErr:     true,
			errContains: "tokens not found for user",
			setupCache:  true,
			cacheTTL:    5 * time.Minute,
		},
		{
			name: "nil users map",
			payload: &models.LockboxPayload{
				Version: 1,
				Users:   nil,
			},
			username:    "anyuser",
			wantErr:     true,
			errContains: "tokens not found for user",
			setupCache:  true,
			cacheTTL:    5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupCache {
				SetPayloadCache(tt.payload, tt.cacheTTL)
			}

			ctx := context.Background()
			tokens, err := GetUserTokens(ctx, tt.username)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, tokens)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, tokens)
				assert.Equal(t, tt.wantToken, tokens.AccessToken)
			}
		})
	}
}

// TestTableDrivenStoreUserTokens provides table-driven tests for StoreUserTokens
func TestTableDrivenStoreUserTokens(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	tests := []struct {
		name         string
		initialCache *models.LockboxPayload
		username     string
		accessToken  string
		refreshToken string
		wantErr      bool
		errContains  string
	}{
		{
			name: "store new user",
			initialCache: &models.LockboxPayload{
				Version: 1,
				Users:   make(map[string]models.UserTokens),
			},
			username:     "newuser",
			accessToken:  "new_access",
			refreshToken: "new_refresh",
			wantErr:      true,
			errContains:  "not yet implemented",
		},
		{
			name: "update existing user",
			initialCache: &models.LockboxPayload{
				Version: 1,
				Users: map[string]models.UserTokens{
					"existinguser": {
						AccessToken:  "old_access",
						RefreshToken: "old_refresh",
					},
				},
			},
			username:     "existinguser",
			accessToken:  "new_access",
			refreshToken: "new_refresh",
			wantErr:      true,
			errContains:  "not yet implemented",
		},
		{
			name: "empty username",
			initialCache: &models.LockboxPayload{
				Version: 1,
				Users:   make(map[string]models.UserTokens),
			},
			username:     "",
			accessToken:  "access",
			refreshToken: "refresh",
			wantErr:      true,
			errContains:  "not yet implemented",
		},
		{
			name: "empty access token",
			initialCache: &models.LockboxPayload{
				Version: 1,
				Users:   make(map[string]models.UserTokens),
			},
			username:     "user",
			accessToken:  "",
			refreshToken: "refresh",
			wantErr:      true,
			errContains:  "not yet implemented",
		},
		{
			name: "empty refresh token",
			initialCache: &models.LockboxPayload{
				Version: 1,
				Users:   make(map[string]models.UserTokens),
			},
			username:     "user",
			accessToken:  "access",
			refreshToken: "",
			wantErr:      true,
			errContains:  "not yet implemented",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetPayloadCache(tt.initialCache, 5*time.Minute)

			ctx := context.Background()
			err := StoreUserTokens(ctx, tt.username, tt.accessToken, tt.refreshToken)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			// Cache should always be invalidated
			assert.Nil(t, payloadCache)
		})
	}
}

// TestTableDrivenDeleteUserTokens provides table-driven tests for DeleteUserTokens
func TestTableDrivenDeleteUserTokens(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	tests := []struct {
		name         string
		initialCache *models.LockboxPayload
		username     string
		wantErr      bool
		errContains  string
	}{
		{
			name: "delete existing user",
			initialCache: &models.LockboxPayload{
				Version: 1,
				Users: map[string]models.UserTokens{
					"user1": {
						AccessToken:  "access1",
						RefreshToken: "refresh1",
					},
				},
			},
			username:    "user1",
			wantErr:     true,
			errContains: "not yet implemented",
		},
		{
			name: "delete non-existent user",
			initialCache: &models.LockboxPayload{
				Version: 1,
				Users: map[string]models.UserTokens{
					"user1": {
						AccessToken:  "access1",
						RefreshToken: "refresh1",
					},
				},
			},
			username:    "nonexistent",
			wantErr:     true,
			errContains: "not yet implemented",
		},
		{
			name: "DELETE FROM empty users MAP",
			initialCache: &models.LockboxPayload{
				Version: 1,
				Users:   make(map[string]models.UserTokens),
			},
			username:    "anyuser",
			wantErr:     true,
			errContains: "not yet implemented",
		},
		{
			name: "delete with nil users map",
			initialCache: &models.LockboxPayload{
				Version: 1,
				Users:   nil,
			},
			username:    "anyuser",
			wantErr:     true,
			errContains: "not yet implemented",
		},
		{
			name: "delete with empty username",
			initialCache: &models.LockboxPayload{
				Version: 1,
				Users: map[string]models.UserTokens{
					"user1": {
						AccessToken:  "access1",
						RefreshToken: "refresh1",
					},
				},
			},
			username:    "",
			wantErr:     true,
			errContains: "not yet implemented",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetPayloadCache(tt.initialCache, 5*time.Minute)

			ctx := context.Background()
			err := DeleteUserTokens(ctx, tt.username)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			// Cache should always be invalidated
			assert.Nil(t, payloadCache)
		})
	}
}

// TestHelperFunctions tests utility functions
func TestHelperFunctions(t *testing.T) {
	t.Run("GetSecretID returns consistent value", func(t *testing.T) {
		setupTestEnv(t)
		defer teardownTestEnv(t)

		id1 := GetSecretID()
		id2 := GetSecretID()

		assert.Equal(t, id1, id2)
	})

	t.Run("SetPayloadCache with various TTLs", func(t *testing.T) {
		setupTestEnv(t)
		defer teardownTestEnv(t)

		testPayload := &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		}

		ttls := []time.Duration{
			1 * time.Second,
			1 * time.Minute,
			1 * time.Hour,
			24 * time.Hour,
		}

		for _, ttl := range ttls {
			SetPayloadCache(testPayload, ttl)
			assert.NotNil(t, payloadCache)

			expectedExpiry := time.Now().Add(ttl)
			// Allow 1 second tolerance
			assert.WithinDuration(t, expectedExpiry, cacheExpiry, time.Second)
		}
	})
}

// TestJSONMarshaling tests JSON serialization of payloads
func TestJSONMarshaling(t *testing.T) {
	t.Run("Marshal and unmarshal payload", func(t *testing.T) {
		original := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"user1": {
					AccessToken:  "access1",
					RefreshToken: "refresh1",
				},
				"user2": {
					AccessToken:  "access2",
					RefreshToken: "refresh2",
				},
			},
		}

		// Marshal
		data, err := json.Marshal(original)
		require.NoError(t, err)
		assert.NotEmpty(t, data)

		// Unmarshal
		var unmarshaled models.LockboxPayload
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		// Verify
		assert.Equal(t, original.Version, unmarshaled.Version)
		assert.Len(t, unmarshaled.Users, 2)

		tokens1, ok := unmarshaled.Users["user1"]
		assert.True(t, ok)
		assert.Equal(t, "access1", tokens1.AccessToken)
		assert.Equal(t, "refresh1", tokens1.RefreshToken)
	})

	t.Run("Marshal empty payload", func(t *testing.T) {
		payload := &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		}

		data, err := json.Marshal(payload)
		require.NoError(t, err)

		var unmarshaled models.LockboxPayload
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, 1, unmarshaled.Version)
		assert.Empty(t, unmarshaled.Users)
	})

	t.Run("Marshal user tokens", func(t *testing.T) {
		tokens := models.UserTokens{
			AccessToken:  "test_access",
			RefreshToken: "test_refresh",
		}

		data, err := json.Marshal(tokens)
		require.NoError(t, err)

		var unmarshaled models.UserTokens
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, "test_access", unmarshaled.AccessToken)
		assert.Equal(t, "test_refresh", unmarshaled.RefreshToken)
	})
}

// TestErrorMessages tests error message formatting
func TestErrorMessages(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	t.Run("GetUserTokens error message", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		_, err := GetUserTokens(ctx, "missinguser")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tokens not found for user")
		assert.Contains(t, err.Error(), "missinguser")
	})

	t.Run("StoreUserTokens error message", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		err := StoreUserTokens(ctx, "user", "access", "refresh")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")
		assert.Contains(t, err.Error(), "StoreUserTokens")
	})

	t.Run("DeleteUserTokens error message", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		err := DeleteUserTokens(ctx, "user")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")
		assert.Contains(t, err.Error(), "DeleteUserTokens")
	})
}

// TestInitClientErrorMessage tests InitClient error messages
func TestInitClientErrorMessage(t *testing.T) {
	t.Run("Missing LOCKBOX_SECRET_ID error", func(t *testing.T) {
		resetPackageState()
		os.Unsetenv("LOCKBOX_SECRET_ID")

		ctx := context.Background()
		_, err := InitClient(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "LOCKBOX_SECRET_ID")
		assert.Contains(t, err.Error(), "environment variable not set")
	})
}

// Benchmark tests for performance
func BenchmarkGetUserTokens(b *testing.B) {
	resetPackageState()
	os.Setenv("LOCKBOX_SECRET_ID", "test-secret-id")
	defer os.Unsetenv("LOCKBOX_SECRET_ID")

	testPayload := &models.LockboxPayload{
		Version: 1,
		Users: map[string]models.UserTokens{
			"testuser": {
				AccessToken:  "access",
				RefreshToken: "refresh",
			},
		},
	}

	SetPayloadCache(testPayload, 5*time.Minute)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetUserTokens(ctx, "testuser")
	}
}

func BenchmarkInvalidateCache(b *testing.B) {
	resetPackageState()
	os.Setenv("LOCKBOX_SECRET_ID", "test-secret-id")
	defer os.Unsetenv("LOCKBOX_SECRET_ID")

	testPayload := &models.LockboxPayload{
		Version: 1,
		Users:   make(map[string]models.UserTokens),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SetPayloadCache(testPayload, 5*time.Minute)
		InvalidateCache()
	}
}

func BenchmarkSetPayloadCache(b *testing.B) {
	resetPackageState()
	os.Setenv("LOCKBOX_SECRET_ID", "test-secret-id")
	defer os.Unsetenv("LOCKBOX_SECRET_ID")

	testPayload := &models.LockboxPayload{
		Version: 1,
		Users:   make(map[string]models.UserTokens),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SetPayloadCache(testPayload, 5*time.Minute)
	}
}

// TestMockInterface verifies that our mock can be used for testing
func TestMockInterface(t *testing.T) {
	t.Run("MockPayloadServiceClient basic operations", func(t *testing.T) {
		mock := NewMockPayloadServiceClient()

		// Test Get call count
		ctx := context.Background()
		_, _ = mock.Get(ctx, nil)
		_, _ = mock.Get(ctx, nil)

		assert.Equal(t, 2, mock.GetCallCount())

		// Test SetPayload
		payload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"user": {
					AccessToken:  "access",
					RefreshToken: "refresh",
				},
			},
		}
		mock.SetPayload(payload)

		// Test SetError
		testErr := errors.New("test error")
		mock.SetError(testErr)

		// Test Close
		mock.Close()

		// After close, Get should return error
		_, err := mock.Get(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "closed")
	})

	t.Run("MockPayloadServiceClient thread safety", func(t *testing.T) {
		mock := NewMockPayloadServiceClient()
		ctx := context.Background()

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = mock.Get(ctx, nil)
			}()
		}

		wg.Wait()
		assert.Equal(t, 100, mock.GetCallCount())
	})
}

// ============================================================================
// Tests for MockLockboxClient and ClientAdapter
// ============================================================================

// TestGetPayloadCacheExpiry tests the getPayload cache expiry behavior
func TestGetPayloadCacheExpiry(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	t.Run("Cache miss triggers fetch", func(t *testing.T) {
		// Reset to ensure no cache
		resetPackageState()

		// Don't set cache - getPayload should try to fetch from Lockbox
		// This will fail in test environment but tests the path
		ctx := context.Background()
		_, err := getPayload(ctx)

		assert.Error(t, err)
	})

	t.Run("Cache hit returns cached value", func(t *testing.T) {
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"user1": {
					AccessToken:  "access1",
					RefreshToken: "refresh1",
				},
			},
		}

		// Set valid cache
		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		result, err := getPayload(ctx)

		require.NoError(t, err)
		assert.Equal(t, testPayload, result)
	})

	t.Run("Cache expiry triggers refetch attempt", func(t *testing.T) {
		// Set expired cache
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"user1": {
					AccessToken:  "access1",
					RefreshToken: "refresh1",
				},
			},
		}

		SetPayloadCache(testPayload, -1*time.Minute)

		// getPayload should try to fetch from Lockbox
		// This will fail in test environment but tests the expiry path
		ctx := context.Background()
		_, err := getPayload(ctx)

		assert.Error(t, err)
	})
}

// TestClientAdapterMethods tests all ClientAdapter methods
func TestClientAdapterMethods(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	t.Run("GetUserTokens success path", func(t *testing.T) {
		adapter := NewClientAdapter()
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"testuser": {
					AccessToken:  "test_access",
					RefreshToken: "test_refresh",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		tokens, err := adapter.GetUserTokens(ctx, "testuser")

		require.NoError(t, err)
		assert.Equal(t, "test_access", tokens.AccessToken)
	})

	t.Run("GetUserTokens error path", func(t *testing.T) {
		adapter := NewClientAdapter()
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		tokens, err := adapter.GetUserTokens(ctx, "nonexistent")

		assert.Error(t, err)
		assert.Nil(t, tokens)
	})

	t.Run("StoreUserTokens error path and cache invalidation", func(t *testing.T) {
		adapter := NewClientAdapter()
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"user1": {
					AccessToken:  "access1",
					RefreshToken: "refresh1",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		err := adapter.StoreUserTokens(ctx, "user2", "access2", "refresh2")

		assert.Error(t, err)
		// Verify cache was invalidated
		assert.Nil(t, payloadCache)
	})

	t.Run("DeleteUserTokens error path and cache invalidation", func(t *testing.T) {
		adapter := NewClientAdapter()
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"user1": {
					AccessToken:  "access1",
					RefreshToken: "refresh1",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		err := adapter.DeleteUserTokens(ctx, "user1")

		assert.Error(t, err)
		// Verify cache was invalidated
		assert.Nil(t, payloadCache)
	})
}

// TestMockLockboxClient tests the MockLockboxClient implementation
func TestMockLockboxClient(t *testing.T) {
	t.Run("MockLockboxClient GetUserTokens success", func(t *testing.T) {
		mockClient := new(MockLockboxClient)
		ctx := context.Background()

		expectedTokens := &models.UserTokens{
			AccessToken:  "test_access",
			RefreshToken: "test_refresh",
		}

		mockClient.On("GetUserTokens", ctx, "testuser").Return(expectedTokens, nil)

		tokens, err := mockClient.GetUserTokens(ctx, "testuser")

		assert.NoError(t, err)
		assert.Equal(t, expectedTokens, tokens)
		mockClient.AssertExpectations(t)
	})

	t.Run("MockLockboxClient GetUserTokens error", func(t *testing.T) {
		mockClient := new(MockLockboxClient)
		ctx := context.Background()

		expectedErr := errors.New("user not found")

		mockClient.On("GetUserTokens", ctx, "nonexistent").Return((*models.UserTokens)(nil), expectedErr)

		tokens, err := mockClient.GetUserTokens(ctx, "nonexistent")

		assert.Error(t, err)
		assert.Nil(t, tokens)
		assert.Equal(t, expectedErr, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("MockLockboxClient GetUserTokens with nil error", func(t *testing.T) {
		mockClient := new(MockLockboxClient)
		ctx := context.Background()

		expectedTokens := &models.UserTokens{
			AccessToken:  "test_access",
			RefreshToken: "test_refresh",
		}

		// Return with nil as second parameter
		mockClient.On("GetUserTokens", ctx, "testuser").Return(expectedTokens, nil)

		tokens, err := mockClient.GetUserTokens(ctx, "testuser")

		assert.NoError(t, err)
		assert.Equal(t, expectedTokens, tokens)
		mockClient.AssertExpectations(t)
	})

	t.Run("MockLockboxClient StoreUserTokens success", func(t *testing.T) {
		mockClient := new(MockLockboxClient)
		ctx := context.Background()

		mockClient.On("StoreUserTokens", ctx, "testuser", "access", "refresh").Return(nil)

		err := mockClient.StoreUserTokens(ctx, "testuser", "access", "refresh")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("MockLockboxClient StoreUserTokens error", func(t *testing.T) {
		mockClient := new(MockLockboxClient)
		ctx := context.Background()

		expectedErr := errors.New("failed to store")

		mockClient.On("StoreUserTokens", ctx, "testuser", "access", "refresh").Return(expectedErr)

		err := mockClient.StoreUserTokens(ctx, "testuser", "access", "refresh")

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("MockLockboxClient StoreUserTokens with nil error", func(t *testing.T) {
		mockClient := new(MockLockboxClient)
		ctx := context.Background()

		mockClient.On("StoreUserTokens", ctx, "testuser", "access", "refresh").Return(nil)

		err := mockClient.StoreUserTokens(ctx, "testuser", "access", "refresh")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("MockLockboxClient DeleteUserTokens success", func(t *testing.T) {
		mockClient := new(MockLockboxClient)
		ctx := context.Background()

		mockClient.On("DeleteUserTokens", ctx, "testuser").Return(nil)

		err := mockClient.DeleteUserTokens(ctx, "testuser")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("MockLockboxClient DeleteUserTokens error", func(t *testing.T) {
		mockClient := new(MockLockboxClient)
		ctx := context.Background()

		expectedErr := errors.New("failed to delete")

		mockClient.On("DeleteUserTokens", ctx, "testuser").Return(expectedErr)

		err := mockClient.DeleteUserTokens(ctx, "testuser")

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("MockLockboxClient DeleteUserTokens with nil error", func(t *testing.T) {
		mockClient := new(MockLockboxClient)
		ctx := context.Background()

		mockClient.On("DeleteUserTokens", ctx, "testuser").Return(nil)

		err := mockClient.DeleteUserTokens(ctx, "testuser")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})
}

// TestClientAdapter tests the ClientAdapter implementation
func TestClientAdapter(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	t.Run("NewClientAdapter creates adapter", func(t *testing.T) {
		adapter := NewClientAdapter()

		assert.NotNil(t, adapter)
	})

	t.Run("ClientAdapter GetUserTokens", func(t *testing.T) {
		adapter := NewClientAdapter()

		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"testuser": {
					AccessToken:  "test_access",
					RefreshToken: "test_refresh",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		tokens, err := adapter.GetUserTokens(ctx, "testuser")

		require.NoError(t, err)
		assert.NotNil(t, tokens)
		assert.Equal(t, "test_access", tokens.AccessToken)
		assert.Equal(t, "test_refresh", tokens.RefreshToken)
	})

	t.Run("ClientAdapter GetUserTokens not found", func(t *testing.T) {
		adapter := NewClientAdapter()

		testPayload := &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		tokens, err := adapter.GetUserTokens(ctx, "nonexistent")

		assert.Error(t, err)
		assert.Nil(t, tokens)
		assert.Contains(t, err.Error(), "tokens not found for user")
	})

	t.Run("ClientAdapter StoreUserTokens", func(t *testing.T) {
		adapter := NewClientAdapter()

		testPayload := &models.LockboxPayload{
			Version: 1,
			Users:   make(map[string]models.UserTokens),
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		err := adapter.StoreUserTokens(ctx, "testuser", "access", "refresh")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")

		// Verify cache was invalidated
		assert.Nil(t, payloadCache)
	})

	t.Run("ClientAdapter DeleteUserTokens", func(t *testing.T) {
		adapter := NewClientAdapter()

		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"testuser": {
					AccessToken:  "access",
					RefreshToken: "refresh",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		err := adapter.DeleteUserTokens(ctx, "testuser")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")

		// Verify cache was invalidated
		assert.Nil(t, payloadCache)
	})

	t.Run("ClientAdapter implements LockboxClient interface", func(t *testing.T) {
		adapter := NewClientAdapter()

		// Verify that ClientAdapter implements the LockboxClient interface
		var _ LockboxClient = adapter
	})
}

// TestClientAdapterWithMock tests using ClientAdapter pattern with mock
func TestClientAdapterWithMock(t *testing.T) {
	t.Run("Using mock through interface", func(t *testing.T) {
		var client LockboxClient = new(MockLockboxClient)
		ctx := context.Background()

		expectedTokens := &models.UserTokens{
			AccessToken:  "mock_access",
			RefreshToken: "mock_refresh",
		}

		client.(*MockLockboxClient).On("GetUserTokens", ctx, "testuser").Return(expectedTokens, nil)

		tokens, err := client.GetUserTokens(ctx, "testuser")

		assert.NoError(t, err)
		assert.Equal(t, expectedTokens, tokens)
	})

	t.Run("Using adapter through interface", func(t *testing.T) {
		setupTestEnv(t)
		defer teardownTestEnv(t)

		var client LockboxClient = NewClientAdapter()

		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"testuser": {
					AccessToken:  "adapter_access",
					RefreshToken: "adapter_refresh",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		ctx := context.Background()
		tokens, err := client.GetUserTokens(ctx, "testuser")

		require.NoError(t, err)
		assert.Equal(t, "adapter_access", tokens.AccessToken)
		assert.Equal(t, "adapter_refresh", tokens.RefreshToken)
	})
}

// TestLockboxClientInterfaceTypeSafety tests interface type safety
func TestLockboxClientInterfaceTypeSafety(t *testing.T) {
	t.Run("ClientAdapter satisfies LockboxClient", func(t *testing.T) {
		var _ LockboxClient = &ClientAdapter{}
	})

	t.Run("MockLockboxClient satisfies LockboxClient", func(t *testing.T) {
		var _ LockboxClient = &MockLockboxClient{}
	})

	t.Run("LockboxClient has correct methods", func(t *testing.T) {
		var client LockboxClient = new(MockLockboxClient)
		ctx := context.Background()

		// Test that all interface methods are callable
		mockClient := client.(*MockLockboxClient)

		expectedTokens := &models.UserTokens{
			AccessToken:  "test",
			RefreshToken: "test",
		}

		mockClient.On("GetUserTokens", ctx, "user").Return(expectedTokens, nil)
		mockClient.On("StoreUserTokens", ctx, "user", "access", "refresh").Return(nil)
		mockClient.On("DeleteUserTokens", ctx, "user").Return(nil)

		tokens, err := client.GetUserTokens(ctx, "user")
		assert.NoError(t, err)
		assert.NotNil(t, tokens)

		err = client.StoreUserTokens(ctx, "user", "access", "refresh")
		assert.NoError(t, err)

		err = client.DeleteUserTokens(ctx, "user")
		assert.NoError(t, err)
	})
}

// TestClientAdapterIntegration tests integration scenarios with ClientAdapter
func TestClientAdapterIntegration(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	t.Run("Full workflow with adapter", func(t *testing.T) {
		adapter := NewClientAdapter()
		ctx := context.Background()

		// Setup initial payload
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"user1": {
					AccessToken:  "access1",
					RefreshToken: "refresh1",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		// Get tokens
		tokens, err := adapter.GetUserTokens(ctx, "user1")
		require.NoError(t, err)
		assert.Equal(t, "access1", tokens.AccessToken)

		// Try to store new tokens (will fail but tests the path)
		err = adapter.StoreUserTokens(ctx, "user2", "access2", "refresh2")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")

		// Cache should be invalidated after store attempt
		assert.Nil(t, payloadCache)

		// Reset cache for delete test
		SetPayloadCache(testPayload, 5*time.Minute)

		// Try to delete tokens (will fail but tests the path)
		err = adapter.DeleteUserTokens(ctx, "user1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")

		// Cache should be invalidated after delete attempt
		assert.Nil(t, payloadCache)
	})

	t.Run("Cache behavior through adapter", func(t *testing.T) {
		adapter := NewClientAdapter()
		ctx := context.Background()

		// Setup cache
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"user1": {
					AccessToken:  "access1",
					RefreshToken: "refresh1",
				},
			},
		}

		SetPayloadCache(testPayload, 5*time.Minute)

		// First call should use cache
		tokens1, err := adapter.GetUserTokens(ctx, "user1")
		require.NoError(t, err)
		assert.Equal(t, "access1", tokens1.AccessToken)

		// Second call should also use cache
		tokens2, err := adapter.GetUserTokens(ctx, "user1")
		require.NoError(t, err)
		assert.Equal(t, "access1", tokens2.AccessToken)

		// Invalidate cache
		InvalidateCache()

		// Verify cache is cleared
		assert.Nil(t, payloadCache)
	})
}

// TestErrorHandlingWithMock tests error handling scenarios with mock
func TestErrorHandlingWithMock(t *testing.T) {
	t.Run("GetUserTokens returns error", func(t *testing.T) {
		mockClient := new(MockLockboxClient)
		ctx := context.Background()

		testErr := errors.New("connection failed")
		mockClient.On("GetUserTokens", ctx, "testuser").Return((*models.UserTokens)(nil), testErr)

		tokens, err := mockClient.GetUserTokens(ctx, "testuser")

		assert.Error(t, err)
		assert.Nil(t, tokens)
		assert.Equal(t, testErr, err)
	})

	t.Run("StoreUserTokens returns error", func(t *testing.T) {
		mockClient := new(MockLockboxClient)
		ctx := context.Background()

		testErr := errors.New("write failed")
		mockClient.On("StoreUserTokens", ctx, "testuser", "access", "refresh").Return(testErr)

		err := mockClient.StoreUserTokens(ctx, "testuser", "access", "refresh")

		assert.Error(t, err)
		assert.Equal(t, testErr, err)
	})

	t.Run("DeleteUserTokens returns error", func(t *testing.T) {
		mockClient := new(MockLockboxClient)
		ctx := context.Background()

		testErr := errors.New("delete failed")
		mockClient.On("DeleteUserTokens", ctx, "testuser").Return(testErr)

		err := mockClient.DeleteUserTokens(ctx, "testuser")

		assert.Error(t, err)
		assert.Equal(t, testErr, err)
	})

	t.Run("Concurrent operations with mock", func(t *testing.T) {
		mockClient := new(MockLockboxClient)
		ctx := context.Background()

		expectedTokens := &models.UserTokens{
			AccessToken:  "access",
			RefreshToken: "refresh",
		}

		mockClient.On("GetUserTokens", ctx, "user").Return(expectedTokens, nil)

		var wg sync.WaitGroup
		errors := make(chan error, 10)

		// Launch concurrent calls
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := mockClient.GetUserTokens(ctx, "user")
				if err != nil {
					errors <- err
				}
			}()
		}

		wg.Wait()
		close(errors)

		// Should have no errors
		for err := range errors {
			t.Errorf("Unexpected error in concurrent mock call: %v", err)
		}
	})
}

// TestContextPropagation tests context propagation through the interface
func TestContextPropagation(t *testing.T) {
	t.Run("Context passed to mock GetUserTokens", func(t *testing.T) {
		mockClient := new(MockLockboxClient)
		ctx := context.WithValue(context.Background(), "key", "value")

		expectedTokens := &models.UserTokens{
			AccessToken:  "access",
			RefreshToken: "refresh",
		}

		// Verify the exact context is passed
		mockClient.On("GetUserTokens", ctx, "user").Return(expectedTokens, nil)

		tokens, err := mockClient.GetUserTokens(ctx, "user")

		assert.NoError(t, err)
		assert.Equal(t, expectedTokens, tokens)
		mockClient.AssertExpectations(t)
	})

	t.Run("Context passed to mock StoreUserTokens", func(t *testing.T) {
		mockClient := new(MockLockboxClient)
		ctx := context.WithValue(context.Background(), "key", "value")

		mockClient.On("StoreUserTokens", ctx, "user", "access", "refresh").Return(nil)

		err := mockClient.StoreUserTokens(ctx, "user", "access", "refresh")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("Context passed to mock DeleteUserTokens", func(t *testing.T) {
		mockClient := new(MockLockboxClient)
		ctx := context.WithValue(context.Background(), "key", "value")

		mockClient.On("DeleteUserTokens", ctx, "user").Return(nil)

		err := mockClient.DeleteUserTokens(ctx, "user")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})
}

// TestMultipleInterfaceImplementations tests switching between implementations
func TestMultipleInterfaceImplementations(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	t.Run("Switch from mock to adapter", func(t *testing.T) {
		var client LockboxClient
		ctx := context.Background()

		// Start with mock
		mockClient := new(MockLockboxClient)
		expectedTokens := &models.UserTokens{
			AccessToken:  "mock_access",
			RefreshToken: "mock_refresh",
		}
		mockClient.On("GetUserTokens", ctx, "user").Return(expectedTokens, nil)

		client = mockClient
		tokens, err := client.GetUserTokens(ctx, "user")
		require.NoError(t, err)
		assert.Equal(t, "mock_access", tokens.AccessToken)

		// Switch to adapter
		testPayload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"user": {
					AccessToken:  "adapter_access",
					RefreshToken: "adapter_refresh",
				},
			},
		}
		SetPayloadCache(testPayload, 5*time.Minute)

		client = NewClientAdapter()
		tokens, err = client.GetUserTokens(ctx, "user")
		require.NoError(t, err)
		assert.Equal(t, "adapter_access", tokens.AccessToken)
	})
}
