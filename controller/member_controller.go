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
	Login(w http.ResponseWriter, r *http.Request, ps httprouter.Params)
	LoginToken(w http.ResponseWriter, r *http.Request, ps httprouter.Params)
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
	loginRequest := dto.LoginTokenRequest{}
	util.ReadFromRequestBody(r, &loginRequest)

	responseDTO, code, err := m.MemberService.LoginToken(r.Context(), loginRequest)
	if err != nil {
		helper.WriteJSONError(w, code, err.Error())
		return
	}
	helper.WriteJSONSuccess(w, responseDTO, "login successfully")
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