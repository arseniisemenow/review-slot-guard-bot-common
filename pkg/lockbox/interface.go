package lockbox

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/models"
)

// LockboxClient defines the interface for Lockbox operations
type LockboxClient interface {
	// GetUserTokens retrieves access and refresh tokens for a specific user
	GetUserTokens(ctx context.Context, reviewerLogin string) (*models.UserTokens, error)

	// StoreUserTokens stores new tokens for a user
	StoreUserTokens(ctx context.Context, reviewerLogin, accessToken, refreshToken string) error

	// DeleteUserTokens removes tokens for a user
	DeleteUserTokens(ctx context.Context, reviewerLogin string) error
}

// MockLockboxClient is a mock implementation of LockboxClient for testing
type MockLockboxClient struct {
	mock.Mock
}

// GetUserTokens provides a mock function with given fields: ctx, reviewerLogin
func (_m *MockLockboxClient) GetUserTokens(ctx context.Context, reviewerLogin string) (*models.UserTokens, error) {
	ret := _m.Called(ctx, reviewerLogin)

	if len(ret) == 0 {
		panic("no return value specified for GetUserTokens")
	}

	r0, ok := ret.Get(0).(*models.UserTokens)
	if !ok {
		panic("GetUserTokens return value 0 is not *models.UserTokens")
	}

	var r1 error
	if len(ret) > 1 {
		r1, _ = ret.Get(1).(error)
	}

	return r0, r1
}

// StoreUserTokens provides a mock function with given fields: ctx, reviewerLogin, accessToken, refreshToken
func (_m *MockLockboxClient) StoreUserTokens(ctx context.Context, reviewerLogin, accessToken, refreshToken string) error {
	ret := _m.Called(ctx, reviewerLogin, accessToken, refreshToken)

	if len(ret) == 0 {
		panic("no return value specified for StoreUserTokens")
	}

	var r0 error
	if len(ret) > 0 {
		r0, _ = ret.Get(0).(error)
	}

	return r0
}

// DeleteUserTokens provides a mock function with given fields: ctx, reviewerLogin
func (_m *MockLockboxClient) DeleteUserTokens(ctx context.Context, reviewerLogin string) error {
	ret := _m.Called(ctx, reviewerLogin)

	if len(ret) == 0 {
		panic("no return value specified for DeleteUserTokens")
	}

	var r0 error
	if len(ret) > 0 {
		r0, _ = ret.Get(0).(error)
	}

	return r0
}
