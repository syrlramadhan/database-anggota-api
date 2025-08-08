package repository

import (
	"context"
	"database/sql"

	"github.com/syrlramadhan/database-anggota-api/model"
)

type MemberRepository interface {
	AddMember(ctx context.Context, tx *sql.Tx, member model.Member) (model.Member, error)
	UpdateMember(ctx context.Context, tx *sql.Tx, member model.Member) (model.Member, error)
	GetJurusanByName(ctx context.Context, tx *sql.Tx, jurusan model.Jurusan, nama_jurusan string) (model.Jurusan, error)
	GetAngkatanById(ctx context.Context, tx *sql.Tx, angkatan model.Angkatan, id_angkatan string) (model.Angkatan, error)
	GetMemberByNRA(ctx context.Context, tx *sql.Tx, nra string) (model.Member, error)
	GetMemberByToken(ctx context.Context, tx *sql.Tx, token string) (model.Member, error)
	GetMemberById(ctx context.Context, tx *sql.Tx, id string) (model.Member, error)
}

type memberRepositoryImpl struct {
}

func NewMemberRepository() MemberRepository {
	return &memberRepositoryImpl{}
}

// AddMember implements MemberRepository.
func (m memberRepositoryImpl) AddMember(ctx context.Context, tx *sql.Tx, member model.Member) (model.Member, error) {
	queryMember := "INSERT INTO member (id_member, nra, nama, angkatan, status_keanggotaan, id_jurusan, tanggal_dikukuhkan, email, no_hp, password, foto) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"

	_, err := tx.ExecContext(ctx, queryMember, member.IdMember, member.NRA, member.Nama, member.AngkatanID, member.StatusKeanggotaan, member.JurusanID, member.TanggalDikukuhkan, member.Email, member.NoHP, member.Password, member.Foto)
	if err != nil {
		return member, err
	}

	return member, nil
}

// UpdateMember implements MemberRepository.
func (m *memberRepositoryImpl) UpdateMember(ctx context.Context, tx *sql.Tx, member model.Member) (model.Member, error) {
	query := "UPDATE member SET nra = ?, nama = ?, angkatan = ?, status_keanggotaan = ?, id_jurusan = ?, tanggal_dikukuhkan = ?, email = ?, no_hp = ?, password = ?, foto = ? WHERE id_member = ?"

	_, err := tx.ExecContext(ctx, query, member.NRA, member.Nama, member.AngkatanID, member.StatusKeanggotaan, member.JurusanID, member.TanggalDikukuhkan, member.Email, member.NoHP, member.Password, member.Foto, member.IdMember)
	if err != nil {
		return member, err
	}

	return member, nil
}

// GetJurusan implements MemberRepository.
func (m *memberRepositoryImpl) GetJurusanByName(ctx context.Context, tx *sql.Tx, jurusan model.Jurusan, nama_jurusan string) (model.Jurusan, error) {
	query := "SELECT id_jurusan, nama_jurusan FROM jurusan WHERE nama_jurusan = ?"

	err := tx.QueryRowContext(ctx, query, nama_jurusan).Scan(&jurusan.IdJurusan, &jurusan.NamaJurusan)
	if err != nil {
		return model.Jurusan{}, err
	}

	return jurusan, nil
}

// GetAngkatanByName implements MemberRepository.
func (m *memberRepositoryImpl) GetAngkatanById(ctx context.Context, tx *sql.Tx, angkatan model.Angkatan, id_angkatan string) (model.Angkatan, error) {
	query := "SELECT id_angkatan, nama_angkatan FROM angkatan WHERE id_angkatan = ?"

	err := tx.QueryRowContext(ctx, query, id_angkatan).Scan(&angkatan.IdAngkatan, &angkatan.NamaAngkatan)
	if err != nil {
		return model.Angkatan{}, err
	}

	return angkatan, nil
}

// GetMemberByNRA implements MemberRepository.
func (m *memberRepositoryImpl) GetMemberByNRA(ctx context.Context, tx *sql.Tx, nra string) (model.Member, error) {
	var member model.Member
	query := "SELECT id_member, nra, nama, angkatan, status_keanggotaan, id_jurusan, tanggal_dikukuhkan, email, no_hp, password, foto, login_token FROM member WHERE nra = ?"

	err := tx.QueryRowContext(ctx, query, nra).Scan(&member.IdMember, &member.NRA, &member.Nama, &member.AngkatanID, &member.StatusKeanggotaan, &member.JurusanID, &member.TanggalDikukuhkan, &member.Email, &member.NoHP, &member.Password, &member.Foto, &member.LoginToken)
	if err != nil {
		return model.Member{}, err
	}

	return member, nil
}

// GetMemberByToken implements MemberRepository.
func (m *memberRepositoryImpl) GetMemberByToken(ctx context.Context, tx *sql.Tx, token string) (model.Member, error) {
	var member model.Member
	query := "SELECT id_member, nra, nama, angkatan, status_keanggotaan, id_jurusan, tanggal_dikukuhkan, email, no_hp, password, foto, login_token FROM member WHERE login_token = ?"

	err := tx.QueryRowContext(ctx, query, token).Scan(&member.IdMember, &member.NRA, &member.Nama, &member.AngkatanID, &member.StatusKeanggotaan, &member.JurusanID, &member.TanggalDikukuhkan, &member.Email, &member.NoHP, &member.Password, &member.Foto, &member.LoginToken)
	if err != nil {
		return model.Member{}, err
	}

	return member, nil
}

// GetMemberById implements MemberRepository.
func (m *memberRepositoryImpl) GetMemberById(ctx context.Context, tx *sql.Tx, id string) (model.Member, error) {
	var member model.Member
	query := "SELECT id_member, nra, nama, angkatan, status_keanggotaan, id_jurusan, tanggal_dikukuhkan, email, no_hp, password, foto, login_token FROM member WHERE id_member = ?"

	err := tx.QueryRowContext(ctx, query, id).Scan(&member.IdMember, &member.NRA, &member.Nama, &member.AngkatanID, &member.StatusKeanggotaan, &member.JurusanID, &member.TanggalDikukuhkan, &member.Email, &member.NoHP, &member.Password, &member.Foto, &member.LoginToken)
	if err != nil {
		return model.Member{}, err
	}

	return member, nil
}
