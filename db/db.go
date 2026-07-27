package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Application struct {
	ID                                    int64     `json:"id"`
	RegistrationNumber                    string    `json:"registration_number"`
	Status                                string    `json:"status"` // Menunggu Verifikasi, Perlu Perbaikan, Sedang Diproses, Menunggu TTD, Selesai
	ApplicantName                         string    `json:"applicant_name"`
	ApplicantNik                          string    `json:"applicant_nik"`
	ApplicantKk                           string    `json:"applicant_kk"`
	ApplicantAddress                      string    `json:"applicant_address"`
	ApplicantPhone                        string    `json:"applicant_phone"`
	ApplicantEmail                        string    `json:"applicant_email"`
	HeirName                              string    `json:"heir_name"`
	DeathDate                             string    `json:"death_date"`
	Relationship                          string    `json:"relationship"` // Anak Kandung, Istri / Suami, Orang Tua, Saudara Kandung, Lainnya
	IsDivorced                            string    `json:"is_divorced"`   // Ya, Tidak, N/A
	FilePermohonan                        string    `json:"file_permohonan"`
	FilePengantarRtRw                     string    `json:"file_pengantar_rt_rw"`
	FilePernyataanKebenaran               string    `json:"file_pernyataan_kebenaran"`
	FileSptjm                             string    `json:"file_sptjm"`
	FileKtpPewaris                        string    `json:"file_ktp_pewaris"`
	FileKtpAhliWaris                      string    `json:"file_ktp_ahli_waris"`
	FileKkAhliWaris                       string    `json:"file_kk_ahli_waris"`
	FileAktaLahirAhliWaris                string    `json:"file_akta_lahir_ahli_waris"`
	FileKtpSaksi                          string    `json:"file_ktp_saksi"`
	FileKematianAhliWarisWafatLebihDulu   string    `json:"file_kematian_ahli_waris_wafat_lebih_dulu"`
	FilePendukungLainnya                  string    `json:"file_pendukung_lainnya"`
	FileSuratNikahPewaris                 string    `json:"file_surat_nikah_pewaris"`
	FileKtpSuami                          string    `json:"file_ktp_suami"`
	FileKtpIstri                          string    `json:"file_ktp_istri"`
	FileAktaCeraiPewaris                  string    `json:"file_akta_cerai_pewaris"`
	AdminNotes                            string    `json:"admin_notes"`
	EstimatedCompletion                   string    `json:"estimated_completion"`
	CreatedAt                             time.Time `json:"created_at"`
	UpdatedAt                             time.Time `json:"updated_at"`
}

// DB is the global database connection handle exported for other packages
var DB *sql.DB

// InitDB initializes the SQLite database
func InitDB(filepath string) *sql.DB {
	var err error
	DB, err = sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatal(err)
	}

	if DB == nil {
		log.Fatal("DB connection could not be opened.")
	}

	createTable()
	return DB
}

func createTable() {
	createTableSQL := `CREATE TABLE IF NOT EXISTS applications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		registration_number TEXT NOT NULL UNIQUE,
		status TEXT NOT NULL,
		applicant_name TEXT NOT NULL,
		applicant_nik TEXT NOT NULL,
		applicant_kk TEXT NOT NULL,
		applicant_address TEXT NOT NULL,
		applicant_phone TEXT NOT NULL,
		applicant_email TEXT NOT NULL,
		heir_name TEXT NOT NULL,
		death_date TEXT NOT NULL,
		relationship TEXT NOT NULL,
		is_divorced TEXT NOT NULL,
		file_permohonan TEXT NOT NULL,
		file_pengantar_rt_rw TEXT NOT NULL,
		file_pernyataan_kebenaran TEXT NOT NULL,
		file_sptjm TEXT NOT NULL,
		file_ktp_pewaris TEXT NOT NULL,
		file_ktp_ahli_waris TEXT NOT NULL,
		file_kk_ahli_waris TEXT NOT NULL,
		file_akta_lahir_ahli_waris TEXT NOT NULL,
		file_ktp_saksi TEXT NOT NULL,
		file_kematian_ahli_waris_wafat_lebih_dulu TEXT,
		file_pendukung_lainnya TEXT,
		file_surat_nikah_pewaris TEXT,
		file_ktp_suami TEXT,
		file_ktp_istri TEXT,
		file_akta_cerai_pewaris TEXT,
		admin_notes TEXT,
		estimated_completion TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	statement, err := DB.Prepare(createTableSQL)
	if err != nil {
		log.Fatal("Error preparing schema migration: ", err)
	}
	_, err = statement.Exec()
	if err != nil {
		log.Fatal("Error executing schema migration: ", err)
	}
	log.Println("Database tables initialized successfully.")
}
