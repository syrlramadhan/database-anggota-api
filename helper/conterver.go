package helper

import (
	"github.com/syrlramadhan/database-anggota-api/dto"
	"github.com/syrlramadhan/database-anggota-api/model"
)

func ConvertMemberToResponseDTO(member model.Member) dto.MemberResponse {
	return dto.MemberResponse{
		IdMember:          member.IdMember,
		NRA:               member.NRA.String,
		Nama:              member.Nama,
		Angkatan:          member.AngkatanID,
		StatusKeanggotaan: member.StatusKeanggotaan,
		Jurusan:           member.JurusanID.String,
		TanggalDikukuhkan: member.TanggalDikukuhkan,
		Email:             member.Email.String,
		NoHP:              member.NoHP.String,
		Password:          member.Password.String,
		Foto:              member.Foto.String,
	}
}
