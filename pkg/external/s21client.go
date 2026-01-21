package external

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	s21client "github.com/arseniisemenow/s21auto-client-go"
	s21auth "github.com/arseniisemenow/s21auto-client-go/auth"
	"github.com/arseniisemenow/s21auto-client-go/requests"
	"github.com/go-resty/resty/v2"

	"github.com/arseniisemenow/review-slot-guard-bot-common/pkg/models"
)

// S21Client wraps the s21auto client with our application logic
type S21Client struct {
	client *s21client.Client
}

// S21AuthProvider implements authentication using stored access token
type S21AuthProvider struct {
	token          s21auth.Token
	schoolID       string
	contextHeaders *s21client.ContextHeaders
	clientID       string // Configurable client_id for token refresh (default: "school21")
}

// refreshTokenWithCustomClientID manually refreshes token using configured client_id
func (provider *S21AuthProvider) refreshTokenWithCustomClientID(ctx context.Context) error {
	// Check if token is still valid (60 second buffer)
	if provider.token.AccessToken != "" && (time.Now().Unix() < provider.token.ExpiryTime-60) {
		return nil // Token still valid, no refresh needed
	}

	// Prepare refresh request
	client := resty.New()

	var formData map[string]string
	if provider.token.RefreshToken != "" {
		// Use refresh token if available
		formData = map[string]string{
			"client_id":     provider.clientID,
			"grant_type":    "refresh_token",
			"refresh_token": provider.token.RefreshToken,
		}
	} else {
		return fmt.Errorf("no refresh token available")
	}

	// Send refresh request
	res, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(formData).
		Post("https://auth.21-school.ru/auth/realms/EduPowerKeycloak/protocol/openid-connect/token")

	if err != nil {
		return fmt.Errorf("token refresh request failed: %w", err)
	}

	if !res.IsSuccess() {
		return fmt.Errorf("token request failed with status %d: %s", res.StatusCode(), res.String())
	}

	// Parse response
	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}
	if err := json.Unmarshal(res.Body(), &tokenResponse); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	// Update token
	provider.token.AccessToken = tokenResponse.AccessToken
	provider.token.RefreshToken = tokenResponse.RefreshToken
	provider.token.ExpiryTime = time.Now().Unix() + tokenResponse.ExpiresIn
	provider.token.IssueTime = time.Now().Unix()

	return nil
}

func (provider *S21AuthProvider) refreshCredentials(ctx context.Context) error {
	if err := provider.refreshTokenWithCustomClientID(ctx); err != nil {
		return err
	}

	if provider.schoolID == "" {
		user, err := s21auth.RequestUserData(provider.token, ctx)

		if err != nil {
			return err
		}

		provider.schoolID = user.Roles[0].SchoolID
	}

	if provider.contextHeaders == nil {
		headers, err := s21auth.RequestContextHeaders(provider.token, ctx)
		if err != nil {
			return err
		}
		provider.contextHeaders = &s21client.ContextHeaders{
			XEDUSchoolID:  headers.XEDUSchoolID,
			XEDUProductID: headers.XEDUProductID,
			XEDUOrgUnitID: headers.XEDUOrgUnitID,
			XEDURouteInfo: headers.XEDURouteInfo,
		}
	}

	return nil
}

// GetAuthCredentials implements AuthProvider interface
func (a *S21AuthProvider) GetAuthCredentials(ctx context.Context) (s21client.AuthCredentials, error) {
	err := a.refreshCredentials(ctx)

	if err != nil {
		return s21client.AuthCredentials{}, err
	}

	creds := s21client.AuthCredentials{
		Token:          a.token.AccessToken,
		SchoolId:       a.schoolID,
		ContextHeaders: a.contextHeaders,
	}

	return creds, nil
}

// NewS21Client creates a new S21 client with token-based auth (deprecated - use NewS21ClientFromTokens)
func NewS21Client(accessToken, refreshToken, clientID string) *S21Client {
	if clientID == "" {
		clientID = "school21" // Default value
	}

	auth := &S21AuthProvider{
		token: s21auth.Token{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
		clientID: clientID,
	}

	return &S21Client{
		client: s21client.New(auth),
	}
}

// NewS21ClientFromTokens creates a new S21 client from stored tokens with expiry tracking
func NewS21ClientFromTokens(accessToken, refreshToken string, issueTime, expiryTime int64, clientID string) *S21Client {
	if clientID == "" {
		clientID = "school21" // Default value
	}

	auth := &S21AuthProvider{
		token: s21auth.Token{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			IssueTime:    issueTime,
			ExpiryTime:   expiryTime,
		},
		clientID: clientID,
	}

	return &S21Client{
		client: s21client.New(auth),
	}
}

// NewS21ClientWithSchoolID creates a new S21 client with full auth context
func NewS21ClientWithSchoolID(accessToken, refreshToken, schoolID string, contextHeaders *s21client.ContextHeaders, clientID string) *S21Client {
	if clientID == "" {
		clientID = "school21" // Default value
	}

	auth := &S21AuthProvider{
		token: s21auth.Token{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
		schoolID:       schoolID,
		contextHeaders: contextHeaders,
		clientID:       clientID,
	}

	return &S21Client{
		client: s21client.New(auth),
	}
}

// NewS21ClientFromCreds creates a new S21 client from username/password
func NewS21ClientFromCreds(username, password string) *S21Client {
	auth := s21client.DefaultAuth(username, password)
	return &S21Client{
		client: s21client.New(auth),
	}
}

// GetCalendarEvents fetches calendar events for a user
func (c *S21Client) GetCalendarEvents(ctx context.Context, from, to time.Time) (*requests.CalendarGetEvents_Data, error) {
	vars := requests.CalendarGetEvents_Variables{
		From: from.UTC(),
		To:   to.UTC(),
	}

	resp, err := c.client.R().SetContext(ctx).CalendarGetEvents(vars)
	if err != nil {
		return nil, fmt.Errorf("failed to get calendar events: %w", err)
	}

	return &resp, nil
}

// ChangeEventSlot modifies a calendar slot timing
func (c *S21Client) ChangeEventSlot(ctx context.Context, slotID string, start, end time.Time) error {
	vars := requests.CalendarChangeEventSlot_Variables{
		ID:    slotID,
		Start: start.UTC(),
		End:   end.UTC(),
	}

	_, err := c.client.R().SetContext(ctx).CalendarChangeEventSlot(vars)
	if err != nil {
		return fmt.Errorf("failed to change event slot: %w", err)
	}

	return nil
}

// DeleteSlot deletes a calendar slot
func (c *S21Client) DeleteSlot(ctx context.Context, slotID string) error {
	vars := requests.CalendarDeleteEventSlot_Variables{
		EventSlotID: slotID,
	}

	_, err := c.client.R().SetContext(ctx).CalendarDeleteEventSlot(vars)
	if err != nil {
		return fmt.Errorf("failed to delete slot: %w", err)
	}

	return nil
}

// CancelSlot cancels a booking (same as DeleteSlot)
func (c *S21Client) CancelSlot(ctx context.Context, slotID string) error {
	return c.DeleteSlot(ctx, slotID)
}

// GetNotifications fetches user notifications
func (c *S21Client) GetNotifications(ctx context.Context, offset, limit int64) (*requests.GetUserNotifications_Data, error) {
	vars := requests.GetUserNotifications_Variables{
		Paging: requests.GetUserNotifications_Variables_Paging{
			Offset: offset,
			Limit:  limit,
		},
	}

	resp, err := c.client.R().SetContext(ctx).GetUserNotifications(vars)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	return &resp, nil
}

// GetCurrentUser fetches current authenticated user information
func (c *S21Client) GetCurrentUser(ctx context.Context) (*requests.GetCurrentUser_Data, error) {
	resp, err := c.client.R().SetContext(ctx).GetCurrentUser(requests.GetCurrentUser_Variables{})
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	return &resp, nil
}

// GetProjectGraph fetches project dependency graph
func (c *S21Client) GetProjectGraph(ctx context.Context, studentID string) (*requests.ProjectMapGetStudentGraphTemplate_Data, error) {
	vars := requests.ProjectMapGetStudentGraphTemplate_Variables{
		StudentID: studentID,
	}

	resp, err := c.client.R().SetContext(ctx).ProjectMapGetStudentGraphTemplate(vars)
	if err != nil {
		return nil, fmt.Errorf("failed to get project graph: %w", err)
	}

	return &resp, nil
}

// ExtractFamilies extracts project families from graph response
func ExtractFamilies(graph *requests.ProjectMapGetStudentGraphTemplate_Data) ([]*models.ProjectFamily, error) {
	var families []*models.ProjectFamily

	for _, node := range graph.HolyGraph.GetStudentGraphTemplate.Nodes {
		familyLabel := node.Label

		for _, item := range node.Items {
			var projectName string

			if item.Goal != nil && item.Goal.ProjectName != "" {
				projectName = item.Goal.ProjectName
			} else if item.Course != nil && item.Course.ProjectName != "" {
				projectName = item.Course.ProjectName
			}

			if projectName != "" {
				families = append(families, &models.ProjectFamily{
					FamilyLabel: familyLabel,
					ProjectName: projectName,
				})
			}
		}
	}

	return families, nil
}

// GetFamilyLabels extracts all family labels from graph
func GetFamilyLabels(graph *requests.ProjectMapGetStudentGraphTemplate_Data) []string {
	labels := make([]string, 0, len(graph.HolyGraph.GetStudentGraphTemplate.Nodes))

	for _, node := range graph.HolyGraph.GetStudentGraphTemplate.Nodes {
		labels = append(labels, node.Label)
	}

	return labels
}

// GetProjectsInFamily extracts projects for a specific family
func GetProjectsInFamily(graph *requests.ProjectMapGetStudentGraphTemplate_Data, familyLabel string) []string {
	var projects []string

	for _, node := range graph.HolyGraph.GetStudentGraphTemplate.Nodes {
		if node.Label == familyLabel {
			for _, item := range node.Items {
				var projectName string

				if item.Goal != nil && item.Goal.ProjectName != "" {
					projectName = item.Goal.ProjectName
				} else if item.Course != nil && item.Course.ProjectName != "" {
					projectName = item.Course.ProjectName
				}

				if projectName != "" {
					projects = append(projects, projectName)
				}
			}
			break
		}
	}

	return projects
}

// Authenticate performs authentication with username/password
func Authenticate(ctx context.Context, username, password string) (*models.TokenResponse, error) {
	token, err := s21auth.RequestToken(username, password, ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	var refreshExpiresIn int64 = 0
	if token.RefreshToken != "" {
		refreshExpiresIn = 1
	}

	return &models.TokenResponse{
		AccessToken:      token.AccessToken,
		RefreshToken:     token.RefreshToken,
		ExpiresIn:        token.ExpiryTime - token.IssueTime,
		RefreshExpiresIn: refreshExpiresIn,
		TokenType:        "Bearer",
		IssueTime:        token.IssueTime,
		ExpiryTime:       token.ExpiryTime,
	}, nil
}

// CalendarSlot represents a simplified calendar slot from API response
type CalendarSlot struct {
	ID    string
	Start time.Time
	End   time.Time
	Type  string
}

// CalendarBooking represents a simplified booking from API response
type CalendarBooking struct {
	ID          string
	EventSlotID string
	Start       time.Time
	End         time.Time
	ProjectName string
}

// ExtractSlots extracts free time slots from calendar events
func ExtractSlots(data *requests.CalendarGetEvents_Data) []CalendarSlot {
	var slots []CalendarSlot

	for _, event := range data.CalendarEventS21.GetMyCalendarEvents {
		for _, slot := range event.EventSlots {
			slots = append(slots, CalendarSlot{
				ID:    slot.ID,
				Start: slot.Start,
				End:   slot.End,
				Type:  slot.Type,
			})
		}
	}

	return slots
}

// ExtractBookings extracts bookings from calendar events
func ExtractBookings(data *requests.CalendarGetEvents_Data) []CalendarBooking {
	var bookings []CalendarBooking

	for _, event := range data.CalendarEventS21.GetMyCalendarEvents {
		for _, b := range event.Bookings {
			if bookingMap, ok := b.(map[string]interface{}); ok {
				booking := CalendarBooking{}

				if id, ok := bookingMap["id"].(string); ok {
					booking.ID = id
				}

				if eventSlotID, ok := bookingMap["eventSlotId"].(string); ok {
					booking.EventSlotID = eventSlotID
				}

				if eventSlot, ok := bookingMap["eventSlot"].(map[string]interface{}); ok {
					if start, ok := eventSlot["start"].(string); ok {
						if t, err := time.Parse(time.RFC3339, start); err == nil {
							booking.Start = t
						}
					}
					if end, ok := eventSlot["end"].(string); ok {
						if t, err := time.Parse(time.RFC3339, end); err == nil {
							booking.End = t
						}
					}

					if task, ok := bookingMap["task"].(map[string]interface{}); ok {
						if goalName, ok := task["goalName"].(string); ok {
							booking.ProjectName = goalName
						}
					}
				}

				if booking.ID != "" {
					bookings = append(bookings, booking)
				}
			}
		}
	}

	return bookings
}

// Notification represents a notification from API response
type Notification struct {
	ID                string
	RelatedObjectType string
	RelatedObjectID   string
	Message           string
	Time              time.Time
	WasRead           bool
	GroupName         string
}

// ExtractNotifications extracts notifications from API response
func ExtractNotifications(data *requests.GetUserNotifications_Data) []Notification {
	var notifications []Notification

	for _, n := range data.S21Notification.GetS21Notifications.Notifications {
		notifications = append(notifications, Notification{
			ID:                n.ID,
			RelatedObjectType: n.RelatedObjectType,
			RelatedObjectID:   n.RelatedObjectID,
			Message:           n.Message,
			Time:              n.Time,
			WasRead:           n.WasRead,
			GroupName:         n.GroupName,
		})
	}

	return notifications
}

// FindNotificationBySlotID finds a notification matching a calendar slot ID and time
func FindNotificationBySlotID(notifications []Notification, slotID string, slotTime time.Time) *Notification {
	for _, n := range notifications {
		if n.RelatedObjectID == slotID {
			if n.Time.Sub(slotTime).Abs() < time.Minute {
				return &n
			}
		}
	}

	return nil
}

// FindNotificationByTime finds a notification matching a specific time
func FindNotificationByTime(notifications []Notification, slotTime time.Time, window time.Duration) *Notification {
	for _, n := range notifications {
		if n.Time.Sub(slotTime).Abs() < window {
			return &n
		}
	}

	return nil
}

// ExtractProjectNameFromMessage attempts to extract a project name from notification message
func ExtractProjectNameFromMessage(message string) string {
	return message
}

// FormatCallbackData creates callback data string for Telegram buttons
func FormatCallbackData(action, reviewRequestID string) string {
	return fmt.Sprintf("%s:%s", action, reviewRequestID)
}
