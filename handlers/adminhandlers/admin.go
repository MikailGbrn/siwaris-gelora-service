package adminhandlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"siwaris-gelora-service/db"
	"siwaris-gelora-service/email"
)

// StatusUpdateRequest represents payload for changing application state
type StatusUpdateRequest struct {
	Status              string `json:"status"`
	AdminNotes          string `json:"admin_notes"`
	EstimatedCompletion string `json:"estimated_completion"`
}

// AdminListApplicationsHandler lists all applications in descending order of creation
func AdminListApplicationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.DB.Query(`SELECT id, registration_number, status, applicant_name, applicant_nik, applicant_kk, 
		applicant_address, applicant_phone, applicant_email, heir_name, death_date, relationship, is_divorced, 
		file_permohonan, file_pengantar_rt_rw, file_pernyataan_kebenaran, file_sptjm, 
		file_ktp_pewaris, file_ktp_ahli_waris, file_kk_ahli_waris, file_akta_lahir_ahli_waris, file_ktp_saksi, 
		file_kematian_ahli_waris_wafat_lebih_dulu, file_pendukung_lainnya, file_surat_nikah_pewaris, 
		file_ktp_suami, file_ktp_istri, file_akta_cerai_pewaris, 
		admin_notes, estimated_completion, created_at, updated_at FROM applications ORDER BY created_at DESC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	apps := []db.Application{}
	for rows.Next() {
		var app db.Application
		err := rows.Scan(
			&app.ID, &app.RegistrationNumber, &app.Status, &app.ApplicantName, &app.ApplicantNik, &app.ApplicantKk,
			&app.ApplicantAddress, &app.ApplicantPhone, &app.ApplicantEmail, &app.HeirName, &app.DeathDate, &app.Relationship, &app.IsDivorced,
			&app.FilePermohonan, &app.FilePengantarRtRw, &app.FilePernyataanKebenaran, &app.FileSptjm,
			&app.FileKtpPewaris, &app.FileKtpAhliWaris, &app.FileKkAhliWaris, &app.FileAktaLahirAhliWaris, &app.FileKtpSaksi,
			&app.FileKematianAhliWarisWafatLebihDulu, &app.FilePendukungLainnya, &app.FileSuratNikahPewaris,
			&app.FileKtpSuami, &app.FileKtpIstri, &app.FileAktaCeraiPewaris,
			&app.AdminNotes, &app.EstimatedCompletion, &app.CreatedAt, &app.UpdatedAt,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		apps = append(apps, app)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apps)
}

// AdminGetApplicationHandler returns a single application by ID
func AdminGetApplicationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}

	var app db.Application
	query := `SELECT id, registration_number, status, applicant_name, applicant_nik, applicant_kk, 
	          applicant_address, applicant_phone, applicant_email, heir_name, death_date, relationship, is_divorced, 
	          file_permohonan, file_pengantar_rt_rw, file_pernyataan_kebenaran, file_sptjm, 
	          file_ktp_pewaris, file_ktp_ahli_waris, file_kk_ahli_waris, file_akta_lahir_ahli_waris, file_ktp_saksi, 
	          file_kematian_ahli_waris_wafat_lebih_dulu, file_pendukung_lainnya, file_surat_nikah_pewaris, 
	          file_ktp_suami, file_ktp_istri, file_akta_cerai_pewaris, 
	          admin_notes, estimated_completion, created_at, updated_at 
	          FROM applications WHERE id = ?`

	err = db.DB.QueryRow(query, id).Scan(
		&app.ID, &app.RegistrationNumber, &app.Status, &app.ApplicantName, &app.ApplicantNik, &app.ApplicantKk,
		&app.ApplicantAddress, &app.ApplicantPhone, &app.ApplicantEmail, &app.HeirName, &app.DeathDate, &app.Relationship, &app.IsDivorced,
		&app.FilePermohonan, &app.FilePengantarRtRw, &app.FilePernyataanKebenaran, &app.FileSptjm,
		&app.FileKtpPewaris, &app.FileKtpAhliWaris, &app.FileKkAhliWaris, &app.FileAktaLahirAhliWaris, &app.FileKtpSaksi,
		&app.FileKematianAhliWarisWafatLebihDulu, &app.FilePendukungLainnya, &app.FileSuratNikahPewaris,
		&app.FileKtpSuami, &app.FileKtpIstri, &app.FileAktaCeraiPewaris,
		&app.AdminNotes, &app.EstimatedCompletion, &app.CreatedAt, &app.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Permohonan tidak ditemukan", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(app)
}

// AdminUpdateStatusHandler updates status, estimated completion, and notes for an application
func AdminUpdateStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}

	var req StatusUpdateRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Payload tidak valid", http.StatusBadRequest)
		return
	}

	// Fetch current details to trigger email correctly
	var app db.Application
	err = db.DB.QueryRow("SELECT applicant_name, applicant_email, registration_number, status FROM applications WHERE id = ?", id).
		Scan(&app.ApplicantName, &app.ApplicantEmail, &app.RegistrationNumber, &app.Status)
	if err == sql.ErrNoRows {
		http.Error(w, "Permohonan tidak ditemukan", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update DB
	_, err = db.DB.Exec(`
		UPDATE applications 
		SET status = ?, admin_notes = ?, estimated_completion = ?, updated_at = ?
		WHERE id = ?
	`, req.Status, req.AdminNotes, req.EstimatedCompletion, time.Now(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If status changed, send email notification
	if app.Status != req.Status {
		go email.SendStatusUpdateEmail(app.ApplicantEmail, app.ApplicantName, app.RegistrationNumber, req.Status, req.AdminNotes)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}
