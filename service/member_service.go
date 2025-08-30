package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/syrlramadhan/database-anggota-api/dto"
	"github.com/syrlramadhan/database-anggota-api/helper"
	"github.com/syrlramadhan/database-anggota-api/model"
	"github.com/syrlramadhan/database-anggota-api/repository"
	"github.com/syrlramadhan/database-anggota-api/util"
)

type MemberService interface {
	AddMember(ctx context.Context, r *http.Request) (dto.MemberCreateResponse, int, error)
	GetAllMember(ctx context.Context) ([]dto.MemberResponse, int, error)
	GetMemberById(ctx context.Context, memberId string) (dto.MemberResponse, int, error)
	UpdateMember(ctx context.Context, r *http.Request, memberId string) (dto.MemberResponse, int, error)
	DeleteMember(ctx context.Context, memberId string) (int, error)
	Login(ctx context.Context, loginRequest dto.LoginRequest) (string, int, error)
	LoginToken(ctx context.Context, r *http.Request) (string, int, error)
	GetProfile(ctx context.Context, r *http.Request) (dto.ProfileResponse, int, error)
	SetPassword(ctx context.Context, r *http.Request) (int, error)
	CompleteProfile(ctx context.Context, r *http.Request) (dto.MemberResponse, int, error)
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
func (m *memberServiceImpl) AddMember(ctx context.Context, r *http.Request) (dto.MemberCreateResponse, int, error) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		return dto.MemberCreateResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to parse form: %v", err)
	}

	memberRequest := dto.MemberRequest{
		NRA:               r.FormValue("nra"),
		Nama:              r.FormValue("nama"),
		Angkatan:          r.FormValue("angkatan"),
		StatusKeanggotaan: r.FormValue("status_keanggotaan"),
		Jurusan:           r.FormValue("jurusan"),
	}

	// Parse tanggal_dikukuhkan if provided (optional)
	tanggalStr := r.FormValue("tanggal_dikukuhkan")
	if tanggalStr != "" {
		parsedTanggal, err := time.Parse("02-01-2006", tanggalStr)
		if err != nil {
			return dto.MemberCreateResponse{}, http.StatusBadRequest, fmt.Errorf("invalid date format, use DD-MM-YYYY: %v", err)
		}
		memberRequest.TanggalDikukuhkan = &util.CustomDate{Time: parsedTanggal}
	}

	// Validasi field required
	if memberRequest.NRA == "" {
		return dto.MemberCreateResponse{}, http.StatusBadRequest, fmt.Errorf("field nra cannot be empty")
	} else if memberRequest.Nama == "" {
		return dto.MemberCreateResponse{}, http.StatusBadRequest, fmt.Errorf("field name cannot be empty")
	} else if memberRequest.Angkatan == "" {
		return dto.MemberCreateResponse{}, http.StatusBadRequest, fmt.Errorf("field generation cannot be empty")
	} else if memberRequest.StatusKeanggotaan == "" {
		return dto.MemberCreateResponse{}, http.StatusBadRequest, fmt.Errorf("field membership status cannot be empty")
	} else if memberRequest.Jurusan == "" {
		return dto.MemberCreateResponse{}, http.StatusBadRequest, fmt.Errorf("the major field cannot be empty")
	} else if memberRequest.NRA != "" && !util.IsValidNRA(memberRequest.NRA) {
		return dto.MemberCreateResponse{}, http.StatusBadRequest, fmt.Errorf("invalid nra format")
	}

	file, header, err := r.FormFile("foto")
	if file == nil {
		return dto.MemberCreateResponse{}, http.StatusBadRequest, fmt.Errorf("the file cannot be empty")
	} else if err != nil {
		return dto.MemberCreateResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to add file")
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
		return dto.MemberCreateResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		return dto.MemberCreateResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to copy file: %v", err)
	}

	memberRequest.Foto = fileName

	tx, err := m.DB.Begin()
	if err != nil {
		return dto.MemberCreateResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Commit()

	var jurusan model.Jurusan
	var angkatan model.Angkatan

	get_jurusan, err := m.MemberRepo.GetJurusanByName(ctx, tx, jurusan, memberRequest.Jurusan)
	if err != nil {
		return dto.MemberCreateResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to get major: %v", err)
	}

	// Generate random token for first-time login
	randomToken, err := helper.GenerateRandomToken(32) // 64 character hex string
	if err != nil {
		return dto.MemberCreateResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to generate token: %v", err)
	}

	// Cek apakah angkatan sudah ada di database SEBELUM membuat member
	get_angkatan, err := m.MemberRepo.GetAngkatanById(ctx, tx, angkatan, memberRequest.Angkatan)
	if err != nil {
		// Jika angkatan tidak ada, buat angkatan baru
		if err == sql.ErrNoRows {
			// Format nama angkatan berdasarkan ID (misal: "015" -> "Angkatan 015")
			newAngkatan := model.Angkatan{
				IdAngkatan:   memberRequest.Angkatan,
				NamaAngkatan: fmt.Sprintf("Angkatan %s", memberRequest.Angkatan),
			}
			
			// Tambahkan angkatan baru ke database
			get_angkatan, err = m.MemberRepo.AddAngkatan(ctx, tx, newAngkatan)
			if err != nil {
				return dto.MemberCreateResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to create new angkatan: %v", err)
			}
		} else {
			return dto.MemberCreateResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to get angkatan: %v", err)
		}
	}

	member := model.Member{
		IdMember:          uuid.New().String(),
		NRA:               sql.NullString{String: memberRequest.NRA, Valid: true},
		Nama:              memberRequest.Nama,
		AngkatanID:        memberRequest.Angkatan, // Gunakan ID angkatan yang sudah divalidasi
		StatusKeanggotaan: memberRequest.StatusKeanggotaan,
		JurusanID:         sql.NullString{String: get_jurusan.IdJurusan, Valid: true},
		TanggalDikukuhkan: memberRequest.TanggalDikukuhkan,
		Foto:              sql.NullString{String: memberRequest.Foto, Valid: true},
		LoginToken:        sql.NullString{String: randomToken, Valid: true},
	}

	addMember, err := m.MemberRepo.AddMember(ctx, tx, member)
	if err != nil {
		return dto.MemberCreateResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to add member: %v", err)
	}

	addMember.AngkatanID = get_angkatan.NamaAngkatan
	addMember.JurusanID = sql.NullString{String: memberRequest.Jurusan, Valid: true}

	return helper.ConvertMemberToCreateResponseDTO(addMember), http.StatusOK, nil
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
	if memberRequest.NRA != "" && !util.IsValidNRA(memberRequest.NRA) {
		return dto.MemberResponse{}, http.StatusBadRequest, fmt.Errorf("invalid NRA format")
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
			// Jika angkatan tidak ada, buat angkatan baru
			if err == sql.ErrNoRows {
				// Format nama angkatan berdasarkan ID (misal: "015" -> "Angkatan 015")
				newAngkatan := model.Angkatan{
					IdAngkatan:   memberRequest.Angkatan,
					NamaAngkatan: fmt.Sprintf("Angkatan %s", memberRequest.Angkatan),
				}
				
				// Tambahkan angkatan baru ke database
				getAngkatan, err = m.MemberRepo.AddAngkatan(ctx, tx, newAngkatan)
				if err != nil {
					return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to create new angkatan: %v", err)
				}
			} else {
				return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to get angkatan: %v", err)
			}
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
		Email:    getMember.Email,
		NoHP:     getMember.NoHP,
		Password: getMember.Password,
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

// GetAllMember implements MemberService.
func (m *memberServiceImpl) GetAllMember(ctx context.Context) ([]dto.MemberResponse, int, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return []dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	members, err := m.MemberRepo.GetAllMember(ctx, tx)
	if err != nil {
		return []dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to get all member: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return []dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return helper.ConvertMemberToListResDTO(members), http.StatusOK, nil
}

// GetMemberById implements MemberService.
func (m *memberServiceImpl) GetMemberById(ctx context.Context, id string) (dto.MemberResponse, int, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	getMember, err := m.MemberRepo.GetMemberById(ctx, tx, id)
	if err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to to get member: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to commit transaction: %v", err)
	}
	return helper.ConvertMemberToResponseDTO(getMember), http.StatusOK, nil

}

// DeleteMember implements MemberService.
func (m *memberServiceImpl) DeleteMember(ctx context.Context, id string) (int, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	getMember, err := m.MemberRepo.GetMemberById(ctx, tx, id)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to get member: %v", err)
	}

	err = m.MemberRepo.DeleteMember(ctx, tx, getMember)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to delete member: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return http.StatusOK, nil
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

	token, err := helper.GenerateJWT(loginRequest.NRA, member.Nama, member.StatusKeanggotaan)
	if err != nil {
		return "", http.StatusBadRequest, fmt.Errorf("failed to generate token: %v", err)
	}

	return token, http.StatusOK, nil
}

// LoginToken implements MemberService.
func (m *memberServiceImpl) LoginToken(ctx context.Context, r *http.Request) (string, int, error) {
	// Parse request body JSON
	var loginRequest dto.LoginTokenRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&loginRequest); err != nil {
		return "", http.StatusBadRequest, fmt.Errorf("invalid JSON format: %v", err)
	}

	tx, err := m.DB.Begin()
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	if loginRequest.Token == "" {
		return "", http.StatusBadRequest, fmt.Errorf("field token cannot be empty")
	}

	member, err := m.MemberRepo.GetMemberByToken(ctx, tx, loginRequest.Token)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("failed to get member: %v", err)
	}

	// Generate JWT token
	jwtToken, err := helper.GenerateJWT(member.NRA.String, member.Nama, member.StatusKeanggotaan)
	if err != nil {
		return "", http.StatusBadRequest, fmt.Errorf("failed to generate token: %v", err)
	}

	// Hapus token dari database setelah berhasil login (token sekali pakai)
	err = m.MemberRepo.UpdateMemberToken(ctx, tx, member.IdMember, "")
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("failed to clear login token: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return jwtToken, http.StatusOK, nil
}

// GetProfile implements MemberService.
func (m *memberServiceImpl) GetProfile(ctx context.Context, r *http.Request) (dto.ProfileResponse, int, error) {
	// Ambil token dari header Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return dto.ProfileResponse{}, http.StatusUnauthorized, fmt.Errorf("authorization header is required")
	}

	// Format: "Bearer <token>"
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return dto.ProfileResponse{}, http.StatusUnauthorized, fmt.Errorf("invalid authorization header format")
	}

	tokenString := tokenParts[1]

	// Validasi token JWT
	claims, err := helper.ValidateJWT(tokenString)
	if err != nil {
		return dto.ProfileResponse{}, http.StatusUnauthorized, fmt.Errorf("invalid or expired token: %v", err)
	}

	// Mulai transaksi database
	tx, err := m.DB.Begin()
	if err != nil {
		return dto.ProfileResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Ambil data member berdasarkan NRA dari claims
	member, err := m.MemberRepo.GetMemberByNRA(ctx, tx, claims.NRA)
	if err != nil {
		return dto.ProfileResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to get member profile: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return dto.ProfileResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return helper.ConvertMemberToProfileResponseDTO(member), http.StatusOK, nil
}

// SetPassword implements MemberService.
func (m *memberServiceImpl) SetPassword(ctx context.Context, r *http.Request) (int, error) {
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

	// Parse request body JSON
	var request dto.SetPasswordRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		return http.StatusBadRequest, fmt.Errorf("invalid JSON format: %v", err)
	}

	// Validasi input
	if request.Password == "" {
		return http.StatusBadRequest, fmt.Errorf("password is required")
	}

	// Hash password
	hashedPassword, err := helper.HashPassword(request.Password)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to hash password: %v", err)
	}

	// Mulai transaksi database
	tx, err := m.DB.Begin()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Ambil data member berdasarkan NRA dari claims untuk mendapatkan ID
	member, err := m.MemberRepo.GetMemberByNRA(ctx, tx, claims.NRA)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to get member: %v", err)
	}

	// Update password
	err = m.MemberRepo.UpdateMemberPassword(ctx, tx, member.IdMember, hashedPassword)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to update password: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return http.StatusOK, nil
}

// CompleteProfile implements MemberService.
func (m *memberServiceImpl) CompleteProfile(ctx context.Context, r *http.Request) (dto.MemberResponse, int, error) {
	// Ambil token dari header Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return dto.MemberResponse{}, http.StatusUnauthorized, fmt.Errorf("authorization header is required")
	}

	// Format: "Bearer <token>"
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return dto.MemberResponse{}, http.StatusUnauthorized, fmt.Errorf("invalid authorization header format")
	}

	tokenString := tokenParts[1]

	// Validasi token JWT
	claims, err := helper.ValidateJWT(tokenString)
	if err != nil {
		return dto.MemberResponse{}, http.StatusUnauthorized, fmt.Errorf("invalid or expired token: %v", err)
	}

	// Parse multipart form data
	err = r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		return dto.MemberResponse{}, http.StatusBadRequest, fmt.Errorf("failed to parse multipart form: %v", err)
	}

	// Ambil data dari form
	tanggalStr := r.FormValue("tanggal_dikukuhkan")
	request := dto.CompleteProfileRequest{
		Email:       r.FormValue("email"),
		NoHP:        r.FormValue("nomor_hp"),
		NamaLengkap: r.FormValue("nama"),
	}

	// Parse tanggal dikukuhkan
	if tanggalStr != "" {
		parsedTanggal, err := time.Parse("02-01-2006", tanggalStr)
		if err != nil {
			return dto.MemberResponse{}, http.StatusBadRequest, fmt.Errorf("invalid date format, use dd-mm-yyyy: %v", err)
		}
		request.TanggalDikukuhkan = &util.CustomDate{Time: parsedTanggal}
	}

	// Validasi input
	if request.Email == "" || request.NoHP == "" {
		return dto.MemberResponse{}, http.StatusBadRequest, fmt.Errorf("email and nomor_hp are required")
	}

	if request.TanggalDikukuhkan == nil {
		return dto.MemberResponse{}, http.StatusBadRequest, fmt.Errorf("tanggal_dikukuhkan is required")
	}

	if request.NamaLengkap == "" {
		return dto.MemberResponse{}, http.StatusBadRequest, fmt.Errorf("nama is required")
	}

	// Handle file upload
	file, header, err := r.FormFile("foto")
	if file == nil {
		return dto.MemberResponse{}, http.StatusBadRequest, fmt.Errorf("foto file is required")
	}
	if err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to get foto file: %v", err)
	}
	defer file.Close()

	// Generate filename menggunakan NRA dan nama lengkap
	fileName := fmt.Sprintf("%s_%s%s", claims.NRA, strings.ReplaceAll(request.NamaLengkap, " ", "_"), filepath.Ext(header.Filename))

	// Buat folder ./uploads jika belum ada
	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.Mkdir(uploadDir, os.ModePerm); err != nil {
			return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to create upload dir: %v", err)
		}
	}

	// Simpan file
	filePath := filepath.Join(uploadDir, fileName)
	out, err := os.Create(filePath)
	if err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to save file: %v", err)
	}

	// Set foto filename ke request
	request.Foto = fileName

	// Mulai transaksi database
	tx, err := m.DB.Begin()
	if err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Ambil data member berdasarkan NRA dari claims untuk mendapatkan ID
	member, err := m.MemberRepo.GetMemberByNRA(ctx, tx, claims.NRA)
	if err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to get member: %v", err)
	}

	// Update profile (email, nomor HP, nama lengkap, foto, dan tanggal dikukuhkan)
	err = m.MemberRepo.UpdateMemberProfile(ctx, tx, member.IdMember, request.Email, request.NoHP, request.NamaLengkap, request.Foto, request.TanggalDikukuhkan)
	if err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to update profile: %v", err)
	}

	// Ambil data member yang sudah diupdate
	updatedMember, err := m.MemberRepo.GetMemberById(ctx, tx, member.IdMember)
	if err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to get updated member: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return dto.MemberResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return helper.ConvertMemberToResponseDTO(updatedMember), http.StatusOK, nil
}
