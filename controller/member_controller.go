package controller

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/syrlramadhan/database-anggota-api/dto"
	"github.com/syrlramadhan/database-anggota-api/helper"
	"github.com/syrlramadhan/database-anggota-api/service"
	"github.com/syrlramadhan/database-anggota-api/util"
)

type MemberController interface {
	AddMember(w http.ResponseWriter, r *http.Request, ps httprouter.Params)
	UpdateMember(w http.ResponseWriter, r *http.Request, ps httprouter.Params)
	UpdateMemberWithNotification(w http.ResponseWriter, r *http.Request, ps httprouter.Params)
	GetAllMember(w http.ResponseWriter, r *http.Request, ps httprouter.Params)
	GetMemberById(w http.ResponseWriter, r *http.Request, ps httprouter.Params)
	DeleteMember(w http.ResponseWriter, r *http.Request, ps httprouter.Params)
	Login(w http.ResponseWriter, r *http.Request, ps httprouter.Params)
	LoginToken(w http.ResponseWriter, r *http.Request, ps httprouter.Params)
	GetProfile(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	SetPassword(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	CompleteProfile(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
}

type memberControllerImpl struct {
	MemberService service.MemberService
}

func NewMemberController(memberService service.MemberService) MemberController {
	return &memberControllerImpl{
		MemberService: memberService,
	}
}

// AddMember implements MemberController.
func (m *memberControllerImpl) AddMember(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	responseDTO, code, err := m.MemberService.AddMember(r.Context(), r)
	if err != nil {
		helper.WriteJSONError(w, code, err.Error())
		return
	}

	helper.WriteJSONSuccess(w, responseDTO, "registration successfully")
}

// UpdateMember implements MemberController.
func (m *memberControllerImpl) UpdateMember(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")

	responseDTO, code, err := m.MemberService.UpdateMember(r.Context(), r, id)
	if err != nil {
		helper.WriteJSONError(w, code, err.Error())
		return
	}

	helper.WriteJSONSuccess(w, responseDTO, "updated successfully")
}

// UpdateMemberWithNotification implements MemberController with DPO notification capability.
func (m *memberControllerImpl) UpdateMemberWithNotification(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")

	// Get fromMemberId from JWT token or header
	fromMemberId := r.Header.Get("X-Member-ID")
	if fromMemberId == "" {
		helper.WriteJSONError(w, http.StatusUnauthorized, "Member ID required in header")
		return
	}

	responseDTO, code, err := m.MemberService.UpdateMemberWithNotification(r.Context(), r, id, fromMemberId)
	if err != nil {
		helper.WriteJSONError(w, code, err.Error())
		return
	}

	helper.WriteJSONSuccess(w, responseDTO, "updated successfully")
}

// GetAllMember implements MemberController.
func (m *memberControllerImpl) GetAllMember(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	responseDTO, code, err := m.MemberService.GetAllMember(r.Context())
	if err != nil {
		helper.WriteJSONError(w, code, err.Error())
		return
	}

	helper.WriteJSONSuccess(w, responseDTO, "get all member successfully")
}

// GetMemberById implements MemberController.
func (m *memberControllerImpl) GetMemberById(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")

	responseDTO, code, err := m.MemberService.GetMemberById(r.Context(), id)
	if err != nil {
		helper.WriteJSONError(w, code, err.Error())
		return
	}

	helper.WriteJSONSuccess(w, responseDTO, "get member successfully")
}

// DeleteMember implements MemberController.
func (m *memberControllerImpl) DeleteMember(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")

	code, err := m.MemberService.DeleteMember(r.Context(), id)
	if err != nil {
		helper.WriteJSONError(w, code, err.Error())
		return
	}

	helper.WriteJSONSuccess(w, "Member deleted successfully", "delete member successfully")
}

// Login implements MemberController.
func (m *memberControllerImpl) Login(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	loginRequest := dto.LoginRequest{}
	util.ReadFromRequestBody(r, &loginRequest)

	responseDTO, code, err := m.MemberService.Login(r.Context(), loginRequest)
	if err != nil {
		helper.WriteJSONError(w, code, err.Error())
		return
	}
	helper.WriteJSONSuccess(w, responseDTO, "login successfully")
}

// LoginToken implements MemberController.
func (m *memberControllerImpl) LoginToken(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	responseDTO, code, err := m.MemberService.LoginToken(r.Context(), r)
	if err != nil {
		helper.WriteJSONError(w, code, err.Error())
		return
	}
	helper.WriteJSONSuccess(w, responseDTO, "login successfully")
}

// GetProfile implements MemberController.
func (m *memberControllerImpl) GetProfile(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	responseDTO, code, err := m.MemberService.GetProfile(r.Context(), r)
	if err != nil {
		helper.WriteJSONError(w, code, err.Error())
		return
	}
	helper.WriteJSONSuccess(w, responseDTO, "get profile successfully")
}

// SetPassword implements MemberController.
func (m *memberControllerImpl) SetPassword(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	code, err := m.MemberService.SetPassword(r.Context(), r)
	if err != nil {
		helper.WriteJSONError(w, code, err.Error())
		return
	}
	helper.WriteJSONSuccess(w, "Password updated successfully", "set password successfully")
}

// CompleteProfile implements MemberController.
func (m *memberControllerImpl) CompleteProfile(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	responseDTO, code, err := m.MemberService.CompleteProfile(r.Context(), r)
	if err != nil {
		helper.WriteJSONError(w, code, err.Error())
		return
	}
	helper.WriteJSONSuccess(w, responseDTO, "complete profile successfully")
}
