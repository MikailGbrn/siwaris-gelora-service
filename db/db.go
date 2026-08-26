package db

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
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
	FileKematianPewaris                   string    `json:"file_kematian_pewaris"`
	FileKkAhliWaris                       string    `json:"file_kk_ahli_waris"`
	FileAktaLahirAhliWaris                string    `json:"file_akta_lahir_ahli_waris"`
	FileKtpSaksi                          string    `json:"file_ktp_saksi"`
	FileKematianAhliWarisWafatLebihDulu   string    `json:"file_kematian_ahli_waris_wafat_lebih_dulu"`
	FilePendukungLainnya                  string    `json:"file_pendukung_lainnya"`
	FileSuratNikahPewaris                 string    `json:"file_surat_nikah_pewaris"`
	FileKtpSuami                          string    `json:"file_ktp_suami"`
	FileKtpIstri                          string    `json:"file_ktp_istri"`
	FileAktaCeraiPewaris                  string    `json:"file_akta_cerai_pewaris"`
	FileSuratKuasa                        string    `json:"file_surat_kuasa"`
	RejectedFiles                         string    `json:"rejected_files"` // JSON array string of rejected keys e.g. ["file_ktp_ahli_waris"]
	AdminNotes                            string    `json:"admin_notes"`
	EstimatedCompletion                   string    `json:"estimated_completion"`
	CreatedAt                             time.Time `json:"created_at"`
	UpdatedAt                             time.Time `json:"updated_at"`
}

// DB is the global database connection handle exported for other packages
var DB *sql.DB

// DbType tracks the current active database driver ("sqlite" or "postgres")
var DbType = "sqlite"

// InitDB initializes either a PostgreSQL or SQLite connection depending on environment variables
func InitDB(filepath string) *sql.DB {
	var err error
	dbURL := os.Getenv("DATABASE_URL")

	if dbURL != "" {
		DbType = "postgres"
		log.Println("DATABASE_URL detected. Connecting to PostgreSQL database...")
		DB, err = sql.Open("postgres", dbURL)
		if err != nil {
			log.Fatal("Failed to open PostgreSQL connection:", err)
		}
	} else {
		DbType = "sqlite"
		log.Printf("Connecting to local SQLite database at: %s\n", filepath)
		DB, err = sql.Open("sqlite3", filepath)
		if err != nil {
			log.Fatal("Failed to open SQLite connection:", err)
		}
	}

	if DB == nil {
		log.Fatal("DB connection could not be opened.")
	}

	createTable()
	return DB
}

// Rebind replaces '?' placeholders in SQL queries with '$1, $2, ...' when running on PostgreSQL
func Rebind(query string) string {
	if DbType != "postgres" {
		return query
	}

	var buf bytes.Buffer
	paramIndex := 1
	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			buf.WriteString(fmt.Sprintf("$%d", paramIndex))
			paramIndex++
		} else {
			buf.WriteByte(query[i])
		}
	}
	return buf.String()
}

func createTable() {
	var createApplicationsSQL string
	var createAdminsSQL string

	if DbType == "postgres" {
		createApplicationsSQL = `CREATE TABLE IF NOT EXISTS applications (
			id SERIAL PRIMARY KEY,
			registration_number VARCHAR(100) NOT NULL UNIQUE,
			status VARCHAR(100) NOT NULL,
			applicant_name VARCHAR(255) NOT NULL,
			applicant_nik VARCHAR(16) NOT NULL,
			applicant_kk VARCHAR(16) NOT NULL,
			applicant_address TEXT NOT NULL,
			applicant_phone VARCHAR(50) NOT NULL,
			applicant_email VARCHAR(255) NOT NULL,
			heir_name VARCHAR(255) NOT NULL,
			death_date VARCHAR(50) NOT NULL,
			relationship VARCHAR(100) NOT NULL,
			is_divorced VARCHAR(10) NOT NULL,
			file_permohonan TEXT NOT NULL,
			file_pengantar_rt_rw TEXT NOT NULL,
			file_pernyataan_kebenaran TEXT NOT NULL,
			file_sptjm TEXT NOT NULL,
			file_ktp_pewaris TEXT NOT NULL,
			file_ktp_ahli_waris TEXT NOT NULL,
			file_kematian_pewaris TEXT NOT NULL DEFAULT '',
			file_kk_ahli_waris TEXT NOT NULL,
			file_akta_lahir_ahli_waris TEXT NOT NULL,
			file_ktp_saksi TEXT NOT NULL,
			file_kematian_ahli_waris_wafat_lebih_dulu TEXT,
			file_pendukung_lainnya TEXT,
			file_surat_nikah_pewaris TEXT,
			file_ktp_suami TEXT,
			file_akta_cerai_pewaris TEXT,
			file_surat_kuasa TEXT,
			rejected_files TEXT NOT NULL DEFAULT '',
			admin_notes TEXT,
			estimated_completion VARCHAR(100),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`

		createAdminsSQL = `CREATE TABLE IF NOT EXISTS admins (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`
	} else {
		createApplicationsSQL = `CREATE TABLE IF NOT EXISTS applications (
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
			file_kematian_pewaris TEXT NOT NULL DEFAULT '',
			file_kk_ahli_waris TEXT NOT NULL,
			file_akta_lahir_ahli_waris TEXT NOT NULL,
			file_ktp_saksi TEXT NOT NULL,
			file_kematian_ahli_waris_wafat_lebih_dulu TEXT,
			file_pendukung_lainnya TEXT,
			file_surat_nikah_pewaris TEXT,
			file_ktp_suami TEXT,
			file_akta_cerai_pewaris TEXT,
			file_surat_kuasa TEXT,
			rejected_files TEXT NOT NULL DEFAULT '',
			admin_notes TEXT,
			estimated_completion TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`

		createAdminsSQL = `CREATE TABLE IF NOT EXISTS admins (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			name TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`
	}

	_, err := DB.Exec(createApplicationsSQL)
	if err != nil {
		log.Fatal("Error executing applications migration: ", err)
	}

	// Alter table to add file_surat_kuasa if it doesn't exist yet (for backward compatibility)
	if DbType == "postgres" {
		_, _ = DB.Exec("ALTER TABLE applications ADD COLUMN IF NOT EXISTS file_surat_kuasa TEXT DEFAULT '';")
		_, _ = DB.Exec("ALTER TABLE applications ADD COLUMN IF NOT EXISTS file_kematian_pewaris TEXT DEFAULT '';")
	} else {
		_, _ = DB.Exec("ALTER TABLE applications ADD COLUMN file_surat_kuasa TEXT;")
		_, _ = DB.Exec("ALTER TABLE applications ADD COLUMN file_kematian_pewaris TEXT;")
	}

	// Migrate Admins Table
	_, err = DB.Exec(createAdminsSQL)
	if err != nil {
		log.Fatal("Error executing admins migration: ", err)
	}

	log.Println("Database tables initialized successfully.")

	// Auto-seed default admin if table is empty
	var count int
	err = DB.QueryRow(Rebind("SELECT COUNT(*) FROM admins")).Scan(&count)
	if err == nil && count == 0 {
		log.Println("Admins table is empty. Seeding default admin account...")
		hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		if err != nil {
			log.Println("Error generating default admin hash:", err)
			return
		}

		_, err = DB.Exec(Rebind("INSERT INTO admins (email, password_hash, name) VALUES (?, ?, ?)"),
			"admin@gelora.go.id", string(hash), "Administrator Utama")
		if err != nil {
			log.Println("Error seeding default admin:", err)
		} else {
			log.Println("Default admin account seeded successfully (admin@gelora.go.id / password123).")
		}
	}
}
