package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/syrlramadhan/database-anggota-api/dto"
	"github.com/syrlramadhan/database-anggota-api/helper"
	"github.com/syrlramadhan/database-anggota-api/model"
	"github.com/syrlramadhan/database-anggota-api/repository"
)

type NotificationService interface {
	GetNotifications(ctx context.Context, r *http.Request) ([]dto.NotificationResponse, int, error)
	MarkNotificationAsRead(ctx context.Context, r *http.Request, notificationID string) (int, error)
	GetUnreadNotificationCount(ctx context.Context, r *http.Request) (dto.UnreadCountResponse, int, error)

	CreateStatusChangeRequest(ctx context.Context, r *http.Request) (dto.StatusChangeResponse, int, error)
	AcceptStatusChangeRequest(ctx context.Context, r *http.Request, requestID string) (dto.StatusChangeAcceptResponse, int, error)
	RejectStatusChangeRequest(ctx context.Context, r *http.Request, requestID string) (dto.StatusChangeRejectResponse, int, error)

	SendStatusChangeNotification(ctx context.Context, fromMemberID, targetMemberID, fromRole, toRole string) error
}

type notificationServiceImpl struct {
	NotificationRepo repository.NotificationRepository
	MemberRepo       repository.MemberRepository
	DB               *sql.DB
}

func NewNotificationService(notificationRepo repository.NotificationRepository, memberRepo repository.MemberRepository, db *sql.DB) NotificationService {
	return &notificationServiceImpl{
		NotificationRepo: notificationRepo,
		MemberRepo:       memberRepo,
		DB:               db,
	}
}

func (n *notificationServiceImpl) GetNotifications(ctx context.Context, r *http.Request) ([]dto.NotificationResponse, int, error) {
	// Ambil token dari header Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, http.StatusUnauthorized, fmt.Errorf("authorization header is required")
	}

	// Format: "Bearer <token>"
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return nil, http.StatusUnauthorized, fmt.Errorf("invalid authorization header format")
	}

	tokenString := tokenParts[1]

	// Validasi token JWT
	claims, err := helper.ValidateJWT(tokenString)
	if err != nil {
		return nil, http.StatusUnauthorized, fmt.Errorf("invalid or expired token: %v", err)
	}

	// Mulai transaksi database
	tx, err := n.DB.Begin()
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Ambil data member berdasarkan NRA dari claims untuk mendapatkan ID
	member, err := n.MemberRepo.GetMemberByNRA(ctx, tx, claims.NRA)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to get member: %v", err)
	}

	// Ambil semua notifikasi untuk member
	notifications, err := n.NotificationRepo.GetNotificationsByMemberId(ctx, tx, member.IdMember)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to get notifications: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Convert ke DTO
	var result []dto.NotificationResponse
	for _, notification := range notifications {
		response := dto.NotificationResponse{
			IdNotification: notification.IdNotification,
			Type:           notification.Type,
			Title:          notification.Title,
			Message:        notification.Message,
			Pending:        notification.Pending,
			CreatedAt:      notification.CreatedAt,
			FromMember: dto.FromMemberDetails{
				IdMember: notification.FromMemberID,
				Nama:     notification.FromMemberName,
				NRA:      notification.FromMemberNRA,
				Role:     notification.FromMemberRole,
			},
		}

		// Handle nullable fields
		if notification.ReadAt.Valid {
			response.ReadAt = &notification.ReadAt.Time
		}

		if notification.Accepted.Valid {
			response.Accepted = &notification.Accepted.Bool
		}

		// Parse metadata jika ada
		if notification.Metadata.Valid {
			var metadata interface{}
			if err := json.Unmarshal([]byte(notification.Metadata.String), &metadata); err == nil {
				response.Metadata = metadata
			}
		}

		result = append(result, response)
	}

	return result, http.StatusOK, nil
}

func (n *notificationServiceImpl) MarkNotificationAsRead(ctx context.Context, r *http.Request, notificationID string) (int, error) {
	// Ambil token dari header Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return http.StatusUnauthorized, fmt.Errorf("authorization header is required")
	}

	// Format: "Bearer <token>"
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return http.StatusUnauthorized, fmt.Errorf("invalid authorization header format")
	}

	tokenString := tokenParts[1]

	// Validasi token JWT
	claims, err := helper.ValidateJWT(tokenString)
	if err != nil {
		return http.StatusUnauthorized, fmt.Errorf("invalid or expired token: %v", err)
	}

	// Mulai transaksi database
	tx, err := n.DB.Begin()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Ambil data member berdasarkan NRA dari claims untuk mendapatkan ID
	member, err := n.MemberRepo.GetMemberByNRA(ctx, tx, claims.NRA)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to get member: %v", err)
	}

	// Mark notification as read
	err = n.NotificationRepo.MarkNotificationAsRead(ctx, tx, notificationID, member.IdMember)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to mark notification as read: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return http.StatusOK, nil
}

func (n *notificationServiceImpl) GetUnreadNotificationCount(ctx context.Context, r *http.Request) (dto.UnreadCountResponse, int, error) {
	// Ambil token dari header Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return dto.UnreadCountResponse{}, http.StatusUnauthorized, fmt.Errorf("authorization header is required")
	}

	// Format: "Bearer <token>"
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return dto.UnreadCountResponse{}, http.StatusUnauthorized, fmt.Errorf("invalid authorization header format")
	}

	tokenString := tokenParts[1]

	// Validasi token JWT
	claims, err := helper.ValidateJWT(tokenString)
	if err != nil {
		return dto.UnreadCountResponse{}, http.StatusUnauthorized, fmt.Errorf("invalid or expired token: %v", err)
	}

	// Mulai transaksi database
	tx, err := n.DB.Begin()
	if err != nil {
		return dto.UnreadCountResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Ambil data member berdasarkan NRA dari claims untuk mendapatkan ID
	member, err := n.MemberRepo.GetMemberByNRA(ctx, tx, claims.NRA)
	if err != nil {
		return dto.UnreadCountResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to get member: %v", err)
	}

	// Get unread count
	count, err := n.NotificationRepo.GetUnreadNotificationCount(ctx, tx, member.IdMember)
	if err != nil {
		return dto.UnreadCountResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to get unread count: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return dto.UnreadCountResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return dto.UnreadCountResponse{UnreadCount: count}, http.StatusOK, nil
}

func (n *notificationServiceImpl) CreateStatusChangeRequest(ctx context.Context, r *http.Request) (dto.StatusChangeResponse, int, error) {
	// Ambil token dari header Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return dto.StatusChangeResponse{}, http.StatusUnauthorized, fmt.Errorf("authorization header is required")
	}

	// Format: "Bearer <token>"
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return dto.StatusChangeResponse{}, http.StatusUnauthorized, fmt.Errorf("invalid authorization header format")
	}

	tokenString := tokenParts[1]

	// Validasi token JWT
	claims, err := helper.ValidateJWT(tokenString)
	if err != nil {
		return dto.StatusChangeResponse{}, http.StatusUnauthorized, fmt.Errorf("invalid or expired token: %v", err)
	}

	// Parse request body JSON
	var request dto.StatusChangeRequestDTO
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		return dto.StatusChangeResponse{}, http.StatusBadRequest, fmt.Errorf("invalid JSON format: %v", err)
	}

	// Validasi input
	if request.TargetMemberID == "" || request.FromRole == "" || request.ToRole == "" {
		return dto.StatusChangeResponse{}, http.StatusBadRequest, fmt.Errorf("target_member_id, from_role, and to_role are required")
	}

	// Mulai transaksi database
	tx, err := n.DB.Begin()
	if err != nil {
		return dto.StatusChangeResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Ambil data member requester berdasarkan NRA dari claims
	requester, err := n.MemberRepo.GetMemberByNRA(ctx, tx, claims.NRA)
	if err != nil {
		return dto.StatusChangeResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to get requester member: %v", err)
	}

	// Validasi permission - hanya BPH yang bisa request role change untuk sesama BPH
	if requester.Role != "bph" {
		return dto.StatusChangeResponse{}, http.StatusForbidden, fmt.Errorf("only BPH members can request role changes")
	}

	// Validasi target member exists
	targetMember, err := n.MemberRepo.GetMemberById(ctx, tx, request.TargetMemberID)
	if err != nil {
		return dto.StatusChangeResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to get target member: %v", err)
	}

	// Validasi role change rules - BPH only changes other BPH
	if targetMember.Role != "bph" || request.FromRole != "bph" {
		return dto.StatusChangeResponse{}, http.StatusBadRequest, fmt.Errorf("role change requests only allowed for BPH members")
	}

	// Buat notifikasi
	notificationID := uuid.New().String()
	requestID := uuid.New().String()

	// Metadata untuk request
	metadata := map[string]interface{}{
		"request_id": requestID,
		"from_role":  request.FromRole,
		"to_role":    request.ToRole,
	}
	metadataJSON, _ := json.Marshal(metadata)

	notification := model.Notification{
		IdNotification: notificationID,
		TargetMemberID: request.TargetMemberID,
		FromMemberID:   requester.IdMember,
		Type:           "status_change_request",
		Title:          "Status Change Request",
		Message:        fmt.Sprintf("%s requests to change your role from %s to %s", requester.Nama, request.FromRole, request.ToRole),
		Metadata:       sql.NullString{String: string(metadataJSON), Valid: true},
		Pending:        true,
	}

	_, err = n.NotificationRepo.CreateNotification(ctx, tx, notification)
	if err != nil {
		return dto.StatusChangeResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to create notification: %v", err)
	}

	// Buat status change request
	statusChangeReq := model.StatusChangeRequest{
		IdRequest:           requestID,
		NotificationID:      notificationID,
		TargetMemberID:      request.TargetMemberID,
		RequestedByMemberID: requester.IdMember,
		FromRole:            request.FromRole,
		ToRole:              request.ToRole,
		Status:              "pending",
	}

	_, err = n.NotificationRepo.CreateStatusChangeRequest(ctx, tx, statusChangeReq)
	if err != nil {
		return dto.StatusChangeResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to create status change request: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return dto.StatusChangeResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return dto.StatusChangeResponse{
		RequestID:      requestID,
		NotificationID: notificationID,
		Message:        "Role change request sent successfully",
	}, http.StatusOK, nil
}

func (n *notificationServiceImpl) AcceptStatusChangeRequest(ctx context.Context, r *http.Request, requestID string) (dto.StatusChangeAcceptResponse, int, error) {
	// Ambil token dari header Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return dto.StatusChangeAcceptResponse{}, http.StatusUnauthorized, fmt.Errorf("authorization header is required")
	}

	// Format: "Bearer <token>"
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return dto.StatusChangeAcceptResponse{}, http.StatusUnauthorized, fmt.Errorf("invalid authorization header format")
	}

	tokenString := tokenParts[1]

	// Validasi token JWT
	claims, err := helper.ValidateJWT(tokenString)
	if err != nil {
		return dto.StatusChangeAcceptResponse{}, http.StatusUnauthorized, fmt.Errorf("invalid or expired token: %v", err)
	}

	// Mulai transaksi database
	tx, err := n.DB.Begin()
	if err != nil {
		return dto.StatusChangeAcceptResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Ambil data member berdasarkan NRA dari claims
	member, err := n.MemberRepo.GetMemberByNRA(ctx, tx, claims.NRA)
	if err != nil {
		return dto.StatusChangeAcceptResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to get member: %v", err)
	}

	// Ambil status change request
	statusChangeReq, err := n.NotificationRepo.GetStatusChangeRequestById(ctx, tx, requestID)
	if err != nil {
		return dto.StatusChangeAcceptResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to get status change request: %v", err)
	}

	// Validasi bahwa user adalah target dari request
	if statusChangeReq.TargetMemberID != member.IdMember {
		return dto.StatusChangeAcceptResponse{}, http.StatusForbidden, fmt.Errorf("you can only accept your own status change requests")
	}

	// Validasi bahwa request masih pending
	if statusChangeReq.Status != "pending" {
		return dto.StatusChangeAcceptResponse{}, http.StatusBadRequest, fmt.Errorf("request has already been processed")
	}

	// Update member role
	memberToUpdate := model.Member{
		IdMember:          member.IdMember,
		NRA:               member.NRA,
		Nama:              member.Nama,
		AngkatanID:        member.AngkatanID,
		StatusKeanggotaan: member.StatusKeanggotaan,
		Role:              statusChangeReq.ToRole, // Update role
		JurusanID:         member.JurusanID,
		TanggalDikukuhkan: member.TanggalDikukuhkan,
		Email:             member.Email,
		NoHP:              member.NoHP,
		Password:          member.Password,
		Foto:              member.Foto,
	}

	_, err = n.MemberRepo.UpdateMember(ctx, tx, memberToUpdate)
	if err != nil {
		return dto.StatusChangeAcceptResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to update member status: %v", err)
	}

	// Update status change request
	err = n.NotificationRepo.UpdateStatusChangeRequest(ctx, tx, requestID, "accepted", time.Now())
	if err != nil {
		return dto.StatusChangeAcceptResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to update status change request: %v", err)
	}

	// Update notification
	err = n.updateNotificationStatus(ctx, tx, statusChangeReq.NotificationID, false, true)
	if err != nil {
		return dto.StatusChangeAcceptResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to update notification: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return dto.StatusChangeAcceptResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return dto.StatusChangeAcceptResponse{
		Message: "Role change accepted",
		NewRole: statusChangeReq.ToRole,
	}, http.StatusOK, nil
}

func (n *notificationServiceImpl) RejectStatusChangeRequest(ctx context.Context, r *http.Request, requestID string) (dto.StatusChangeRejectResponse, int, error) {
	// Ambil token dari header Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return dto.StatusChangeRejectResponse{}, http.StatusUnauthorized, fmt.Errorf("authorization header is required")
	}

	// Format: "Bearer <token>"
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return dto.StatusChangeRejectResponse{}, http.StatusUnauthorized, fmt.Errorf("invalid authorization header format")
	}

	tokenString := tokenParts[1]

	// Validasi token JWT
	claims, err := helper.ValidateJWT(tokenString)
	if err != nil {
		return dto.StatusChangeRejectResponse{}, http.StatusUnauthorized, fmt.Errorf("invalid or expired token: %v", err)
	}

	// Mulai transaksi database
	tx, err := n.DB.Begin()
	if err != nil {
		return dto.StatusChangeRejectResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Ambil data member berdasarkan NRA dari claims
	member, err := n.MemberRepo.GetMemberByNRA(ctx, tx, claims.NRA)
	if err != nil {
		return dto.StatusChangeRejectResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to get member: %v", err)
	}

	// Ambil status change request
	statusChangeReq, err := n.NotificationRepo.GetStatusChangeRequestById(ctx, tx, requestID)
	if err != nil {
		return dto.StatusChangeRejectResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to get status change request: %v", err)
	}

	// Validasi bahwa user adalah target dari request
	if statusChangeReq.TargetMemberID != member.IdMember {
		return dto.StatusChangeRejectResponse{}, http.StatusForbidden, fmt.Errorf("you can only reject your own status change requests")
	}

	// Validasi bahwa request masih pending
	if statusChangeReq.Status != "pending" {
		return dto.StatusChangeRejectResponse{}, http.StatusBadRequest, fmt.Errorf("request has already been processed")
	}

	// Update status change request
	err = n.NotificationRepo.UpdateStatusChangeRequest(ctx, tx, requestID, "rejected", time.Now())
	if err != nil {
		return dto.StatusChangeRejectResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to update status change request: %v", err)
	}

	// Update notification
	err = n.updateNotificationStatus(ctx, tx, statusChangeReq.NotificationID, false, false)
	if err != nil {
		return dto.StatusChangeRejectResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to update notification: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return dto.StatusChangeRejectResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return dto.StatusChangeRejectResponse{
		Message: "Role change rejected",
	}, http.StatusOK, nil
}

func (n *notificationServiceImpl) SendStatusChangeNotification(ctx context.Context, fromMemberID, targetMemberID, fromRole, toRole string) error {
	tx, err := n.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Ambil data requester member
	requester, err := n.MemberRepo.GetMemberById(ctx, tx, fromMemberID)
	if err != nil {
		return fmt.Errorf("failed to get requester member: %v", err)
	}

	// Buat notifikasi
	notificationID := uuid.New().String()
	requestID := uuid.New().String()

	// Metadata untuk request
	metadata := map[string]interface{}{
		"request_id": requestID,
		"from_role":  fromRole,
		"to_role":    toRole,
	}
	metadataJSON, _ := json.Marshal(metadata)

	notification := model.Notification{
		IdNotification: notificationID,
		TargetMemberID: targetMemberID,
		FromMemberID:   fromMemberID,
		Type:           "status_change_request",
		Title:          "Status Change Request",
		Message:        fmt.Sprintf("%s requests to change your role from %s to %s", requester.Nama, fromRole, toRole),
		Metadata:       sql.NullString{String: string(metadataJSON), Valid: true},
		Pending:        true,
	}

	_, err = n.NotificationRepo.CreateNotification(ctx, tx, notification)
	if err != nil {
		return fmt.Errorf("failed to create notification: %v", err)
	}

	// Buat status change request
	statusChangeReq := model.StatusChangeRequest{
		IdRequest:           requestID,
		NotificationID:      notificationID,
		TargetMemberID:      targetMemberID,
		RequestedByMemberID: fromMemberID,
		FromRole:            fromRole,
		ToRole:              toRole,
		Status:              "pending",
	}

	_, err = n.NotificationRepo.CreateStatusChangeRequest(ctx, tx, statusChangeReq)
	if err != nil {
		return fmt.Errorf("failed to create status change request: %v", err)
	}

	return tx.Commit()
}

// Helper function to update notification status
func (n *notificationServiceImpl) updateNotificationStatus(ctx context.Context, tx *sql.Tx, notificationID string, pending bool, accepted bool) error {
	query := `
		UPDATE notifications 
		SET pending = ?, accepted = ?, read_at = NOW(), updated_at = NOW()
		WHERE id_notification = ?
	`

	_, err := tx.ExecContext(ctx, query, pending, accepted, notificationID)
	if err != nil {
		return fmt.Errorf("failed to update notification status: %v", err)
	}

	return nil
}
