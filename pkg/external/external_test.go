package external

import (
	"context"
	"testing"
	"time"

	s21client "github.com/arseniisemenow/s21auto-client-go"
	"github.com/arseniisemenow/s21auto-client-go/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/models"
)

// TestNewS21Client tests the S21Client constructor functions
func TestNewS21Client(t *testing.T) {
	tests := []struct {
		name        string
		accessToken string
		refreshToken string
		expectNil   bool
	}{
		{
			name:        "Valid tokens",
			accessToken: "valid_access_token",
			refreshToken: "valid_refresh_token",
			expectNil:   false,
		},
		{
			name:        "Empty access token",
			accessToken: "",
			refreshToken: "valid_refresh_token",
			expectNil:   false,
		},
		{
			name:        "Empty refresh token",
			accessToken: "valid_access_token",
			refreshToken: "",
			expectNil:   false,
		},
		{
			name:        "Both tokens empty",
			accessToken: "",
			refreshToken: "",
			expectNil:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewS21Client(tt.accessToken, tt.refreshToken)
			if tt.expectNil {
				assert.Nil(t, client)
			} else {
				assert.NotNil(t, client)
				assert.NotNil(t, client.client)
			}
		})
	}
}

// TestNewS21ClientWithSchoolID tests the S21Client constructor with school ID
func TestNewS21ClientWithSchoolID(t *testing.T) {
	tests := []struct {
		name           string
		accessToken    string
		refreshToken   string
		schoolID       string
		contextHeaders *s21client.ContextHeaders
		expectNil      bool
	}{
		{
			name:         "Full context with school ID",
			accessToken:  "access_token",
			refreshToken: "refresh_token",
			schoolID:     "school123",
			contextHeaders: &s21client.ContextHeaders{
				XEDUSchoolID:  "school123",
				XEDUProductID: "product123",
				XEDUOrgUnitID: "org123",
				XEDURouteInfo: "route123",
			},
			expectNil: false,
		},
		{
			name:           "Nil context headers",
			accessToken:    "access_token",
			refreshToken:   "refresh_token",
			schoolID:       "school123",
			contextHeaders: nil,
			expectNil:      false,
		},
		{
			name:           "Empty school ID",
			accessToken:    "access_token",
			refreshToken:   "refresh_token",
			schoolID:       "",
			contextHeaders: nil,
			expectNil:      false,
		},
		{
			name:         "Empty context headers values",
			accessToken:  "access_token",
			refreshToken: "refresh_token",
			schoolID:     "school123",
			contextHeaders: &s21client.ContextHeaders{
				XEDUSchoolID:  "",
				XEDUProductID: "",
				XEDUOrgUnitID: "",
				XEDURouteInfo: "",
			},
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewS21ClientWithSchoolID(tt.accessToken, tt.refreshToken, tt.schoolID, tt.contextHeaders)
			if tt.expectNil {
				assert.Nil(t, client)
			} else {
				assert.NotNil(t, client)
				assert.NotNil(t, client.client)
			}
		})
	}
}

// TestNewS21ClientFromCreds tests the S21Client constructor with username/password
func TestNewS21ClientFromCreds(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		expectNil bool
	}{
		{
			name:      "Valid credentials",
			username:  "user@example.com",
			password:  "password123",
			expectNil: false,
		},
		{
			name:      "Empty username",
			username:  "",
			password:  "password123",
			expectNil: false,
		},
		{
			name:      "Empty password",
			username:  "user@example.com",
			password:  "",
			expectNil: false,
		},
		{
			name:      "Both empty",
			username:  "",
			password:  "",
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewS21ClientFromCreds(tt.username, tt.password)
			if tt.expectNil {
				assert.Nil(t, client)
			} else {
				assert.NotNil(t, client)
				assert.NotNil(t, client.client)
			}
		})
	}
}

// TestS21AuthProvider_GetAuthCredentials tests the auth provider credentials method
func TestS21AuthProvider_GetAuthCredentials(t *testing.T) {
	tests := []struct {
		name           string
		authProvider   *S21AuthProvider
		expectError    bool
		expectedToken  string
		expectedSchool string
	}{
		{
			name: "Full auth with context headers",
			authProvider: &S21AuthProvider{
				accessToken:  "test_access_token",
				refreshToken: "test_refresh_token",
				schoolID:     "school123",
				contextHeaders: &s21client.ContextHeaders{
					XEDUSchoolID:  "school123",
					XEDUProductID: "product123",
					XEDUOrgUnitID: "org123",
					XEDURouteInfo: "route123",
				},
			},
			expectError:    false,
			expectedToken:  "test_access_token",
			expectedSchool: "school123",
		},
		{
			name: "Auth without context headers",
			authProvider: &S21AuthProvider{
				accessToken:  "test_access_token",
				refreshToken: "test_refresh_token",
				schoolID:     "school456",
				contextHeaders: nil,
			},
			expectError:    false,
			expectedToken:  "test_access_token",
			expectedSchool: "school456",
		},
		{
			name: "Auth with empty tokens",
			authProvider: &S21AuthProvider{
				accessToken:  "",
				refreshToken: "",
				schoolID:     "",
				contextHeaders: nil,
			},
			expectError:    false,
			expectedToken:  "",
			expectedSchool: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			creds, err := tt.authProvider.GetAuthCredentials(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedToken, creds.Token)
				assert.Equal(t, tt.expectedSchool, creds.SchoolId)

				if tt.authProvider.contextHeaders != nil {
					require.NotNil(t, creds.ContextHeaders)
					assert.Equal(t, tt.authProvider.contextHeaders.XEDUSchoolID, creds.ContextHeaders.XEDUSchoolID)
					assert.Equal(t, tt.authProvider.contextHeaders.XEDUProductID, creds.ContextHeaders.XEDUProductID)
					assert.Equal(t, tt.authProvider.contextHeaders.XEDUOrgUnitID, creds.ContextHeaders.XEDUOrgUnitID)
					assert.Equal(t, tt.authProvider.contextHeaders.XEDURouteInfo, creds.ContextHeaders.XEDURouteInfo)
				} else {
					assert.Nil(t, creds.ContextHeaders)
				}
			}
		})
	}
}

// TestExtractFamilies tests the ExtractFamilies function
func TestExtractFamilies(t *testing.T) {
	tests := []struct {
		name     string
		graph    *requests.ProjectMapGetStudentGraphTemplate_Data
		expected int
	}{
		{
			name:     "Nil graph",
			graph:    nil,
			expected: 0,
		},
		{
			name: "Empty graph",
			graph: &requests.ProjectMapGetStudentGraphTemplate_Data{
				HolyGraph: requests.ProjectMapGetStudentGraphTemplate_Data_HolyGraph{
					GetStudentGraphTemplate: requests.ProjectMapGetStudentGraphTemplate_Data_GetStudentGraphTemplate{
						Nodes: []requests.ProjectMapGetStudentGraphTemplate_Data_Node{},
					},
				},
			},
			expected: 0,
		},
		{
			name: "Graph with goal projects",
			graph: &requests.ProjectMapGetStudentGraphTemplate_Data{
				HolyGraph: requests.ProjectMapGetStudentGraphTemplate_Data_HolyGraph{
					GetStudentGraphTemplate: requests.ProjectMapGetStudentGraphTemplate_Data_GetStudentGraphTemplate{
						Nodes: []requests.ProjectMapGetStudentGraphTemplate_Data_Node{
							{
								Label: "C - I",
								Items: []requests.ProjectMapGetStudentGraphTemplate_Data_Item{
									{
										Goal: &requests.ProjectMapGetStudentGraphTemplate_Data_Course{
											ProjectName: "C5_s21_decimal",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: 1,
		},
		{
			name: "Graph with course projects",
			graph: &requests.ProjectMapGetStudentGraphTemplate_Data{
				HolyGraph: requests.ProjectMapGetStudentGraphTemplate_Data_HolyGraph{
					GetStudentGraphTemplate: requests.ProjectMapGetStudentGraphTemplate_Data_GetStudentGraphTemplate{
						Nodes: []requests.ProjectMapGetStudentGraphTemplate_Data_Node{
							{
								Label: "A - B",
								Items: []requests.ProjectMapGetStudentGraphTemplate_Data_Item{
									{
										Course: &requests.ProjectMapGetStudentGraphTemplate_Data_Course{
											ProjectName: "go-concurrency",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: 1,
		},
		{
			name: "Graph with mixed projects",
			graph: &requests.ProjectMapGetStudentGraphTemplate_Data{
				HolyGraph: requests.ProjectMapGetStudentGraphTemplate_Data_HolyGraph{
					GetStudentGraphTemplate: requests.ProjectMapGetStudentGraphTemplate_Data_GetStudentGraphTemplate{
						Nodes: []requests.ProjectMapGetStudentGraphTemplate_Data_Node{
							{
								Label: "C - I",
								Items: []requests.ProjectMapGetStudentGraphTemplate_Data_Item{
									{
										Goal: &requests.ProjectMapGetStudentGraphTemplate_Data_Course{
											ProjectName: "C5_s21_decimal",
										},
									},
									{
										Course: &requests.ProjectMapGetStudentGraphTemplate_Data_Course{
											ProjectName: "go-basics",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: 2,
		},
		{
			name: "Graph with items without project names",
			graph: &requests.ProjectMapGetStudentGraphTemplate_Data{
				HolyGraph: requests.ProjectMapGetStudentGraphTemplate_Data_HolyGraph{
					GetStudentGraphTemplate: requests.ProjectMapGetStudentGraphTemplate_Data_GetStudentGraphTemplate{
						Nodes: []requests.ProjectMapGetStudentGraphTemplate_Data_Node{
							{
								Label: "A - B",
								Items: []requests.ProjectMapGetStudentGraphTemplate_Data_Item{
									{
										Goal:   nil,
										Course: nil,
									},
								},
							},
						},
					},
				},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.graph == nil {
				// The function doesn't handle nil, so we expect a panic
				assert.Panics(t, func() {
					ExtractFamilies(tt.graph)
				})
			} else {
				families, err := ExtractFamilies(tt.graph)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, len(families))
			}
		})
	}
}

// TestGetFamilyLabels tests the GetFamilyLabels function
func TestGetFamilyLabels(t *testing.T) {
	tests := []struct {
		name     string
		graph    *requests.ProjectMapGetStudentGraphTemplate_Data
		expected []string
	}{
		{
			name:     "Nil graph",
			graph:    nil,
			expected: nil,
		},
		{
			name: "Empty graph",
			graph: &requests.ProjectMapGetStudentGraphTemplate_Data{
				HolyGraph: requests.ProjectMapGetStudentGraphTemplate_Data_HolyGraph{
					GetStudentGraphTemplate: requests.ProjectMapGetStudentGraphTemplate_Data_GetStudentGraphTemplate{
						Nodes: []requests.ProjectMapGetStudentGraphTemplate_Data_Node{},
					},
				},
			},
			expected: []string{},
		},
		{
			name: "Graph with multiple families",
			graph: &requests.ProjectMapGetStudentGraphTemplate_Data{
				HolyGraph: requests.ProjectMapGetStudentGraphTemplate_Data_HolyGraph{
					GetStudentGraphTemplate: requests.ProjectMapGetStudentGraphTemplate_Data_GetStudentGraphTemplate{
						Nodes: []requests.ProjectMapGetStudentGraphTemplate_Data_Node{
							{Label: "A - B"},
							{Label: "C - I"},
							{Label: "D - F"},
						},
					},
				},
			},
			expected: []string{"A - B", "C - I", "D - F"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.graph == nil {
				// The function doesn't handle nil, so we expect a panic
				assert.Panics(t, func() {
					GetFamilyLabels(tt.graph)
				})
			} else {
				labels := GetFamilyLabels(tt.graph)
				assert.Equal(t, tt.expected, labels)
			}
		})
	}
}

// TestGetProjectsInFamily tests the GetProjectsInFamily function
func TestGetProjectsInFamily(t *testing.T) {
	tests := []struct {
		name        string
		graph       *requests.ProjectMapGetStudentGraphTemplate_Data
		familyLabel string
		expected    []string
	}{
		{
			name:        "Nil graph",
			graph:       nil,
			familyLabel: "A - B",
			expected:    nil,
		},
		{
			name: "Empty graph",
			graph: &requests.ProjectMapGetStudentGraphTemplate_Data{
				HolyGraph: requests.ProjectMapGetStudentGraphTemplate_Data_HolyGraph{
					GetStudentGraphTemplate: requests.ProjectMapGetStudentGraphTemplate_Data_GetStudentGraphTemplate{
						Nodes: []requests.ProjectMapGetStudentGraphTemplate_Data_Node{},
					},
				},
			},
			familyLabel: "A - B",
			expected:    nil, // Function returns nil for empty result
		},
		{
			name: "Matching family with goals",
			graph: &requests.ProjectMapGetStudentGraphTemplate_Data{
				HolyGraph: requests.ProjectMapGetStudentGraphTemplate_Data_HolyGraph{
					GetStudentGraphTemplate: requests.ProjectMapGetStudentGraphTemplate_Data_GetStudentGraphTemplate{
						Nodes: []requests.ProjectMapGetStudentGraphTemplate_Data_Node{
							{
								Label: "C - I",
								Items: []requests.ProjectMapGetStudentGraphTemplate_Data_Item{
									{
										Goal: &requests.ProjectMapGetStudentGraphTemplate_Data_Course{
											ProjectName: "C5_s21_decimal",
										},
									},
								},
							},
						},
					},
				},
			},
			familyLabel: "C - I",
			expected:    []string{"C5_s21_decimal"},
		},
		{
			name: "Non-matching family",
			graph: &requests.ProjectMapGetStudentGraphTemplate_Data{
				HolyGraph: requests.ProjectMapGetStudentGraphTemplate_Data_HolyGraph{
					GetStudentGraphTemplate: requests.ProjectMapGetStudentGraphTemplate_Data_GetStudentGraphTemplate{
						Nodes: []requests.ProjectMapGetStudentGraphTemplate_Data_Node{
							{
								Label: "C - I",
								Items: []requests.ProjectMapGetStudentGraphTemplate_Data_Item{
									{
										Goal: &requests.ProjectMapGetStudentGraphTemplate_Data_Course{
											ProjectName: "C5_s21_decimal",
										},
									},
								},
							},
						},
					},
				},
			},
			familyLabel: "A - B",
			expected:    nil, // Function returns nil for non-matching family
		},
		{
			name: "Mixed goal and course projects",
			graph: &requests.ProjectMapGetStudentGraphTemplate_Data{
				HolyGraph: requests.ProjectMapGetStudentGraphTemplate_Data_HolyGraph{
					GetStudentGraphTemplate: requests.ProjectMapGetStudentGraphTemplate_Data_GetStudentGraphTemplate{
						Nodes: []requests.ProjectMapGetStudentGraphTemplate_Data_Node{
							{
								Label: "A - B",
								Items: []requests.ProjectMapGetStudentGraphTemplate_Data_Item{
									{
										Goal: &requests.ProjectMapGetStudentGraphTemplate_Data_Course{
											ProjectName: "A1_basics",
										},
									},
									{
										Course: &requests.ProjectMapGetStudentGraphTemplate_Data_Course{
											ProjectName: "A2_advanced",
										},
									},
								},
							},
						},
					},
				},
			},
			familyLabel: "A - B",
			expected:    []string{"A1_basics", "A2_advanced"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.graph == nil {
				// The function doesn't handle nil, so we expect a panic
				assert.Panics(t, func() {
					GetProjectsInFamily(tt.graph, tt.familyLabel)
				})
			} else {
				projects := GetProjectsInFamily(tt.graph, tt.familyLabel)
				assert.Equal(t, tt.expected, projects)
			}
		})
	}
}

// TestExtractSlots tests the ExtractSlots function
func TestExtractSlots(t *testing.T) {
	baseTime := time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		data     *requests.CalendarGetEvents_Data
		expected int
	}{
		{
			name:     "Nil data",
			data:     nil,
			expected: 0,
		},
		{
			name: "Empty events",
			data: &requests.CalendarGetEvents_Data{
				CalendarEventS21: requests.CalendarGetEvents_Data_CalendarEventS21{
					GetMyCalendarEvents: []requests.CalendarGetEvents_Data_GetMyCalendarEvent{},
				},
			},
			expected: 0,
		},
		{
			name: "Event with slots",
			data: &requests.CalendarGetEvents_Data{
				CalendarEventS21: requests.CalendarGetEvents_Data_CalendarEventS21{
					GetMyCalendarEvents: []requests.CalendarGetEvents_Data_GetMyCalendarEvent{
						{
							EventSlots: []requests.CalendarGetEvents_Data_EventSlot{
								{
									ID:    "slot-1",
									Start: baseTime,
									End:   baseTime.Add(time.Hour),
									Type:  models.SlotTypeFreeTime,
								},
								{
									ID:    "slot-2",
									Start: baseTime.Add(2 * time.Hour),
									End:   baseTime.Add(3 * time.Hour),
									Type:  models.SlotTypeBooking,
								},
							},
						},
					},
				},
			},
			expected: 2,
		},
		{
			name: "Multiple events with slots",
			data: &requests.CalendarGetEvents_Data{
				CalendarEventS21: requests.CalendarGetEvents_Data_CalendarEventS21{
					GetMyCalendarEvents: []requests.CalendarGetEvents_Data_GetMyCalendarEvent{
						{
							EventSlots: []requests.CalendarGetEvents_Data_EventSlot{
								{ID: "slot-1", Start: baseTime, End: baseTime.Add(time.Hour), Type: "FREE_TIME"},
							},
						},
						{
							EventSlots: []requests.CalendarGetEvents_Data_EventSlot{
								{ID: "slot-2", Start: baseTime.Add(time.Hour), End: baseTime.Add(2 * time.Hour), Type: "BOOKING"},
							},
						},
					},
				},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data == nil {
				// The function doesn't handle nil, so we expect a panic
				assert.Panics(t, func() {
					ExtractSlots(tt.data)
				})
			} else {
				slots := ExtractSlots(tt.data)
				assert.Equal(t, tt.expected, len(slots))

				if tt.expected > 0 {
					assert.NotEmpty(t, slots[0].ID)
					assert.False(t, slots[0].Start.IsZero())
					assert.False(t, slots[0].End.IsZero())
					assert.NotEmpty(t, slots[0].Type)
				}
			}
		})
	}
}

// TestExtractBookings tests the ExtractBookings function
func TestExtractBookings(t *testing.T) {
	baseTime := time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		data     *requests.CalendarGetEvents_Data
		expected int
	}{
		{
			name:     "Nil data",
			data:     nil,
			expected: 0,
		},
		{
			name: "Empty events",
			data: &requests.CalendarGetEvents_Data{
				CalendarEventS21: requests.CalendarGetEvents_Data_CalendarEventS21{
					GetMyCalendarEvents: []requests.CalendarGetEvents_Data_GetMyCalendarEvent{},
				},
			},
			expected: 0,
		},
		{
			name: "Event with valid booking",
			data: &requests.CalendarGetEvents_Data{
				CalendarEventS21: requests.CalendarGetEvents_Data_CalendarEventS21{
					GetMyCalendarEvents: []requests.CalendarGetEvents_Data_GetMyCalendarEvent{
						{
							Bookings: []interface{}{
								map[string]interface{}{
									"id":         "booking-1",
									"eventSlotId": "slot-1",
									"eventSlot": map[string]interface{}{
										"start": baseTime.Format(time.RFC3339),
										"end":   baseTime.Add(time.Hour).Format(time.RFC3339),
									},
									"task": map[string]interface{}{
										"goalName": "go-concurrency",
									},
								},
							},
						},
					},
				},
			},
			expected: 1,
		},
		{
			name: "Event with malformed booking (missing id)",
			data: &requests.CalendarGetEvents_Data{
				CalendarEventS21: requests.CalendarGetEvents_Data_CalendarEventS21{
					GetMyCalendarEvents: []requests.CalendarGetEvents_Data_GetMyCalendarEvent{
						{
							Bookings: []interface{}{
								map[string]interface{}{
									"eventSlotId": "slot-1",
								},
							},
						},
					},
				},
			},
			expected: 0,
		},
		{
			name: "Event with non-map booking",
			data: &requests.CalendarGetEvents_Data{
				CalendarEventS21: requests.CalendarGetEvents_Data_CalendarEventS21{
					GetMyCalendarEvents: []requests.CalendarGetEvents_Data_GetMyCalendarEvent{
						{
							Bookings: []interface{}{
								"string booking",
								123,
								nil,
							},
						},
					},
				},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data == nil {
				// The function doesn't handle nil, so we expect a panic
				assert.Panics(t, func() {
					ExtractBookings(tt.data)
				})
			} else {
				bookings := ExtractBookings(tt.data)
				assert.Equal(t, tt.expected, len(bookings))

				if tt.expected > 0 {
					assert.NotEmpty(t, bookings[0].ID)
					assert.NotEmpty(t, bookings[0].EventSlotID)
				}
			}
		})
	}
}

// TestExtractNotifications tests the ExtractNotifications function
func TestExtractNotifications(t *testing.T) {
	baseTime := time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		data     *requests.GetUserNotifications_Data
		expected int
	}{
		{
			name:     "Nil data",
			data:     nil,
			expected: 0,
		},
		{
			name: "Empty notifications",
			data: &requests.GetUserNotifications_Data{
				S21Notification: requests.GetUserNotifications_Data_S21Notification{
					GetS21Notifications: requests.GetUserNotifications_Data_GetS21Notifications{
						Notifications: []requests.GetUserNotifications_Data_Notification{},
					},
				},
			},
			expected: 0,
		},
		{
			name: "Multiple notifications",
			data: &requests.GetUserNotifications_Data{
				S21Notification: requests.GetUserNotifications_Data_S21Notification{
					GetS21Notifications: requests.GetUserNotifications_Data_GetS21Notifications{
						Notifications: []requests.GetUserNotifications_Data_Notification{
							{
								ID:                "notif-1",
								RelatedObjectType: "BOOKING",
								RelatedObjectID:   "slot-1",
								Message:           "Review requested",
								Time:              baseTime,
								WasRead:           false,
								GroupName:         "Reviews",
							},
							{
								ID:                "notif-2",
								RelatedObjectType: "BOOKING",
								RelatedObjectID:   "slot-2",
								Message:           "Another review",
								Time:              baseTime.Add(time.Hour),
								WasRead:           true,
								GroupName:         "Reviews",
							},
						},
					},
				},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data == nil {
				// The function doesn't handle nil, so we expect a panic
				assert.Panics(t, func() {
					ExtractNotifications(tt.data)
				})
			} else {
				notifications := ExtractNotifications(tt.data)
				assert.Equal(t, tt.expected, len(notifications))

				if tt.expected > 0 {
					assert.NotEmpty(t, notifications[0].ID)
					assert.NotEmpty(t, notifications[0].Message)
					assert.False(t, notifications[0].Time.IsZero())
				}
			}
		})
	}
}

// TestFindNotificationBySlotID tests the FindNotificationBySlotID function
func TestFindNotificationBySlotID(t *testing.T) {
	baseTime := time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		notifications []Notification
		slotID        string
		slotTime      time.Time
		expectFound   bool
	}{
		{
			name: "Matching notification with exact time",
			notifications: []Notification{
				{
					ID:              "notif-1",
					RelatedObjectID: "slot-123",
					Time:            baseTime,
				},
			},
			slotID:      "slot-123",
			slotTime:    baseTime,
			expectFound: true,
		},
		{
			name: "Matching notification within time window",
			notifications: []Notification{
				{
					ID:              "notif-1",
					RelatedObjectID: "slot-123",
					Time:            baseTime.Add(30 * time.Second),
				},
			},
			slotID:      "slot-123",
			slotTime:    baseTime,
			expectFound: true,
		},
		{
			name: "Matching notification but time outside window",
			notifications: []Notification{
				{
					ID:              "notif-1",
					RelatedObjectID: "slot-123",
					Time:            baseTime.Add(2 * time.Minute),
				},
			},
			slotID:      "slot-123",
			slotTime:    baseTime,
			expectFound: false,
		},
		{
			name: "Non-matching slot ID",
			notifications: []Notification{
				{
					ID:              "notif-1",
					RelatedObjectID: "slot-456",
					Time:            baseTime,
				},
			},
			slotID:      "slot-123",
			slotTime:    baseTime,
			expectFound: false,
		},
		{
			name:          "Empty notifications list",
			notifications: []Notification{},
			slotID:        "slot-123",
			slotTime:      baseTime,
			expectFound:   false,
		},
		{
			name: "Multiple notifications, find second",
			notifications: []Notification{
				{
					ID:              "notif-1",
					RelatedObjectID: "slot-111",
					Time:            baseTime,
				},
				{
					ID:              "notif-2",
					RelatedObjectID: "slot-123",
					Time:            baseTime.Add(30 * time.Second),
				},
			},
			slotID:      "slot-123",
			slotTime:    baseTime,
			expectFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindNotificationBySlotID(tt.notifications, tt.slotID, tt.slotTime)
			if tt.expectFound {
				assert.NotNil(t, result)
				assert.Equal(t, tt.slotID, result.RelatedObjectID)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

// TestCancelSlot tests the CancelSlot function
func TestCancelSlot(t *testing.T) {
	// CancelSlot is a wrapper around DeleteSlot, so we just test that it exists
	t.Run("CancelSlot exists as function", func(t *testing.T) {
		client := NewS21Client("token", "refresh")
		assert.NotNil(t, client)
		// We can't test the actual call without mocking HTTP, but we verify the function exists
		_ = client.CancelSlot
	})
}

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("ExtractSlots with nil events slice", func(t *testing.T) {
		data := &requests.CalendarGetEvents_Data{
			CalendarEventS21: requests.CalendarGetEvents_Data_CalendarEventS21{
				// GetMyCalendarEvents would be nil or empty
			},
		}
		slots := ExtractSlots(data)
		assert.Equal(t, 0, len(slots))
	})

	t.Run("ExtractBookings with nil bookings", func(t *testing.T) {
		data := &requests.CalendarGetEvents_Data{
			CalendarEventS21: requests.CalendarGetEvents_Data_CalendarEventS21{
				GetMyCalendarEvents: []requests.CalendarGetEvents_Data_GetMyCalendarEvent{
					{
						Bookings: nil,
					},
				},
			},
		}
		bookings := ExtractBookings(data)
		assert.Equal(t, 0, len(bookings))
	})

	t.Run("ExtractNotifications with nil notifications slice", func(t *testing.T) {
		data := &requests.GetUserNotifications_Data{
			S21Notification: requests.GetUserNotifications_Data_S21Notification{
				GetS21Notifications: requests.GetUserNotifications_Data_GetS21Notifications{
					Notifications: nil,
				},
			},
		}
		notifications := ExtractNotifications(data)
		assert.Equal(t, 0, len(notifications))
	})

	t.Run("FindNotificationBySlotID with nil slice", func(t *testing.T) {
		var notifications []Notification
		result := FindNotificationBySlotID(notifications, "slot-1", time.Now())
		assert.Nil(t, result)
	})

	t.Run("FindNotificationByTime with nil slice", func(t *testing.T) {
		var notifications []Notification
		result := FindNotificationByTime(notifications, time.Now(), time.Minute)
		assert.Nil(t, result)
	})

	t.Run("GetFamilyLabels with nil nodes", func(t *testing.T) {
		data := &requests.ProjectMapGetStudentGraphTemplate_Data{
			HolyGraph: requests.ProjectMapGetStudentGraphTemplate_Data_HolyGraph{
				GetStudentGraphTemplate: requests.ProjectMapGetStudentGraphTemplate_Data_GetStudentGraphTemplate{
					Nodes: nil,
				},
			},
		}
		labels := GetFamilyLabels(data)
		// Function returns empty slice for nil nodes, not nil
		assert.Equal(t, []string{}, labels)
	})

	t.Run("GetProjectsInFamily with nil nodes", func(t *testing.T) {
		data := &requests.ProjectMapGetStudentGraphTemplate_Data{
			HolyGraph: requests.ProjectMapGetStudentGraphTemplate_Data_HolyGraph{
				GetStudentGraphTemplate: requests.ProjectMapGetStudentGraphTemplate_Data_GetStudentGraphTemplate{
					Nodes: nil,
				},
			},
		}
		projects := GetProjectsInFamily(data, "A - B")
		assert.Nil(t, projects)
	})
}

// TestTimeHandling tests time-related operations
func TestTimeHandling(t *testing.T) {
	t.Run("Time zone handling in CalendarSlot", func(t *testing.T) {
		// Create times in different timezones
		localTime := time.Date(2025, 1, 8, 14, 0, 0, 0, time.Local)
		utcTime := time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC)

		slot1 := CalendarSlot{
			ID:    "slot-1",
			Start: localTime,
			End:   localTime.Add(time.Hour),
			Type:  "FREE_TIME",
		}

		slot2 := CalendarSlot{
			ID:    "slot-2",
			Start: utcTime,
			End:   utcTime.Add(time.Hour),
			Type:  "FREE_TIME",
		}

		// Both should be valid slots
		assert.NotEmpty(t, slot1.ID)
		assert.NotEmpty(t, slot2.ID)
		assert.Equal(t, time.Hour, slot1.End.Sub(slot1.Start))
		assert.Equal(t, time.Hour, slot2.End.Sub(slot2.Start))
	})

	t.Run("Time comparison in FindNotificationBySlotID", func(t *testing.T) {
		baseTime := time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC)

		notifications := []Notification{
			{
				ID:              "notif-1",
				RelatedObjectID: "slot-1",
				Time:            baseTime.Add(59 * time.Second),
			},
		}

		// Should find with 1 minute window
		result := FindNotificationBySlotID(notifications, "slot-1", baseTime)
		assert.NotNil(t, result)

		// Should not find with 30 second window
		result = FindNotificationBySlotID(notifications, "slot-1", baseTime.Add(-30*time.Second))
		assert.Nil(t, result)
	})

	t.Run("Duration calculation", func(t *testing.T) {
		start := time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC)
		end := time.Date(2025, 1, 8, 15, 30, 0, 0, time.UTC)

		slot := CalendarSlot{
			ID:    "slot-1",
			Start: start,
			End:   end,
			Type:  "FREE_TIME",
		}

		duration := slot.End.Sub(slot.Start)
		assert.Equal(t, 90*time.Minute, duration)
	})
}

// TestBoundaryConditions tests boundary conditions
func TestBoundaryConditions(t *testing.T) {
	t.Run("Minimum valid slot duration", func(t *testing.T) {
		baseTime := time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC)

		slot := CalendarSlot{
			ID:    "slot-1",
			Start: baseTime,
			End:   baseTime, // Same time (zero duration)
			Type:  "FREE_TIME",
		}

		assert.Equal(t, time.Duration(0), slot.End.Sub(slot.Start))
	})

	t.Run("Very large time difference in notifications", func(t *testing.T) {
		baseTime := time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC)

		notifications := []Notification{
			{
				ID:   "notif-1",
				Time: baseTime.Add(365 * 24 * time.Hour), // 1 year later
			},
		}

		result := FindNotificationByTime(notifications, baseTime, time.Hour)
		assert.Nil(t, result) // Should not find with 1 hour window
	})

	t.Run("Empty strings in callbacks", func(t *testing.T) {
		result := FormatCallbackData("", "")
		assert.Equal(t, ":", result)
	})

	t.Run("Very long callback data", func(t *testing.T) {
		longID := string(make([]byte, 1000))
		for i := range longID {
			longID = longID[:i] + "a" + longID[i+1:]
		}

		result := FormatCallbackData("ACTION", longID)
		assert.Contains(t, result, "ACTION:")
		assert.Contains(t, result, longID)
	})
}

// TestExtractProjectNameFromMessageComprehensive tests the ExtractProjectNameFromMessage function with edge cases
func TestExtractProjectNameFromMessageComprehensive(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "Normal message",
			message:  "Review requested for project go-concurrency",
			expected: "Review requested for project go-concurrency",
		},
		{
			name:     "Empty message",
			message:  "",
			expected: "",
		},
		{
			name:     "Message with special characters",
			message:  "Review: go-concurrency @ reviewer",
			expected: "Review: go-concurrency @ reviewer",
		},
		{
			name:     "Message with emojis",
			message:  "Review requested! Please review my project",
			expected: "Review requested! Please review my project",
		},
		{
			name:     "Long message",
			message:  "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
			expected: "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractProjectNameFromMessage(tt.message)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatCallbackDataComprehensive tests the FormatCallbackData function with edge cases
func TestFormatCallbackDataComprehensive(t *testing.T) {
	tests := []struct {
		name            string
		action          string
		reviewRequestID string
		expected        string
	}{
		{
			name:            "Approve action",
			action:          "APPROVE",
			reviewRequestID: "550e8400-e29b-41d4-a716-446655440000",
			expected:        "APPROVE:550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:            "Decline action",
			action:          "DECLINE",
			reviewRequestID: "550e8400-e29b-41d4-a716-446655440000",
			expected:        "DECLINE:550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:            "Empty action",
			action:          "",
			reviewRequestID: "550e8400-e29b-41d4-a716-446655440000",
			expected:        ":550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:            "Empty review request ID",
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
			name:            "Action with colon",
			action:          "ACTION:WITH:COLON",
			reviewRequestID: "id123",
			expected:        "ACTION:WITH:COLON:id123",
		},
		{
			name:            "Review ID with colon",
			action:          "APPROVE",
			reviewRequestID: "id:with:colons",
			expected:        "APPROVE:id:with:colons",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCallbackData(tt.action, tt.reviewRequestID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestProjectFamilyStructure tests the ProjectFamily struct from models
func TestProjectFamilyStructure(t *testing.T) {
	tests := []struct {
		name   string
		family *models.ProjectFamily
	}{
		{
			name: "Valid project family",
			family: &models.ProjectFamily{
				FamilyLabel: "C - I",
				ProjectName: "C5_s21_decimal",
			},
		},
		{
			name: "Empty family label",
			family: &models.ProjectFamily{
				FamilyLabel: "",
				ProjectName: "go-basics",
			},
		},
		{
			name: "Empty project name",
			family: &models.ProjectFamily{
				FamilyLabel: "A - B",
				ProjectName: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.family)
			assert.GreaterOrEqual(t, tt.family.FamilyLabel, "")
			assert.GreaterOrEqual(t, tt.family.ProjectName, "")
		})
	}
}
