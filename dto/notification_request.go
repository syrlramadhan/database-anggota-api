package dto

import "time"

// Notification DTOs
type NotificationResponse struct {
	IdNotification string            `json:"id_notification"`
	Type           string            `json:"type"`
	Title          string            `json:"title"`
	Message        string            `json:"message"`
	Pending        bool              `json:"pending"`
	Accepted       *bool             `json:"accepted"`
	ReadAt         *time.Time        `json:"read_at"`
	CreatedAt      time.Time         `json:"created_at"`
	FromMember     FromMemberDetails `json:"from_member"`
	Metadata       interface{}       `json:"metadata,omitempty"`
}

type FromMemberDetails struct {
	IdMember          string `json:"id_member"`
	Nama              string `json:"nama"`
	NRA               string `json:"nra"`
	StatusKeanggotaan string `json:"status_keanggotaan"`
}

type UnreadCountResponse struct {
	UnreadCount int `json:"unread_count"`
}

// Status Change DTOs
type StatusChangeRequestDTO struct {
	TargetMemberID string `json:"target_member_id" validate:"required"`
	FromStatus     string `json:"from_status" validate:"required"`
	ToStatus       string `json:"to_status" validate:"required"`
}

type StatusChangeResponse struct {
	RequestID      string `json:"request_id"`
	NotificationID string `json:"notification_id"`
	Message        string `json:"message"`
}

type StatusChangeAcceptResponse struct {
	Message   string `json:"message"`
	NewStatus string `json:"new_status"`
}

type StatusChangeRejectResponse struct {
	Message string `json:"message"`
}

// Enhanced Member Update Response (for notification integration)
type MemberUpdateWithNotificationResponse struct {
	Member           MemberResponse `json:"member"`
	NotificationSent bool           `json:"notification_sent"`
	NotificationID   string         `json:"notification_id,omitempty"`
}