package domain

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"goroutine/internal/secrecy"

	"github.com/google/uuid"
)

const (
	ErrInvalidTelegramLinkToken = "Telegram link token must be a valid UUIDv7"
	ErrInvalidTelegramChatID    = "Telegram chat ID must be a non-zero 64-bit integer"
	ErrInvalidTelegramUsername  = "Telegram username must start with '@' and be 5-32 alphanumeric characters or underscores"
	ErrInvalidTelegramToken     = "Telegram bot token must be in format digits:alphanumeric"
	ErrInvalidTelegramMessage   = "Telegram message must be 1-4096 characters"
)

var (
	telegramUsernameRegex = regexp.MustCompile(`^@[a-zA-Z][a-zA-Z0-9_]{4,31}$`)
	telegramTokenRegex    = regexp.MustCompile(`^\d{8,10}:[\w-]{30,}$`)
)

// Valid uuidv7 wrapped in SecretString
type TelegramLinkToken struct {
	secrecy.SecretString
}

func NewTelegramLinkToken(token string) (TelegramLinkToken, error) {
	trimmed := strings.TrimSpace(token)

	u, err := uuid.Parse(trimmed)
	if err != nil || u.Version() != 7 {
		return TelegramLinkToken{}, &ErrValidation{Issues: []string{ErrInvalidTelegramLinkToken}}
	}

	return TelegramLinkToken{SecretString: secrecy.SecretString(trimmed)}, nil
}

// Valid telegram chat id
type TelegramChatID struct {
	value int64
}

func NewTelegramChatID(id int64) (TelegramChatID, error) {
	if id == 0 {
		return TelegramChatID{}, &ErrValidation{Issues: []string{ErrInvalidTelegramChatID}}
	}
	return TelegramChatID{value: id}, nil
}

func ParseTelegramChatID(s string) (TelegramChatID, error) {
	trimmed := strings.TrimSpace(s)
	id, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return TelegramChatID{}, &ErrValidation{Issues: []string{ErrInvalidTelegramChatID}}
	}

	return NewTelegramChatID(id)
}

func (c TelegramChatID) Int64() int64 {
	return c.value
}

func (c TelegramChatID) Value() (driver.Value, error) {
	return c.value, nil
}

func (c TelegramChatID) String() string {
	return strconv.FormatInt(c.value, 10)
}

// Valid telegram username starting with @
type TelegramUsername struct {
	value string
}

func NewTelegramUsername(username string) (TelegramUsername, error) {
	trimmed := strings.TrimSpace(username)
	if !telegramUsernameRegex.MatchString(trimmed) {
		return TelegramUsername{}, &ErrValidation{Issues: []string{ErrInvalidTelegramUsername}}
	}
	return TelegramUsername{value: trimmed}, nil
}

func (u TelegramUsername) String() string {
	return u.value
}

func (u TelegramUsername) Value() (driver.Value, error) {
	return u.value, nil
}

// Valid bot token (digits:alphanumeric), wraps SecretString
type TelegramToken struct {
	secrecy.SecretString
}

func NewTelegramToken(token string) (TelegramToken, error) {
	trimmed := strings.TrimSpace(token)
	if !telegramTokenRegex.MatchString(trimmed) {
		return TelegramToken{}, &ErrValidation{Issues: []string{ErrInvalidTelegramToken}}
	}
	return TelegramToken{SecretString: secrecy.SecretString(trimmed)}, nil
}

// Valid message text (1-4096 chars)
type TelegramMessage struct {
	value string
}

func NewTelegramMessage(text string) (TelegramMessage, error) {
	trimmed := strings.TrimSpace(text)
	if len(trimmed) < 1 || len(trimmed) > 4096 {
		return TelegramMessage{}, &ErrValidation{Issues: []string{ErrInvalidTelegramMessage}}
	}
	return TelegramMessage{value: trimmed}, nil
}

func (m TelegramMessage) String() string {
	return m.value
}

func MustTelegramMessage(text string) TelegramMessage {
	m, err := NewTelegramMessage(text)
	if err != nil {
		panic(fmt.Sprintf("BUG: MustTelegramMessage(%q): %v", text, err))
	}
	return m
}
