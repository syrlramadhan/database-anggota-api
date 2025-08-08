package model

import (
	"database/sql"

	"github.com/syrlramadhan/database-anggota-api/util"
)

type Member struct {
	IdMember          string
	NRA               sql.NullString
	Nama              string
	AngkatanID        string
	StatusKeanggotaan string
	JurusanID         sql.NullString
	TanggalDikukuhkan *util.CustomDate
	JenisKelamin      sql.NullString
	Email             sql.NullString
	NoHP              sql.NullString
	Password          sql.NullString
	Foto              sql.NullString
	LoginToken        sql.NullString

	Angkatan Angkatan
	Jurusan  Jurusan
}

type Angkatan struct {
	IdAngkatan   string
	NamaAngkatan string
}

type Jurusan struct {
	IdJurusan   string
	NamaJurusan string
}
