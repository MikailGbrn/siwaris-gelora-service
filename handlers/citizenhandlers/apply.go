package citizenhandlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"siwaris-gelora-service/db"
	"siwaris-gelora-service/email"
)

// UploadDir points to local folder for storing user documents
const UploadDir = "./uploads"

// EnsureUploadDirExists creates upload folder if it doesn't exist
func EnsureUploadDirExists() {
	if _, err := os.Stat(UploadDir); os.IsNotExist(err) {
		err := os.MkdirAll(UploadDir, 0755)
		if err != nil {
			log.Fatal("Failed to create upload directory:", err)
		}
	}
}

// ApplyHandler handles new application submission with file uploads
func ApplyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload size (5MB per file, max 30MB total)
	err := r.ParseMultipartForm(30 << 20)
	if err != nil {
		http.Error(w, "Ukuran file terlalu besar", http.StatusBadRequest)
		return
	}

	// Retrieve string fields
	applicantName := r.FormValue("applicant_name")
	applicantNik := r.FormValue("applicant_nik")
	applicantKk := r.FormValue("applicant_kk")
	applicantAddress := r.FormValue("applicant_address")
	applicantPhone := r.FormValue("applicant_phone")
	applicantEmail := r.FormValue("applicant_email")
	heirName := r.FormValue("heir_name")
	deathDate := r.FormValue("death_date")
	relationship := r.FormValue("relationship")

	if applicantName == "" || applicantNik == "" || applicantKk == "" || applicantEmail == "" || heirName == "" {
		http.Error(w, "Semua field utama wajib diisi", http.StatusBadRequest)
		return
	}

	EnsureUploadDirExists()

	// Save files
	fileKeys := []string{"file_ktp", "file_kk", "file_death_cert", "file_rt_rw", "file_other"}
	filePaths := make(map[string]string)

	for _, key := range fileKeys {
		file, handler, err := r.FormFile(key)
		if err != nil {
			if key == "file_other" {
				filePaths[key] = ""
				continue
			}
			http.Error(w, fmt.Sprintf("File %s wajib diunggah", key), http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Generate unique name
		filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), handler.Filename)
		destPath := filepath.Join(UploadDir, filename)
		out, err := os.Create(destPath)
		if err != nil {
			http.Error(w, "Gagal menyimpan berkas", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, "Gagal menulis berkas", http.StatusInternalServerError)
			return
		}
		filePaths[key] = "/uploads/" + filename
	}

	// Start transaction
	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert with placeholder registration number
	res, err := tx.Exec(`
		INSERT INTO applications (
			registration_number, status, applicant_name, applicant_nik, applicant_kk, 
			applicant_address, applicant_phone, applicant_email, heir_name, death_date, 
			relationship, file_ktp, file_kk, file_death_cert, file_rt_rw, file_other, 
			admin_notes, estimated_completion, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "TEMP", "Menunggu Verifikasi", applicantName, applicantNik, applicantKk,
		applicantAddress, applicantPhone, applicantEmail, heirName, deathDate,
		relationship, filePaths["file_ktp"], filePaths["file_kk"], filePaths["file_death_cert"],
		filePaths["file_rt_rw"], filePaths["file_other"], "", "", time.Now(), time.Now())

	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Format permanent registration number: SWG-[YEAR]-[4 DIGIT ID] (e.g. SWG-2026-0012)
	regNum := fmt.Sprintf("SWG-%d-%04d", time.Now().Year(), lastID)

	_, err = tx.Exec("UPDATE applications SET registration_number = ? WHERE id = ?", regNum, lastID)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Trigger mock email in background
	go email.SendSubmissionEmail(applicantEmail, applicantName, regNum)

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":             true,
		"registration_number": regNum,
	})
}
