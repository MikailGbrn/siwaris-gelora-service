package adminhandlers

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"siwaris-gelora-service/db"
	"siwaris-gelora-service/email"
)

// StatusUpdateRequest represents payload for changing application state
type StatusUpdateRequest struct {
	Status              string `json:"status"`
	AdminNotes          string `json:"admin_notes"`
	EstimatedCompletion string `json:"estimated_completion"`
	RejectedFiles       string `json:"rejected_files"` // JSON array string e.g. '["file_ktp_ahli_waris"]'
}

// AdminListApplicationsHandler lists all applications in descending order of creation
func AdminListApplicationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.DB.Query(db.Rebind(`SELECT id, registration_number, status, applicant_name, applicant_nik, applicant_kk, 
		applicant_address, applicant_phone, applicant_email, heir_name, death_date, relationship, is_divorced, 
		file_permohonan, file_pengantar_rt_rw, file_pernyataan_kebenaran, file_sptjm, 
		file_ktp_pewaris, file_ktp_ahli_waris, file_kematian_pewaris, file_kk_ahli_waris, file_akta_lahir_ahli_waris, file_ktp_saksi, 
		file_kematian_ahli_waris_wafat_lebih_dulu, file_pendukung_lainnya, file_surat_nikah_pewaris, 
		file_ktp_suami, file_ktp_istri, file_akta_cerai_pewaris, file_surat_kuasa, rejected_files, 
		admin_notes, estimated_completion, created_at, updated_at FROM applications ORDER BY created_at DESC`))
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
			&app.FileKtpPewaris, &app.FileKtpAhliWaris, &app.FileKematianPewaris, &app.FileKkAhliWaris, &app.FileAktaLahirAhliWaris, &app.FileKtpSaksi,
			&app.FileKematianAhliWarisWafatLebihDulu, &app.FilePendukungLainnya, &app.FileSuratNikahPewaris,
			&app.FileKtpSuami, &app.FileKtpIstri, &app.FileAktaCeraiPewaris, &app.FileSuratKuasa, &app.RejectedFiles,
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
	          file_ktp_pewaris, file_ktp_ahli_waris, file_kematian_pewaris, file_kk_ahli_waris, file_akta_lahir_ahli_waris, file_ktp_saksi, 
	          file_kematian_ahli_waris_wafat_lebih_dulu, file_pendukung_lainnya, file_surat_nikah_pewaris, 
	          file_ktp_suami, file_ktp_istri, file_akta_cerai_pewaris, file_surat_kuasa, rejected_files, 
	          admin_notes, estimated_completion, created_at, updated_at 
	          FROM applications WHERE id = ?`

	err = db.DB.QueryRow(db.Rebind(query), id).Scan(
		&app.ID, &app.RegistrationNumber, &app.Status, &app.ApplicantName, &app.ApplicantNik, &app.ApplicantKk,
		&app.ApplicantAddress, &app.ApplicantPhone, &app.ApplicantEmail, &app.HeirName, &app.DeathDate, &app.Relationship, &app.IsDivorced,
		&app.FilePermohonan, &app.FilePengantarRtRw, &app.FilePernyataanKebenaran, &app.FileSptjm,
		&app.FileKtpPewaris, &app.FileKtpAhliWaris, &app.FileKematianPewaris, &app.FileKkAhliWaris, &app.FileAktaLahirAhliWaris, &app.FileKtpSaksi,
		&app.FileKematianAhliWarisWafatLebihDulu, &app.FilePendukungLainnya, &app.FileSuratNikahPewaris,
		&app.FileKtpSuami, &app.FileKtpIstri, &app.FileAktaCeraiPewaris, &app.FileSuratKuasa, &app.RejectedFiles,
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
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
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
	var draftFilename string
	var draftBase64 string
	var draftSavedPath string

	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		err = r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "Gagal memproses form", http.StatusBadRequest)
			return
		}

		req.Status = r.FormValue("status")
		req.AdminNotes = r.FormValue("admin_notes")
		req.EstimatedCompletion = r.FormValue("estimated_completion")
		req.RejectedFiles = r.FormValue("rejected_files")

		file, header, err := r.FormFile("file_draft")
		if err == nil {
			defer file.Close()
			os.MkdirAll("./uploads", os.ModePerm)
			draftFilename = header.Filename
			draftSavedPath = fmt.Sprintf("uploads/draft_%d_%s", time.Now().Unix(), header.Filename)

			out, err := os.Create(draftSavedPath)
			if err == nil {
				defer out.Close()
				buf := new(bytes.Buffer)
				io.Copy(out, io.TeeReader(file, buf))
				draftBase64 = base64.StdEncoding.EncodeToString(buf.Bytes())
			}
		}
	} else {
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Payload tidak valid", http.StatusBadRequest)
			return
		}
	}

	// Fetch current details to trigger email correctly
	var app db.Application
	err = db.DB.QueryRow(db.Rebind("SELECT applicant_name, applicant_email, registration_number, status, file_draft FROM applications WHERE id = ?"), id).
		Scan(&app.ApplicantName, &app.ApplicantEmail, &app.RegistrationNumber, &app.Status, &app.FileDraft)
	if err == sql.ErrNoRows {
		http.Error(w, "Permohonan tidak ditemukan", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update DB (including rejected_files text column and file_draft if uploaded)
	if draftSavedPath != "" {
		_, err = db.DB.Exec(db.Rebind(`
			UPDATE applications 
			SET status = ?, admin_notes = ?, estimated_completion = ?, rejected_files = ?, file_draft = ?, updated_at = ?
			WHERE id = ?
		`), req.Status, req.AdminNotes, req.EstimatedCompletion, req.RejectedFiles, draftSavedPath, time.Now(), id)
	} else {
		_, err = db.DB.Exec(db.Rebind(`
			UPDATE applications 
			SET status = ?, admin_notes = ?, estimated_completion = ?, rejected_files = ?, updated_at = ?
			WHERE id = ?
		`), req.Status, req.AdminNotes, req.EstimatedCompletion, req.RejectedFiles, time.Now(), id)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send email notification
	if draftBase64 != "" {
		go email.SendStatusUpdateWithAttachmentEmail(app.ApplicantEmail, app.ApplicantName, app.RegistrationNumber, req.Status, req.AdminNotes, draftFilename, draftBase64)
	} else if app.Status != req.Status {
		go email.SendStatusUpdateEmail(app.ApplicantEmail, app.ApplicantName, app.RegistrationNumber, req.Status, req.AdminNotes)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}
