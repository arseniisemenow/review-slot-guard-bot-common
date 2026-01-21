package telegram

import (
	"github.com/stretchr/testify/mock"
)

// MockBotSender is a mock implementation of BotSender interface
type MockBotSender struct {
	mock.Mock
}

// NewMockBotSender creates a new mock bot sender
func NewMockBotSender() *MockBotSender {
	return &MockBotSender{}
}

// SendPlainMessage sends a plain text message
func (m *MockBotSender) SendPlainMessage(chatID int64, text string) error {
	args := m.Called(chatID, text)
	return args.Error(0)
}

// SendInlineKeyboardMessage sends a message with inline keyboard buttons
func (m *MockBotSender) SendInlineKeyboardMessage(chatID int64, text string, buttons []InlineKeyboardButton) (int, error) {
	args := m.Called(chatID, text, buttons)
	return args.Int(0), args.Error(1)
}

// SendTwoButtonKeyboard sends a message with two buttons (Approve/Decline pattern)
func (m *MockBotSender) SendTwoButtonKeyboard(chatID int64, text string, approveData, declineData string) (int, error) {
	args := m.Called(chatID, text, approveData, declineData)
	return args.Int(0), args.Error(1)
}

// EditMessage edits an existing message
func (m *MockBotSender) EditMessage(chatID int64, messageID int, text string) error {
	args := m.Called(chatID, messageID, text)
	return args.Error(0)
}

// EditMessageWithKeyboard edits a message and adds a keyboard
func (m *MockBotSender) EditMessageWithKeyboard(chatID int64, messageID int, text string, buttons []InlineKeyboardButton) error {
	args := m.Called(chatID, messageID, text, buttons)
	return args.Error(0)
}

// AnswerCallbackQuery acknowledges a button click
func (m *MockBotSender) AnswerCallbackQuery(callbackQueryID, text string) error {
	args := m.Called(callbackQueryID, text)
	return args.Error(0)
}

// DeleteMessage deletes a message
func (m *MockBotSender) DeleteMessage(chatID int64, messageID int) error {
	args := m.Called(chatID, messageID)
	return args.Error(0)
}
