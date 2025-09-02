package model

import (
	"database/sql"
	"time"
)

type Notification struct {
	IdNotification string         `db:"id_notification"`
	TargetMemberID string         `db:"target_member_id"`
	FromMemberID   string         `db:"from_member_id"`
	Type           string         `db:"type"`
	Title          string         `db:"title"`
	Message        string         `db:"message"`
	Metadata       sql.NullString `db:"metadata"`
	ReadAt         sql.NullTime   `db:"read_at"`
	Pending        bool           `db:"pending"`
	Accepted       sql.NullBool   `db:"accepted"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}

type StatusChangeRequest struct {
	IdRequest           string       `db:"id_request"`
	NotificationID      string       `db:"notification_id"`
	TargetMemberID      string       `db:"target_member_id"`
	RequestedByMemberID string       `db:"requested_by_member_id"`
	FromStatus          string       `db:"from_status"`
	ToStatus            string       `db:"to_status"`
	Status              string       `db:"status"`
	ProcessedAt         sql.NullTime `db:"processed_at"`
	CreatedAt           time.Time    `db:"created_at"`
}

// For JOIN queries with member details
type NotificationWithMember struct {
	Notification
	FromMemberName   string `db:"from_member_name"`
	FromMemberNRA    string `db:"from_member_nra"`
	FromMemberStatus string `db:"from_member_status"`
}
