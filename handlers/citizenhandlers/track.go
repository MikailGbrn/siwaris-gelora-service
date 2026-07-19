package citizenhandlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"siwaris-gelora-service/db"
)

// TrackHandler searches for application status by registration number and NIK
func TrackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	regNum := r.URL.Query().Get("reg_num")
	nik := r.URL.Query().Get("nik")

	if regNum == "" || nik == "" {
		http.Error(w, "Nomor registrasi dan NIK harus diisi", http.StatusBadRequest)
		return
	}

	var app db.Application
	query := `SELECT id, registration_number, status, applicant_name, applicant_nik, applicant_kk, 
	          applicant_address, applicant_phone, applicant_email, heir_name, death_date, 
	          relationship, file_ktp, file_kk, file_death_cert, file_rt_rw, file_other, 
	          admin_notes, estimated_completion, created_at, updated_at 
	          FROM applications WHERE registration_number = ? AND applicant_nik = ?`

	err := db.DB.QueryRow(query, regNum, nik).Scan(
		&app.ID, &app.RegistrationNumber, &app.Status, &app.ApplicantName, &app.ApplicantNik, &app.ApplicantKk,
		&app.ApplicantAddress, &app.ApplicantPhone, &app.ApplicantEmail, &app.HeirName, &app.DeathDate,
		&app.Relationship, &app.FileKtp, &app.FileKk, &app.FileDeathCert, &app.FileRtRw, &app.FileOther,
		&app.AdminNotes, &app.EstimatedCompletion, &app.CreatedAt, &app.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Permohonan tidak ditemukan. Silakan periksa kembali Nomor Registrasi dan NIK Anda.", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(app)
}
