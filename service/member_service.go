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

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/syrlramadhan/database-anggota-api/dto"
	"github.com/syrlramadhan/database-anggota-api/helper"
	"github.com/syrlramadhan/database-anggota-api/model"
	"github.com/syrlramadhan/database-anggota-api/repository"
	"github.com/syrlramadhan/database-anggota-api/util"
	"golang.org/x/crypto/bcrypt"
)

type MemberService interface {
	AddMember(ctx context.Context, r *http.Request) (dto.MemberResponse, int, error)
	Login(ctx context.Context, loginRequest dto.LoginRequest) (string, int, error)
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

// Function for hash password
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Function to verify password
func verifyPassword(storedHash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	return err == nil
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

	if memberRequest.Nama == "" {
		return dto.MemberResponse{}, http.StatusBadRequest, fmt.Errorf("field name cannot be empty")
	} else if memberRequest.Angkatan == "" {
		return dto.MemberResponse{}, http.StatusBadRequest, fmt.Errorf("field generation cannot be empty")
	} else if memberRequest.StatusKeanggotaan == "" {
		return dto.MemberResponse{}, http.StatusBadRequest, fmt.Errorf("field membership status cannot be empty")
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

	hassedPass, err := hashPassword(memberRequest.Password)
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

type Claims struct {
	NRA  string `json:"nra"`
	Nama string `json:"nama"`
	jwt.StandardClaims
}

func (m *memberServiceImpl) GenerateJWT(nra, nama string) (string, error) {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	jwtKey := os.Getenv("JWT_SECRET")
	expirationTime := time.Now().Add(5 * time.Minute)

	claims := &Claims{
		NRA: nra,
		Nama: nama,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(jwtKey))
}

// Login implements MemberService.
func (m *memberServiceImpl) Login(ctx context.Context, loginRequest dto.LoginRequest) (string, int, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Commit()

	member, err := m.MemberRepo.GetMemberByNRA(ctx, tx, loginRequest.NRA)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("failed to get member: %v", err)
	}

	if !verifyPassword(member.Password.String, loginRequest.Password) {
		return "", http.StatusBadRequest, fmt.Errorf("invalid nra or password: %v", err)
	}

	token, err := m.GenerateJWT(loginRequest.NRA, member.Nama)
	if err != nil {
		return "", http.StatusBadRequest, fmt.Errorf("failed to generate token: %v", err)
	}

	return token, http.StatusOK, nil
}
