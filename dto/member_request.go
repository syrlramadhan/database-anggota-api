package dto

import (
	"github.com/syrlramadhan/database-anggota-api/util"
)

type MemberRequest struct {
	NRA               string           `json:"nra"`
	Nama              string           `json:"nama"`
	Angkatan          string           `json:"angkatan"`
	StatusKeanggotaan string           `json:"status_keanggotaan"`
	Jurusan           string           `json:"jurusan"`
	TanggalDikukuhkan *util.CustomDate `json:"tanggal_dikukuhkan,omitempty"`
	Foto              string           `json:"foto"`
}

type LoginRequest struct {
	NRA      string `json:"nra"`
	Password string `json:"password"`
}

type LoginTokenRequest struct {
	Token string `json:"token"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type SetPasswordRequest struct {
	Password string `json:"password"`
}

type CompleteProfileRequest struct {
	Email             string           `json:"email"`
	NoHP              string           `json:"nomor_hp"`
	TanggalDikukuhkan *util.CustomDate `json:"tanggal_dikukuhkan"`
	NamaLengkap       string           `json:"nama"`
	Foto              string           `json:"-"` // Not from JSON, handled as file upload
}
