package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/syrlramadhan/database-anggota-api/model"
)

type NotificationRepository interface {
	CreateNotification(ctx context.Context, tx *sql.Tx, notification model.Notification) (model.Notification, error)
	GetNotificationsByMemberId(ctx context.Context, tx *sql.Tx, memberID string) ([]model.NotificationWithMember, error)
	GetNotificationById(ctx context.Context, tx *sql.Tx, notificationID string) (model.Notification, error)
	MarkNotificationAsRead(ctx context.Context, tx *sql.Tx, notificationID string, memberID string) error
	GetUnreadNotificationCount(ctx context.Context, tx *sql.Tx, memberID string) (int, error)

	CreateStatusChangeRequest(ctx context.Context, tx *sql.Tx, request model.StatusChangeRequest) (model.StatusChangeRequest, error)
	GetStatusChangeRequestById(ctx context.Context, tx *sql.Tx, requestID string) (model.StatusChangeRequest, error)
	UpdateStatusChangeRequest(ctx context.Context, tx *sql.Tx, requestID string, status string, processedAt time.Time) error
	GetStatusChangeRequestByNotificationId(ctx context.Context, tx *sql.Tx, notificationID string) (model.StatusChangeRequest, error)
}

type notificationRepositoryImpl struct{}

func NewNotificationRepository() NotificationRepository {
	return &notificationRepositoryImpl{}
}

func (n *notificationRepositoryImpl) CreateNotification(ctx context.Context, tx *sql.Tx, notification model.Notification) (model.Notification, error) {
	query := `
		INSERT INTO notifications (id_notification, target_member_id, from_member_id, type, title, message, metadata, pending, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	var metadataStr sql.NullString
	if notification.Metadata.Valid {
		metadataStr = notification.Metadata
	}

	_, err := tx.ExecContext(ctx, query,
		notification.IdNotification,
		notification.TargetMemberID,
		notification.FromMemberID,
		notification.Type,
		notification.Title,
		notification.Message,
		metadataStr,
		notification.Pending,
	)

	if err != nil {
		return model.Notification{}, fmt.Errorf("failed to create notification: %v", err)
	}

	// Return the created notification with timestamps
	notification.CreatedAt = time.Now()
	notification.UpdatedAt = time.Now()

	return notification, nil
}

func (n *notificationRepositoryImpl) GetNotificationsByMemberId(ctx context.Context, tx *sql.Tx, memberID string) ([]model.NotificationWithMember, error) {
	query := `
		SELECT 
			n.id_notification, n.target_member_id, n.from_member_id, n.type, n.title, n.message, 
			n.metadata, n.read_at, n.pending, n.accepted, n.created_at, n.updated_at,
			m.nama as from_member_name, m.nra as from_member_nra, m.role as from_member_role
		FROM notifications n
		LEFT JOIN member m ON n.from_member_id = m.id_member
		WHERE n.target_member_id = ?
		ORDER BY n.created_at DESC
	`

	rows, err := tx.QueryContext(ctx, query, memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %v", err)
	}
	defer rows.Close()

	var notifications []model.NotificationWithMember
	for rows.Next() {
		var notification model.NotificationWithMember
		err := rows.Scan(
			&notification.IdNotification,
			&notification.TargetMemberID,
			&notification.FromMemberID,
			&notification.Type,
			&notification.Title,
			&notification.Message,
			&notification.Metadata,
			&notification.ReadAt,
			&notification.Pending,
			&notification.Accepted,
			&notification.CreatedAt,
			&notification.UpdatedAt,
			&notification.FromMemberName,
			&notification.FromMemberNRA,
			&notification.FromMemberRole,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %v", err)
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

func (n *notificationRepositoryImpl) GetNotificationById(ctx context.Context, tx *sql.Tx, notificationID string) (model.Notification, error) {
	query := `
		SELECT id_notification, target_member_id, from_member_id, type, title, message, 
			   metadata, read_at, pending, accepted, created_at, updated_at
		FROM notifications 
		WHERE id_notification = ?
	`

	var notification model.Notification
	err := tx.QueryRowContext(ctx, query, notificationID).Scan(
		&notification.IdNotification,
		&notification.TargetMemberID,
		&notification.FromMemberID,
		&notification.Type,
		&notification.Title,
		&notification.Message,
		&notification.Metadata,
		&notification.ReadAt,
		&notification.Pending,
		&notification.Accepted,
		&notification.CreatedAt,
		&notification.UpdatedAt,
	)

	if err != nil {
		return model.Notification{}, fmt.Errorf("failed to get notification: %v", err)
	}

	return notification, nil
}

func (n *notificationRepositoryImpl) MarkNotificationAsRead(ctx context.Context, tx *sql.Tx, notificationID string, memberID string) error {
	query := `
		UPDATE notifications 
		SET read_at = NOW(), updated_at = NOW()
		WHERE id_notification = ? AND target_member_id = ?
	`

	result, err := tx.ExecContext(ctx, query, notificationID, memberID)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found or not owned by member")
	}

	return nil
}

func (n *notificationRepositoryImpl) GetUnreadNotificationCount(ctx context.Context, tx *sql.Tx, memberID string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM notifications 
		WHERE target_member_id = ? AND read_at IS NULL
	`

	var count int
	err := tx.QueryRowContext(ctx, query, memberID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread notification count: %v", err)
	}

	return count, nil
}

func (n *notificationRepositoryImpl) CreateStatusChangeRequest(ctx context.Context, tx *sql.Tx, request model.StatusChangeRequest) (model.StatusChangeRequest, error) {
	query := `
		INSERT INTO status_change_requests (id_request, notification_id, target_member_id, requested_by_member_id, from_role, to_role, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW())
	`

	_, err := tx.ExecContext(ctx, query,
		request.IdRequest,
		request.NotificationID,
		request.TargetMemberID,
		request.RequestedByMemberID,
		request.FromRole,
		request.ToRole,
		request.Status,
	)

	if err != nil {
		return model.StatusChangeRequest{}, fmt.Errorf("failed to create status change request: %v", err)
	}

	request.CreatedAt = time.Now()
	return request, nil
}

func (n *notificationRepositoryImpl) GetStatusChangeRequestById(ctx context.Context, tx *sql.Tx, requestID string) (model.StatusChangeRequest, error) {
	query := `
		SELECT id_request, notification_id, target_member_id, requested_by_member_id, 
			   from_role, to_role, status, processed_at, created_at
		FROM status_change_requests 
		WHERE id_request = ?
	`

	var request model.StatusChangeRequest
	err := tx.QueryRowContext(ctx, query, requestID).Scan(
		&request.IdRequest,
		&request.NotificationID,
		&request.TargetMemberID,
		&request.RequestedByMemberID,
		&request.FromRole,
		&request.ToRole,
		&request.Status,
		&request.ProcessedAt,
		&request.CreatedAt,
	)

	if err != nil {
		return model.StatusChangeRequest{}, fmt.Errorf("failed to get status change request: %v", err)
	}

	return request, nil
}

func (n *notificationRepositoryImpl) UpdateStatusChangeRequest(ctx context.Context, tx *sql.Tx, requestID string, status string, processedAt time.Time) error {
	query := `
		UPDATE status_change_requests 
		SET status = ?, processed_at = ?
		WHERE id_request = ?
	`

	_, err := tx.ExecContext(ctx, query, status, processedAt, requestID)
	if err != nil {
		return fmt.Errorf("failed to update status change request: %v", err)
	}

	return nil
}

func (n *notificationRepositoryImpl) GetStatusChangeRequestByNotificationId(ctx context.Context, tx *sql.Tx, notificationID string) (model.StatusChangeRequest, error) {
	query := `
		SELECT id_request, notification_id, target_member_id, requested_by_member_id, 
			   from_role, to_role, status, processed_at, created_at
		FROM status_change_requests 
		WHERE notification_id = ?
	`

	var request model.StatusChangeRequest
	err := tx.QueryRowContext(ctx, query, notificationID).Scan(
		&request.IdRequest,
		&request.NotificationID,
		&request.TargetMemberID,
		&request.RequestedByMemberID,
		&request.FromRole,
		&request.ToRole,
		&request.Status,
		&request.ProcessedAt,
		&request.CreatedAt,
	)

	if err != nil {
		return model.StatusChangeRequest{}, fmt.Errorf("failed to get status change request by notification id: %v", err)
	}

	return request, nil
}
