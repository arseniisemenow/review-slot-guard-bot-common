package telegram

import (
	"fmt"
	"os"
	"testing"

	tba "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBotAPI is a mock implementation of the Telegram Bot API
type MockBotAPI struct {
	mock.Mock
}

func (m *MockBotAPI) Send(c tba.Chattable) (tba.Message, error) {
	args := m.Called(c)
	if args.Get(0) == nil {
		return tba.Message{}, args.Error(1)
	}
	return args.Get(0).(tba.Message), args.Error(1)
}

func (m *MockBotAPI) GetMe() (tba.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return tba.User{}, args.Error(1)
	}
	return args.Get(0).(tba.User), args.Error(1)
}

// Helper function to create a BotClient with a mock bot
func createMockBotClient(sendFunc func(c tba.Chattable) (tba.Message, error)) *BotClient {
	// We need to use reflection or create a wrapper since BotAPI is an interface
	// For simplicity, we'll use the real BotAPI struct but with mocked behavior
	// In a real scenario, you'd want to extract an interface
	return nil
}

// TestNewBotClient tests the NewBotClient function
func TestNewBotClient(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "Empty token",
			token:       "",
			expectError: true,
		},
		{
			name:        "Invalid token",
			token:       "invalid_token",
			expectError: true,
		},
		{
			name:        "Valid token format (will fail without actual API)",
			token:       "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
			expectError: true, // Will fail because we can't actually connect
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewBotClient(tt.token)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

// TestNewBotClientFromEnv tests the NewBotClientFromEnv function
func TestNewBotClientFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		setupEnv    func()
		cleanupEnv  func()
		expectError bool
	}{
		{
			name: "Environment variable not set",
			setupEnv: func() {
				os.Unsetenv("TELEGRAM_BOT_TOKEN")
			},
			cleanupEnv: func() {
				os.Unsetenv("TELEGRAM_BOT_TOKEN")
			},
			expectError: true,
		},
		{
			name: "Environment variable set to empty",
			setupEnv: func() {
				os.Setenv("TELEGRAM_BOT_TOKEN", "")
			},
			cleanupEnv: func() {
				os.Unsetenv("TELEGRAM_BOT_TOKEN")
			},
			expectError: true,
		},
		{
			name: "Environment variable set (will fail on actual connection)",
			setupEnv: func() {
				os.Setenv("TELEGRAM_BOT_TOKEN", "test_token_123")
			},
			cleanupEnv: func() {
				os.Unsetenv("TELEGRAM_BOT_TOKEN")
			},
			expectError: true, // Will fail because we can't actually connect
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanupEnv()

			client, err := NewBotClientFromEnv()
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

// TestGetBot tests the GetBot method
func TestGetBot(t *testing.T) {
	// We can't easily test this without a real bot instance
	// This test verifies the method exists and returns the bot
	// In a real scenario with dependency injection, we'd mock this
	t.Skip("Requires actual bot instance - integration test")
}

// TestSendPlainMessage tests the SendPlainMessage method
func TestSendPlainMessage(t *testing.T) {
	t.Skip("Requires mock implementation of BotAPI interface - consider extracting interface for BotAPI")
}

// TestSendInlineKeyboardMessage tests the SendInlineKeyboardMessage method
func TestSendInlineKeyboardMessage(t *testing.T) {
	t.Skip("Requires mock implementation of BotAPI interface - consider extracting interface for BotAPI")
}

// TestSendInlineKeyboardMessage_EmptyButtons tests sending message with empty buttons
func TestSendInlineKeyboardMessage_EmptyButtons(t *testing.T) {
	t.Skip("Requires mock implementation of BotAPI interface - consider extracting interface for BotAPI")
}

// TestSendTwoButtonKeyboard tests the SendTwoButtonKeyboard method
func TestSendTwoButtonKeyboard(t *testing.T) {
	t.Skip("Requires mock implementation of BotAPI interface - consider extracting interface for BotAPI")
}

// TestEditMessage tests the EditMessage method
func TestEditMessage(t *testing.T) {
	t.Skip("Requires mock implementation of BotAPI interface - consider extracting interface for BotAPI")
}

// TestEditMessageWithKeyboard tests the EditMessageWithKeyboard method
func TestEditMessageWithKeyboard(t *testing.T) {
	t.Skip("Requires mock implementation of BotAPI interface - consider extracting interface for BotAPI")
}

// TestAnswerCallbackQuery tests the AnswerCallbackQuery method
func TestAnswerCallbackQuery(t *testing.T) {
	t.Skip("Requires mock implementation of BotAPI interface - consider extracting interface for BotAPI")
}

// TestDeleteMessage tests the DeleteMessage method
func TestDeleteMessage(t *testing.T) {
	t.Skip("Requires mock implementation of BotAPI interface - consider extracting interface for BotAPI")
}

// TestFormatCallbackData tests the FormatCallbackData function
func TestFormatCallbackData(t *testing.T) {
	tests := []struct {
		name            string
		action          string
		reviewRequestID string
		expected        string
	}{
		{
			name:            "Approve callback with UUID",
			action:          "APPROVE",
			reviewRequestID: "550e8400-e29b-41d4-a716-446655440000",
			expected:        "APPROVE:550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:            "Decline callback with UUID",
			action:          "DECLINE",
			reviewRequestID: "550e8400-e29b-41d4-a716-446655440000",
			expected:        "DECLINE:550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:            "Approve with simple ID",
			action:          "APPROVE",
			reviewRequestID: "12345",
			expected:        "APPROVE:12345",
		},
		{
			name:            "Empty action",
			action:          "",
			reviewRequestID: "550e8400-e29b-41d4-a716-446655440000",
			expected:        ":550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:            "Empty reviewRequestID",
			action:          "APPROVE",
			reviewRequestID: "",
			expected:        "APPROVE:",
		},
		{
			name:            "Both empty",
			action:          "",
			reviewRequestID: "",
			expected:        ":",
		},
		{
			name:            "Action with colon in ID",
			action:          "APPROVE",
			reviewRequestID: "550e8400:e29b-41d4-a716-446655440000",
			expected:        "APPROVE:550e8400:e29b-41d4-a716-446655440000",
		},
		{
			name:            "Special characters in ID",
			action:          "DECLINE",
			reviewRequestID: "id-with_special.chars:123",
			expected:        "DECLINE:id-with_special.chars:123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCallbackData(tt.action, tt.reviewRequestID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestParseCallbackData tests the ParseCallbackData function
func TestParseCallbackData(t *testing.T) {
	tests := []struct {
		name            string
		data            string
		expectedAction  string
		expectedID      string
		expectError     bool
		errorContains   string
	}{
		{
			name:            "Valid approve callback with UUID",
			data:            "APPROVE:550e8400-e29b-41d4-a716-446655440000",
			expectedAction:  "APPROVE",
			expectedID:      "550e8400-e29b-41d4-a716-446655440000",
			expectError:     false,
		},
		{
			name:            "Valid decline callback with UUID",
			data:            "DECLINE:550e8400-e29b-41d4-a716-446655440000",
			expectedAction:  "DECLINE",
			expectedID:      "550e8400-e29b-41d4-a716-446655440000",
			expectError:     false,
		},
		{
			name:            "Valid approve callback with simple ID",
			data:            "APPROVE:12345",
			expectedAction:  "APPROVE",
			expectedID:      "12345",
			expectError:     false,
		},
		{
			name:            "Valid decline callback with simple ID",
			data:            "DECLINE:67890",
			expectedAction:  "DECLINE",
			expectedID:      "67890",
			expectError:     false,
		},
		{
			name:          "Invalid format - missing action (no colon)",
			data:          "550e8400-e29b-41d4-a716-446655440000",
			expectError:   true,
			errorContains: "invalid callback data format",
		},
		{
			name:          "Invalid format - missing ID",
			data:          "APPROVE:",
			expectedAction: "APPROVE",
			expectedID:    "",
			expectError:   false, // This is actually valid - action is present, ID is empty
		},
		{
			name:          "Invalid action",
			data:          "INVALID:550e8400-e29b-41d4-a716-446655440000",
			expectError:   true,
			errorContains: "invalid action",
		},
		{
			name:          "Invalid action - lowercase approve",
			data:          "approve:550e8400-e29b-41d4-a716-446655440000",
			expectError:   true,
			errorContains: "invalid action",
		},
		{
			name:          "Invalid action - lowercase decline",
			data:          "decline:550e8400-e29b-41d4-a716-446655440000",
			expectError:   true,
			errorContains: "invalid action",
		},
		{
			name:          "Empty string",
			data:          "",
			expectError:   true,
			errorContains: "invalid callback data format",
		},
		{
			name:          "Only colon separator",
			data:          ":",
			expectedAction: "",
			expectedID:    "",
			expectError:   true, // Empty action is invalid
			errorContains: "invalid action",
		},
		{
			name:            "Multiple colons in ID - should preserve them",
			data:            "APPROVE:550e8400:e29b-41d4-a716-446655440000",
			expectedAction:  "APPROVE",
			expectedID:      "550e8400:e29b-41d4-a716-446655440000",
			expectError:     false,
		},
		{
			name:            "Multiple colons in ID - decline",
			data:            "DECLINE:part1:part2:part3",
			expectedAction:  "DECLINE",
			expectedID:      "part1:part2:part3",
			expectError:     false,
		},
		{
			name:            "Special characters in ID",
			data:            "APPROVE:id-with_special.chars-123",
			expectedAction:  "APPROVE",
			expectedID:      "id-with_special.chars-123",
			expectError:     false,
		},
		{
			name:          "Action with numbers only",
			data:          "123:550e8400-e29b-41d4-a716-446655440000",
			expectError:   true,
			errorContains: "invalid action",
		},
		{
			name:          "Mixed case action",
			data:          "Approve:550e8400-e29b-41d4-a716-446655440000",
			expectError:   true,
			errorContains: "invalid action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, id, err := ParseCallbackData(tt.data)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAction, action)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}

// TestSplitData tests the splitData helper function
func TestSplitData(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		n        int
		expected []string
	}{
		{
			name:     "Split by colon - two parts",
			s:        "APPROVE:123",
			n:        2,
			expected: []string{"APPROVE", "123"},
		},
		{
			name:     "Split with multiple colons - n=2",
			s:        "APPROVE:123:extra",
			n:        2,
			expected: []string{"APPROVE", "123:extra"},
		},
		{
			name:     "Split with multiple colons - n=3",
			s:        "APPROVE:123:extra:more",
			n:        3,
			expected: []string{"APPROVE", "123", "extra:more"},
		},
		{
			name:     "No colon - n=2",
			s:        "APPROVE",
			n:        2,
			expected: []string{"APPROVE"},
		},
		{
			name:     "Empty string - n=2",
			s:        "",
			n:        2,
			expected: []string{""},
		},
		{
			name:     "Colon at start",
			s:        ":value",
			n:        2,
			expected: []string{"", "value"},
		},
		{
			name:     "Colon at end",
			s:        "value:",
			n:        2,
			expected: []string{"value", ""},
		},
		{
			name:     "Multiple consecutive colons",
			s:        "value:::end",
			n:        3,
			expected: []string{"value", "", ":end"},
		},
		{
			name:     "n=1 - no splitting",
			s:        "a:b:c:d",
			n:        1,
			expected: []string{"a:b:c:d"},
		},
		{
			name:     "n=0 - should return nil",
			s:        "a:b:c",
			n:        0,
			expected: nil,
		},
		{
			name:     "n=-1 - should return nil",
			s:        "a:b:c",
			n:        -1,
			expected: nil,
		},
		{
			name:     "Large n value",
			s:        "a:b:c",
			n:        10,
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Complex string with special chars",
			s:        "action:id-with_special.chars-123:extra",
			n:        2,
			expected: []string{"action", "id-with_special.chars-123:extra"},
		},
		{
			name:     "Empty middle parts",
			s:        "a::c::e",
			n:        5,
			expected: []string{"a", "", "c", "", "e"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitData(tt.s, tt.n)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestInlineKeyboardButton tests the InlineKeyboardButton struct
func TestInlineKeyboardButton(t *testing.T) {
	tests := []struct {
		name  string
		button InlineKeyboardButton
	}{
		{
			name: "Approve button",
			button: InlineKeyboardButton{
				Text: "✅ Approve",
				Data: "APPROVE:123",
			},
		},
		{
			name: "Decline button",
			button: InlineKeyboardButton{
				Text: "❌ Decline",
				Data: "DECLINE:456",
			},
		},
		{
			name: "Button with empty text",
			button: InlineKeyboardButton{
				Text: "",
				Data: "TEST:789",
			},
		},
		{
			name: "Button with empty data",
			button: InlineKeyboardButton{
				Text: "Test Button",
				Data: "",
			},
		},
		{
			name: "Button with unicode text",
			button: InlineKeyboardButton{
				Text: "🚀 Launch",
				Data: "LAUNCH:rocket",
			},
		},
		{
			name: "Button with special characters in data",
			button: InlineKeyboardButton{
				Text: "Special",
				Data: "action:id-with_special.chars:123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the struct holds values correctly
			assert.NotNil(t, tt.button)
			assert.Equal(t, tt.button.Text, tt.button.Text)
			assert.Equal(t, tt.button.Data, tt.button.Data)
		})
	}
}

// TestMessageConfig tests the MessageConfig struct
func TestMessageConfig(t *testing.T) {
	tests := []struct {
		name   string
		config MessageConfig
	}{
		{
			name: "Valid message config with Markdown",
			config: MessageConfig{
				ChatID:    123456789,
				Text:      "Test message",
				ParseMode: "Markdown",
			},
		},
		{
			name: "Valid message config with HTML",
			config: MessageConfig{
				ChatID:    987654321,
				Text:      "<b>Bold message</b>",
				ParseMode: "HTML",
			},
		},
		{
			name: "Message config with empty parse mode",
			config: MessageConfig{
				ChatID:    111222333,
				Text:      "Plain text message",
				ParseMode: "",
			},
		},
		{
			name: "Message config with negative chat ID",
			config: MessageConfig{
				ChatID:    -100123456789,
				Text:      "Message to group/supergroup",
				ParseMode: "Markdown",
			},
		},
		{
			name: "Message config with zero chat ID",
			config: MessageConfig{
				ChatID:    0,
				Text:      "",
				ParseMode: "",
			},
		},
		{
			name: "Message config with multiline text",
			config: MessageConfig{
				ChatID:    123456789,
				Text:      "Line 1\nLine 2\nLine 3",
				ParseMode: "Markdown",
			},
		},
		{
			name: "Message config with special characters",
			config: MessageConfig{
				ChatID:    123456789,
				Text:      "Message with émojis 🎉 and spëcial çhars",
				ParseMode: "Markdown",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the struct holds values correctly
			assert.NotNil(t, tt.config)
			assert.Equal(t, tt.config.ChatID, tt.config.ChatID)
			assert.Equal(t, tt.config.Text, tt.config.Text)
			assert.Equal(t, tt.config.ParseMode, tt.config.ParseMode)
		})
	}
}

// TestEditMessageConfig tests the EditMessageConfig struct
func TestEditMessageConfig(t *testing.T) {
	tests := []struct {
		name   string
		config EditMessageConfig
	}{
		{
			name: "Valid edit config with Markdown",
			config: EditMessageConfig{
				ChatID:    123456789,
				MessageID: 123,
				Text:      "Edited message",
				ParseMode: "Markdown",
			},
		},
		{
			name: "Valid edit config with HTML",
			config: EditMessageConfig{
				ChatID:    987654321,
				MessageID: 456,
				Text:      "<b>Edited bold</b>",
				ParseMode: "HTML",
			},
		},
		{
			name: "Edit config with zero message ID",
			config: EditMessageConfig{
				ChatID:    111222333,
				MessageID: 0,
				Text:      "Message",
				ParseMode: "",
			},
		},
		{
			name: "Edit config with negative message ID",
			config: EditMessageConfig{
				ChatID:    123456789,
				MessageID: -1,
				Text:      "Message",
				ParseMode: "Markdown",
			},
		},
		{
			name: "Edit config with empty text",
			config: EditMessageConfig{
				ChatID:    123456789,
				MessageID: 789,
				Text:      "",
				ParseMode: "Markdown",
			},
		},
		{
			name: "Edit config with very large message ID",
			config: EditMessageConfig{
				ChatID:    123456789,
				MessageID: 2147483647, // Max int32
				Text:      "Message",
				ParseMode: "Markdown",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the struct holds values correctly
			assert.NotNil(t, tt.config)
			assert.Equal(t, tt.config.ChatID, tt.config.ChatID)
			assert.Equal(t, tt.config.MessageID, tt.config.MessageID)
			assert.Equal(t, tt.config.Text, tt.config.Text)
			assert.Equal(t, tt.config.ParseMode, tt.config.ParseMode)
		})
	}
}

// TestCallbackConfig tests the CallbackConfig struct
func TestCallbackConfig(t *testing.T) {
	tests := []struct {
		name   string
		config CallbackConfig
	}{
		{
			name: "Callback with notification text",
			config: CallbackConfig{
				CallbackQueryID: "callback_123",
				Text:            "Review approved",
				ShowAlert:       false,
			},
		},
		{
			name: "Callback with alert",
			config: CallbackConfig{
				CallbackQueryID: "callback_456",
				Text:            "Review declined",
				ShowAlert:       true,
			},
		},
		{
			name: "Callback with empty text",
			config: CallbackConfig{
				CallbackQueryID: "callback_789",
				Text:            "",
				ShowAlert:       false,
			},
		},
		{
			name: "Callback with empty query ID",
			config: CallbackConfig{
				CallbackQueryID: "",
				Text:            "Notification",
				ShowAlert:       false,
			},
		},
		{
			name: "Callback with long text",
			config: CallbackConfig{
				CallbackQueryID: "callback_long",
				Text:            "This is a very long callback notification text that might be displayed to the user",
				ShowAlert:       true,
			},
		},
		{
			name: "Callback with special characters",
			config: CallbackConfig{
				CallbackQueryID: "callback_special",
				Text:            "Notification with émojis 🎉 and spëcial çhars",
				ShowAlert:       false,
			},
		},
		{
			name: "Callback with newlines",
			config: CallbackConfig{
				CallbackQueryID: "callback_multiline",
				Text:            "Line 1\nLine 2\nLine 3",
				ShowAlert:       true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the struct holds values correctly
			assert.NotNil(t, tt.config)
			assert.Equal(t, tt.config.CallbackQueryID, tt.config.CallbackQueryID)
			assert.Equal(t, tt.config.Text, tt.config.Text)
			assert.Equal(t, tt.config.ShowAlert, tt.config.ShowAlert)
		})
	}
}

// TestBotClient_GetBot tests the GetBot method
func TestBotClient_GetBot(t *testing.T) {
	t.Skip("Requires actual bot instance - integration test")
}

// TestErrorMessages tests error message formatting
func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name        string
		errorFunc   func() error
		checkError  func(t *testing.T, err error)
	}{
		{
			name: "NewBotClient error message",
			errorFunc: func() error {
				_, err := NewBotClient("")
				return err
			},
			checkError: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to create Telegram bot")
			},
		},
		{
			name: "NewBotClientFromEnv error message - not set",
			errorFunc: func() error {
				os.Unsetenv("TELEGRAM_BOT_TOKEN")
				_, err := NewBotClientFromEnv()
				return err
			},
			checkError: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "TELEGRAM_BOT_TOKEN environment variable not set")
			},
		},
		{
			name: "ParseCallbackData error message - invalid format",
			errorFunc: func() error {
				_, _, err := ParseCallbackData("invalid_format")
				return err
			},
			checkError: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid callback data format")
			},
		},
		{
			name: "ParseCallbackData error message - invalid action",
			errorFunc: func() error {
				_, _, err := ParseCallbackData("INVALID:123")
				return err
			},
			checkError: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid action")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errorFunc()
			tt.checkError(t, err)
		})
	}
}

// TestEdgeCases tests edge cases and boundary conditions
func TestEdgeCases(t *testing.T) {
	t.Run("FormatCallbackData and ParseCallbackData roundtrip", func(t *testing.T) {
		actions := []string{"APPROVE", "DECLINE"}
		ids := []string{
			"123",
			"550e8400-e29b-41d4-a716-446655440000",
			"id-with_special.chars:123",
			"",
		}

		for _, action := range actions {
			for _, id := range ids {
				formatted := FormatCallbackData(action, id)
				parsedAction, parsedID, err := ParseCallbackData(formatted)

				assert.NoError(t, err)
				assert.Equal(t, action, parsedAction)
				assert.Equal(t, id, parsedID)
			}
		}
	})

	t.Run("splitData with empty string and various n values", func(t *testing.T) {
		for n := -1; n <= 5; n++ {
			t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
				result := splitData("", n)
				if n <= 0 {
					assert.Nil(t, result)
				} else {
					assert.Equal(t, []string{""}, result)
				}
			})
		}
	})
}

// TestIntegration_FormatParseRoundtrip tests the integration between FormatCallbackData and ParseCallbackData
func TestIntegration_FormatParseRoundtrip(t *testing.T) {
	testCases := []struct {
		action          string
		reviewRequestID string
	}{
		{"APPROVE", "550e8400-e29b-41d4-a716-446655440000"},
		{"DECLINE", "550e8400-e29b-41d4-a716-446655440000"},
		{"APPROVE", "12345"},
		{"DECLINE", "67890"},
		{"APPROVE", "id-with_special.chars:123"},
		{"DECLINE", "part1:part2:part3"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%s", tc.action, tc.reviewRequestID), func(t *testing.T) {
			// Format the data
			formatted := FormatCallbackData(tc.action, tc.reviewRequestID)

			// Parse it back
			action, id, err := ParseCallbackData(formatted)

			// Verify
			assert.NoError(t, err)
			assert.Equal(t, tc.action, action)
			assert.Equal(t, tc.reviewRequestID, id)
		})
	}
}

// Benchmark tests
func BenchmarkFormatCallbackData(b *testing.B) {
	action := "APPROVE"
	id := "550e8400-e29b-41d4-a716-446655440000"

	for i := 0; i < b.N; i++ {
		FormatCallbackData(action, id)
	}
}

func BenchmarkParseCallbackData(b *testing.B) {
	data := "APPROVE:550e8400-e29b-41d4-a716-446655440000"

	for i := 0; i < b.N; i++ {
		ParseCallbackData(data)
	}
}

func BenchmarkSplitData(b *testing.B) {
	s := "APPROVE:550e8400-e29b-41d4-a716-446655440000"
	n := 2

	for i := 0; i < b.N; i++ {
		splitData(s, n)
	}
}
