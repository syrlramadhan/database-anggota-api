package controller

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/syrlramadhan/database-anggota-api/helper"
	"github.com/syrlramadhan/database-anggota-api/service"
)

type NotificationController interface {
	GetNotifications(w http.ResponseWriter, r *http.Request, ps httprouter.Params)
	MarkNotificationAsRead(w http.ResponseWriter, r *http.Request, ps httprouter.Params)
	GetUnreadNotificationCount(w http.ResponseWriter, r *http.Request, ps httprouter.Params)

	CreateStatusChangeRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params)
	AcceptStatusChangeRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params)
	RejectStatusChangeRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params)
}

type notificationControllerImpl struct {
	NotificationService service.NotificationService
}

func NewNotificationController(notificationService service.NotificationService) NotificationController {
	return &notificationControllerImpl{
		NotificationService: notificationService,
	}
}

func (n *notificationControllerImpl) GetNotifications(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := context.Background()

	notifications, statusCode, err := n.NotificationService.GetNotifications(ctx, r)
	if err != nil {
		helper.WriteJSONError(w, statusCode, err.Error())
		return
	}

	helper.WriteJSONSuccess(w, notifications, "Notifications retrieved successfully")
}

func (n *notificationControllerImpl) MarkNotificationAsRead(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := context.Background()
	notificationID := ps.ByName("id")

	statusCode, err := n.NotificationService.MarkNotificationAsRead(ctx, r, notificationID)
	if err != nil {
		helper.WriteJSONError(w, statusCode, err.Error())
		return
	}

	helper.WriteJSONSuccess(w, map[string]string{
		"message": "Notification marked as read",
	}, "Notification marked as read")
}

func (n *notificationControllerImpl) GetUnreadNotificationCount(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := context.Background()

	count, statusCode, err := n.NotificationService.GetUnreadNotificationCount(ctx, r)
	if err != nil {
		helper.WriteJSONError(w, statusCode, err.Error())
		return
	}

	helper.WriteJSONSuccess(w, count, "Unread notification count retrieved successfully")
}

func (n *notificationControllerImpl) CreateStatusChangeRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := context.Background()

	response, statusCode, err := n.NotificationService.CreateStatusChangeRequest(ctx, r)
	if err != nil {
		helper.WriteJSONError(w, statusCode, err.Error())
		return
	}

	helper.WriteJSONSuccess(w, response, "Status change request created successfully")
}

func (n *notificationControllerImpl) AcceptStatusChangeRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := context.Background()
	requestID := ps.ByName("id")

	response, statusCode, err := n.NotificationService.AcceptStatusChangeRequest(ctx, r, requestID)
	if err != nil {
		helper.WriteJSONError(w, statusCode, err.Error())
		return
	}

	helper.WriteJSONSuccess(w, response, "Status change request accepted successfully")
}

func (n *notificationControllerImpl) RejectStatusChangeRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := context.Background()
	requestID := ps.ByName("id")

	response, statusCode, err := n.NotificationService.RejectStatusChangeRequest(ctx, r, requestID)
	if err != nil {
		helper.WriteJSONError(w, statusCode, err.Error())
		return
	}

	helper.WriteJSONSuccess(w, response, "Status change request rejected successfully")
}