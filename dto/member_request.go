package dto

import (
	"github.com/syrlramadhan/database-anggota-api/util"
)

type MemberRequest struct {
	NRA               string          `json:"nra"`
	Nama              string          `json:"nama"`
	Angkatan          string          `json:"angkatan"`
	StatusKeanggotaan string          `json:"status_keanggotaan"`
	Jurusan           string          `json:"jurusan"`
	TanggalDikukuhkan *util.CustomDate `json:"tanggal_dikukuhkan"`
	Email             string          `json:"email"`
	NoHP              string          `json:"nomor_hp"`
	Password          string          `json:"password"`
	Foto              string          `json:"foto"`
}

type LoginRequest struct {
	NRA      string `json:"nra"`
	Password string `json:"password"`
}
