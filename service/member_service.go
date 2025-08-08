package service

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/syrlramadhan/database-anggota-api/dto"
	"github.com/syrlramadhan/database-anggota-api/helper"
	"github.com/syrlramadhan/database-anggota-api/model"
	"github.com/syrlramadhan/database-anggota-api/repository"
	"github.com/syrlramadhan/database-anggota-api/util"
)

type MemberService interface {
	AddMember(ctx context.Context, r *http.Request) (dto.MemberResponse, int, error)
	UpdateMember(ctx context.Context, r *http.Request, id string) (dto.MemberResponse, int, error)
	Login(ctx context.Context, loginRequest dto.LoginRequest) (string, int, error)
	LoginToken(ctx context.Context, loginRequest dto.LoginTokenRequest) (string, int, error)
}

type memberServiceImpl struct {
	MemberRepo repository.MemberRepository
	DB         *sql.DB
}

func NewMemberService(memberRepo repository.MemberRepository, db *sql.DB) MemberService {
	return &memberServiceImpl{
		MemberRepo: memberRepo,
		DB:         db,
	}
}

// AddMember implements MemberService.
func (m *memberServiceImpl) AddMember(ctx context.Context, r *http.Request) (dto.MemberResponse, int, error) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to parse form: %v", err)
	}

	tanggalStr := r.FormValue("tanggal_dikukuhkan")
	parsedTanggal, err := time.Parse("02-01-2006", tanggalStr)
	if err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to parse date: %v", err)
	}

	memberRequest := dto.MemberRequest{
		NRA:               r.FormValue("nra"),
		Nama:              r.FormValue("nama"),
		Angkatan:          r.FormValue("angkatan"),
		StatusKeanggotaan: r.FormValue("status_keanggotaan"),
		Jurusan:           r.FormValue("jurusan"),
		TanggalDikukuhkan: &util.CustomDate{Time: parsedTanggal},
		Email:             r.FormValue("email"),
		NoHP:              r.FormValue("nomor_hp"),
		Password:          r.FormValue("password"),
	}

	if memberRequest.NRA == "" {
		return dto.MemberResponse{}, http.StatusBadRequest, fmt.Errorf("field nra cannot be empty")
	} else if memberRequest.Nama == "" {
		return dto.MemberResponse{}, http.StatusBadRequest, fmt.Errorf("field name cannot be empty")
	} else if memberRequest.Angkatan == "" {
		return dto.MemberResponse{}, http.StatusBadRequest, fmt.Errorf("field generation cannot be empty")
	} else if memberRequest.StatusKeanggotaan == "" {
		return dto.MemberResponse{}, http.StatusBadRequest, fmt.Errorf("field membership status cannot be empty")
	} else if memberRequest.Jurusan == "" {
		return dto.MemberResponse{}, http.StatusBadRequest, fmt.Errorf("the major field cannot be empty")
	} else if memberRequest.TanggalDikukuhkan.String() == "" {
		return dto.MemberResponse{}, http.StatusBadRequest, fmt.Errorf("field date column cannot be empty")
	} else if memberRequest.Email != "" && !util.IsValidEmail(memberRequest.Email) {
		return dto.MemberResponse{}, http.StatusBadRequest, fmt.Errorf("invalid email format")
	} else if memberRequest.NRA != "" && !util.IsValidNRA(memberRequest.NRA) {
		return dto.MemberResponse{}, http.StatusBadRequest, fmt.Errorf("invalid nra format")
	} else if memberRequest.Password != "" && len(memberRequest.Password) < 6 {
		return dto.MemberResponse{}, http.StatusBadRequest, fmt.Errorf("passwords must be at least 6 digits long")
	}

	file, header, err := r.FormFile("foto")
	if file == nil {
		return dto.MemberResponse{}, http.StatusBadRequest, fmt.Errorf("the file cannot be empty")
	} else if err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to add file")
	}
	defer file.Close()

	fileName := fmt.Sprintf("%s_%s%s", memberRequest.NRA, memberRequest.Nama, filepath.Ext(header.Filename))

	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, os.ModePerm)
	}

	filePath := filepath.Join(uploadDir, fileName)
	out, err := os.Create(filePath)
	if err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to copy file: %v", err)
	}

	memberRequest.Foto = fileName

	tx, err := m.DB.Begin()
	if err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Commit()

	var jurusan model.Jurusan
	var angkatan model.Angkatan

	get_jurusan, err := m.MemberRepo.GetJurusanByName(ctx, tx, jurusan, memberRequest.Jurusan)
	if err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to get major: %v", err)
	}

	hassedPass, err := helper.HashPassword(memberRequest.Password)
	if err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to hash password: %v", err)
	}

	member := model.Member{
		IdMember:          uuid.New().String(),
		NRA:               sql.NullString{String: memberRequest.NRA, Valid: true},
		Nama:              memberRequest.Nama,
		AngkatanID:        memberRequest.Angkatan,
		StatusKeanggotaan: memberRequest.StatusKeanggotaan,
		JurusanID:         sql.NullString{String: get_jurusan.IdJurusan, Valid: true},
		TanggalDikukuhkan: memberRequest.TanggalDikukuhkan,
		Email:             sql.NullString{String: memberRequest.Email, Valid: true},
		NoHP:              sql.NullString{String: memberRequest.NoHP, Valid: true},
		Password:          sql.NullString{String: hassedPass, Valid: true},
		Foto:              sql.NullString{String: memberRequest.Foto, Valid: true},
	}

	addMember, err := m.MemberRepo.AddMember(ctx, tx, member)
	if err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to add member: %v", err)
	}

	get_angkatan, err := m.MemberRepo.GetAngkatanById(ctx, tx, angkatan, memberRequest.Angkatan)
	if err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to get the draft: %v", err)
	}

	addMember.AngkatanID = get_angkatan.NamaAngkatan
	addMember.JurusanID = sql.NullString{String: memberRequest.Jurusan, Valid: true}

	return helper.ConvertMemberToResponseDTO(addMember), http.StatusOK, nil
}

// UpdateMember implements MemberService.
func (m *memberServiceImpl) UpdateMember(ctx context.Context, r *http.Request, id string) (dto.MemberResponse, int, error) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to parse form: %v", err)
	}

	// Ambil data dari form
	tanggalStr := r.FormValue("tanggal_dikukuhkan")
	memberRequest := dto.MemberRequest{
		NRA:               r.FormValue("nra"),
		Nama:              r.FormValue("nama"),
		Angkatan:          r.FormValue("angkatan"),
		StatusKeanggotaan: r.FormValue("status_keanggotaan"),
		Jurusan:           r.FormValue("jurusan"),
		Email:             r.FormValue("email"),
		NoHP:              r.FormValue("nomor_hp"),
		Password:          r.FormValue("password"),
	}

	if tanggalStr != "" {
		parsedTanggal, err := time.Parse("02-01-2006", tanggalStr)
		if err != nil {
			return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to parse date: %v", err)
		}
		memberRequest.TanggalDikukuhkan = &util.CustomDate{Time: parsedTanggal}
	}

	// Ambil data member dari DB dulu (untuk fallback)
	tx, err := m.DB.Begin()
	if err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	getMember, err := m.MemberRepo.GetMemberById(ctx, tx, id)
	if err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to get member: %v", err)
	}

	var fn_nra, fn_nama string

	if memberRequest.NRA == "" {
		fn_nra = getMember.NRA.String
	} else {
		fn_nra = memberRequest.NRA
	}

	if memberRequest.Nama == "" {
		fn_nama = getMember.Nama
	} else {
		fn_nama = memberRequest.Nama
	}

	// Handle optional foto
	file, header, err := r.FormFile("foto")

	if err == nil && header != nil {
		defer file.Close()

		// Buat nama file baru berdasarkan NRA dan Nama baru
		fileName := fmt.Sprintf("%s_%s%s", fn_nra, fn_nama, filepath.Ext(header.Filename))

		// Buat folder ./uploads jika belum ada
		uploadDir := "./uploads"
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			if err := os.Mkdir(uploadDir, os.ModePerm); err != nil {
				return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to create upload dir: %v", err)
			}
		}

		// Simpan file baru
		filePath := filepath.Join(uploadDir, fileName)
		out, err := os.Create(filePath)
		if err != nil {
			return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to create file: %v", err)
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to copy file: %v", err)
		}

		// Simpan nama file ke struct request
		memberRequest.Foto = fileName

		// (Opsional) Hapus file lama jika perlu:
		if getMember.Foto.Valid {
			oldFilePath := filepath.Join("./uploads", getMember.Foto.String)
			if getMember.Foto.String != fileName {
				_ = os.Remove(oldFilePath) // Tidak wajib, tergantung kebutuhan
			}
		}
	} else if err == http.ErrMissingFile && getMember.Foto.Valid {
		// Ambil ekstensi dari file lama
		oldExt := filepath.Ext(getMember.Foto.String)
		newFileName := fmt.Sprintf("%s_%s%s", fn_nra, fn_nama, oldExt)
		oldFilePath := filepath.Join("./uploads", getMember.Foto.String)
		newFilePath := filepath.Join("./uploads", newFileName)

		if getMember.Foto.String != newFileName {
			// Rename file jika nama file berubah
			if err := os.Rename(oldFilePath, newFilePath); err != nil {
				return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to rename file: %v", err)
			}
		}

		// Tetapkan nama file baru ke request
		memberRequest.Foto = newFileName
	} else {
		memberRequest.Foto = ""
	}

	// Validasi
	if memberRequest.Email != "" && !util.IsValidEmail(memberRequest.Email) {
		return dto.MemberResponse{}, http.StatusBadRequest, fmt.Errorf("invalid email format")
	}
	if memberRequest.NRA != "" && !util.IsValidNRA(memberRequest.NRA) {
		return dto.MemberResponse{}, http.StatusBadRequest, fmt.Errorf("invalid NRA format")
	}
	if memberRequest.Password != "" && len(memberRequest.Password) < 6 {
		return dto.MemberResponse{}, http.StatusBadRequest, fmt.Errorf("password must be at least 6 characters")
	}

	// Jurusan
	var getJurusan model.Jurusan
	if memberRequest.Jurusan != "" {
		getJurusan, err = m.MemberRepo.GetJurusanByName(ctx, tx, model.Jurusan{}, memberRequest.Jurusan)
		if err != nil {
			return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to get jurusan: %v", err)
		}
	}

	// Angkatan
	var getAngkatan model.Angkatan
	if memberRequest.Angkatan != "" {
		getAngkatan, err = m.MemberRepo.GetAngkatanById(ctx, tx, model.Angkatan{}, memberRequest.Angkatan)
		if err != nil {
			return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to get angkatan: %v", err)
		}
	}

	// Hash password jika ada yang baru
	hassedPass := getMember.Password.String
	if memberRequest.Password != "" {
		hassedPass, err = helper.HashPassword(memberRequest.Password)
		if err != nil {
			return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to hash password: %v", err)
		}
	}

	// Bangun object final member
	member := model.Member{
		IdMember: getMember.IdMember,
		NRA:      helper.ChooseNullString(memberRequest.NRA, getMember.NRA),
		Nama:     helper.ChooseString(memberRequest.Nama, getMember.Nama),
		AngkatanID: func() string {
			if memberRequest.Angkatan != "" {
				return memberRequest.Angkatan
			}
			return getMember.AngkatanID
		}(),
		StatusKeanggotaan: helper.ChooseString(memberRequest.StatusKeanggotaan, getMember.StatusKeanggotaan),
		JurusanID: func() sql.NullString {
			if memberRequest.Jurusan != "" {
				return sql.NullString{String: getJurusan.IdJurusan, Valid: true}
			}
			return getMember.JurusanID
		}(),
		TanggalDikukuhkan: func() *util.CustomDate {
			if memberRequest.TanggalDikukuhkan != nil {
				return memberRequest.TanggalDikukuhkan
			}
			return getMember.TanggalDikukuhkan
		}(),
		Email:    helper.ChooseNullString(memberRequest.Email, getMember.Email),
		NoHP:     helper.ChooseNullString(memberRequest.NoHP, getMember.NoHP),
		Password: sql.NullString{String: hassedPass, Valid: true},
		Foto: func() sql.NullString {
			if memberRequest.Foto != "" {
				return sql.NullString{String: memberRequest.Foto, Valid: true}
			}
			return getMember.Foto
		}(),
	}

	updatedMember, err := m.MemberRepo.UpdateMember(ctx, tx, member)
	if err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to update member: %v", err)
	}

	// Inject nama angkatan dan jurusan ke response
	updatedMember.AngkatanID = func() string {
		if memberRequest.Angkatan != "" {
			return getAngkatan.NamaAngkatan
		}
		return getMember.AngkatanID
	}()
	updatedMember.JurusanID = func() sql.NullString {
		if memberRequest.Jurusan != "" {
			return sql.NullString{String: getJurusan.NamaJurusan}
		}
		return sql.NullString{String: getMember.JurusanID.String}
	}()

	if err := tx.Commit(); err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return helper.ConvertMemberToResponseDTO(updatedMember), http.StatusOK, nil
}

// Login implements MemberService.
func (m *memberServiceImpl) Login(ctx context.Context, loginRequest dto.LoginRequest) (string, int, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Commit()

	if loginRequest.NRA == "" {
		return "", http.StatusBadRequest, fmt.Errorf("field nra cannot be empty")
	} else if loginRequest.Password == "" {
		return "", http.StatusBadRequest, fmt.Errorf("field password cannot be empty")
	}

	member, err := m.MemberRepo.GetMemberByNRA(ctx, tx, loginRequest.NRA)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("failed to get member: %v", err)
	}

	if !helper.VerifyPassword(member.Password.String, loginRequest.Password) {
		return "", http.StatusBadRequest, fmt.Errorf("invalid nra or password: %v", err)
	}

	token, err := helper.GenerateJWT(loginRequest.NRA, member.Nama)
	if err != nil {
		return "", http.StatusBadRequest, fmt.Errorf("failed to generate token: %v", err)
	}

	return token, http.StatusOK, nil
}

// LoginToken implements MemberService.
func (m *memberServiceImpl) LoginToken(ctx context.Context, loginRequest dto.LoginTokenRequest) (string, int, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Commit()

	if loginRequest.Token == "" {
		return "", http.StatusBadRequest, fmt.Errorf("field token cannot be empty")
	}

	member, err := m.MemberRepo.GetMemberByToken(ctx, tx, loginRequest.Token)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("failed to get member: %v", err)
	}

	token, err := helper.GenerateJWT(member.NRA.String, member.Nama)
	if err != nil {
		return "", http.StatusBadRequest, fmt.Errorf("failed to generate token: %v", err)
	}

	return token, http.StatusOK, nil
}
