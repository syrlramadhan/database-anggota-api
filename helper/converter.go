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
		Angkatan:          member.Angkatan.NamaAngkatan,
		StatusKeanggotaan: member.StatusKeanggotaan,
		Jurusan:           member.Jurusan.NamaJurusan,
		TanggalDikukuhkan: member.TanggalDikukuhkan,
		Email:             member.Email.String,
		NoHP:              member.NoHP.String,
		Password:          member.Password.String,
		Foto:              member.Foto.String,
	}
}

func ConvertMemberToProfileResponseDTO(member model.Member) dto.ProfileResponse {
	return dto.ProfileResponse{
		IdMember:          member.IdMember,
		NRA:               member.NRA.String,
		Nama:              member.Nama,
		Angkatan:          member.Angkatan.NamaAngkatan,
		StatusKeanggotaan: member.StatusKeanggotaan,
		Jurusan:           member.Jurusan.NamaJurusan,
		TanggalDikukuhkan: member.TanggalDikukuhkan,
		Email:             member.Email.String,
		NoHP:              member.NoHP.String,
		Foto:              member.Foto.String,
	}
}

func ConvertMemberToCreateResponseDTO(member model.Member) dto.MemberCreateResponse {
	return dto.MemberCreateResponse{
		IdMember:          member.IdMember,
		NRA:               member.NRA.String,
		Nama:              member.Nama,
		Angkatan:          member.Angkatan.NamaAngkatan,
		StatusKeanggotaan: member.StatusKeanggotaan,
		Jurusan:           member.Jurusan.NamaJurusan,
		TanggalDikukuhkan: member.TanggalDikukuhkan,
		Foto:              member.Foto.String,
		LoginToken:        member.LoginToken.String,
	}
}

func ConvertMemberToListResDTO(members []model.Member) []dto.MemberResponse {
	var memberResponse []dto.MemberResponse

	for _, member := range members {
		memberResponse = append(memberResponse, ConvertMemberToResponseDTO(member))
	}

	return memberResponse
}
