package telegram

import (
	"fmt"
	"os"

	tba "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BotClient wraps Telegram Bot API client
type BotClient struct {
	bot *tba.BotAPI
}

// MessageConfig holds configuration for sending messages
type MessageConfig struct {
	ChatID    int64
	Text      string
	ParseMode string
}

// EditMessageConfig holds configuration for editing messages
type EditMessageConfig struct {
	ChatID    int64
	MessageID int
	Text      string
	ParseMode string
}

// CallbackConfig holds configuration for answering callback queries
type CallbackConfig struct {
	CallbackQueryID string
	Text            string
	ShowAlert       bool
}

// InlineKeyboardButton represents a button in an inline keyboard
type InlineKeyboardButton struct {
	Text string
	Data string
}

// NewBotClient creates a new Telegram bot client
func NewBotClient(token string) (*BotClient, error) {
	bot, err := tba.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %w", err)
	}
	return &BotClient{bot: bot}, nil
}

// NewBotClientFromEnv creates a new Telegram bot client using TELEGRAM_BOT_TOKEN env var
func NewBotClientFromEnv() (*BotClient, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN environment variable not set")
	}
	return NewBotClient(token)
}

// SendPlainMessage sends a plain text message
func (bc *BotClient) SendPlainMessage(chatID int64, text string) error {
	msg := tba.NewMessage(chatID, text)
	_, err := bc.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

// SendInlineKeyboardMessage sends a message with inline keyboard buttons
func (bc *BotClient) SendInlineKeyboardMessage(chatID int64, text string, buttons []InlineKeyboardButton) (int, error) {
	if len(buttons) == 0 {
		return 0, fmt.Errorf("at least one button is required")
	}

	// Create single row keyboard
	row := make([]tba.InlineKeyboardButton, len(buttons))
	for i, btn := range buttons {
		row[i] = tba.NewInlineKeyboardButtonData(btn.Text, btn.Data)
	}

	keyboard := tba.NewInlineKeyboardMarkup(row)
	keyboardPtr := &keyboard

	msg := tba.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboardPtr
	msg.ParseMode = "Markdown"

	sent, err := bc.bot.Send(msg)
	if err != nil {
		return 0, fmt.Errorf("failed to send message with keyboard: %w", err)
	}

	return sent.MessageID, nil
}

// SendTwoButtonKeyboard sends a message with two buttons (Approve/Decline pattern)
func (bc *BotClient) SendTwoButtonKeyboard(chatID int64, text string, approveData, declineData string) (int, error) {
	buttons := []InlineKeyboardButton{
		{Text: "✅ Approve", Data: approveData},
		{Text: "❌ Decline", Data: declineData},
	}
	return bc.SendInlineKeyboardMessage(chatID, text, buttons)
}

// EditMessage edits an existing message
func (bc *BotClient) EditMessage(chatID int64, messageID int, text string) error {
	msg := tba.NewEditMessageText(chatID, messageID, text)
	msg.ParseMode = "Markdown"

	_, err := bc.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to edit message: %w", err)
	}

	return nil
}

// EditMessageWithKeyboard edits a message and adds a keyboard
func (bc *BotClient) EditMessageWithKeyboard(chatID int64, messageID int, text string, buttons []InlineKeyboardButton) error {
	row := make([]tba.InlineKeyboardButton, len(buttons))
	for i, btn := range buttons {
		row[i] = tba.NewInlineKeyboardButtonData(btn.Text, btn.Data)
	}

	keyboard := tba.NewInlineKeyboardMarkup(row)
	keyboardPtr := &keyboard

	msg := tba.NewEditMessageText(chatID, messageID, text)
	msg.ReplyMarkup = keyboardPtr
	msg.ParseMode = "Markdown"

	_, err := bc.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to edit message with keyboard: %w", err)
	}

	return nil
}

// AnswerCallbackQuery acknowledges a button click
func (bc *BotClient) AnswerCallbackQuery(callbackQueryID, text string) error {
	callback := tba.NewCallback(callbackQueryID, text)
	_, err := bc.bot.Send(callback)
	if err != nil {
		return fmt.Errorf("failed to answer callback: %w", err)
	}

	return nil
}

// DeleteMessage deletes a message
func (bc *BotClient) DeleteMessage(chatID int64, messageID int) error {
	msg := tba.NewDeleteMessage(chatID, messageID)
	_, err := bc.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

// GetBot returns the underlying bot API client
func (bc *BotClient) GetBot() *tba.BotAPI {
	return bc.bot
}

// FormatCallbackData creates callback data string
func FormatCallbackData(action, reviewRequestID string) string {
	return fmt.Sprintf("%s:%s", action, reviewRequestID)
}

// ParseCallbackData parses callback data string
func ParseCallbackData(data string) (action, reviewRequestID string, err error) {
	// Expected format: "ACTION:uuid"
	parts := splitData(data, 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid callback data format: %s", data)
	}

	action = parts[0]
	reviewRequestID = parts[1]

	if action != "APPROVE" && action != "DECLINE" {
		return "", "", fmt.Errorf("invalid action: %s", action)
	}

	return action, reviewRequestID, nil
}

// splitData is a helper to split strings
func splitData(s string, n int) []string {
	if n <= 0 {
		return nil
	}

	result := make([]string, 0, n)
	current := ""
	count := 1

	for i := 0; i < len(s); i++ {
		if s[i] == ':' && count < n {
			result = append(result, current)
			current = ""
			count++
		} else {
			current += string(s[i])
		}
	}

	result = append(result, current)
	return result
}
