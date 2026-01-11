package telegram

// BotSender defines the interface for sending Telegram messages
type BotSender interface {
	// SendPlainMessage sends a plain text message
	SendPlainMessage(chatID int64, text string) error

	// SendInlineKeyboardMessage sends a message with inline keyboard buttons
	SendInlineKeyboardMessage(chatID int64, text string, buttons []InlineKeyboardButton) (int, error)

	// SendTwoButtonKeyboard sends a message with two buttons (Approve/Decline pattern)
	SendTwoButtonKeyboard(chatID int64, text string, approveData, declineData string) (int, error)

	// EditMessage edits an existing message
	EditMessage(chatID int64, messageID int, text string) error

	// EditMessageWithKeyboard edits a message and adds a keyboard
	EditMessageWithKeyboard(chatID int64, messageID int, text string, buttons []InlineKeyboardButton) error

	// AnswerCallbackQuery acknowledges a button click
	AnswerCallbackQuery(callbackQueryID, text string) error

	// DeleteMessage deletes a message
	DeleteMessage(chatID int64, messageID int) error
}
