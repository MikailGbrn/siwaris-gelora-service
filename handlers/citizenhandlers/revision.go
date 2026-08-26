package citizenhandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"siwaris-gelora-service/db"
	"siwaris-gelora-service/email"
)

// RevisionHandler processes document corrections from citizens
func RevisionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload size (5MB per file, max 30MB total for revisions)
	err := r.ParseMultipartForm(30 << 20)
	if err != nil {
		http.Error(w, "Ukuran berkas perbaikan terlalu besar", http.StatusBadRequest)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID Permohonan tidak valid", http.StatusBadRequest)
		return
	}

	// Fetch current details
	var app db.Application
	err = db.DB.QueryRow(db.Rebind("SELECT applicant_name, applicant_email, registration_number, rejected_files FROM applications WHERE id = ?"), id).
		Scan(&app.ApplicantName, &app.ApplicantEmail, &app.RegistrationNumber, &app.RejectedFiles)
	if err != nil {
		http.Error(w, "Permohonan tidak ditemukan", http.StatusNotFound)
		return
	}

	// Extract which keys are allowed to be replaced (only rejected ones)
	rejectedKeys := []string{}
	if app.RejectedFiles != "" {
		// Clean up e.g. ["file_ktp_ahli_waris"] -> file_ktp_ahli_waris
		cleaned := strings.ReplaceAll(app.RejectedFiles, "[", "")
		cleaned = strings.ReplaceAll(cleaned, "]", "")
		cleaned = strings.ReplaceAll(cleaned, "\"", "")
		cleaned = strings.ReplaceAll(cleaned, " ", "")
		if cleaned != "" {
			rejectedKeys = strings.Split(cleaned, ",")
		}
	}

	// Save new uploads
	updatedFields := []string{}
	args := []interface{}{}

	for _, key := range rejectedKeys {
		file, handler, err := r.FormFile(key)
		if err != nil {
			// Citizen might not submit all files at once, but they should submit the ones they are fixing
			continue
		}
		defer file.Close()

		path, err := saveUploadedFile(file, handler.Filename)
		if err != nil {
			http.Error(w, "Gagal menulis berkas baru", http.StatusInternalServerError)
			return
		}

		updatedFields = append(updatedFields, key+" = ?")
		args = append(args, path)
	}

	if len(updatedFields) == 0 {
		http.Error(w, "Tidak ada berkas perbaikan yang dikirim", http.StatusBadRequest)
		return
	}

	// Append common fields to update
	updatedFields = append(updatedFields, "status = ?", "rejected_files = ?", "updated_at = ?")
	args = append(args, "Menunggu Verifikasi", "", time.Now(), id)

	// Build update query
	query := fmt.Sprintf("UPDATE applications SET %s WHERE id = ?", strings.Join(updatedFields, ", "))

	_, err = db.DB.Exec(db.Rebind(query), args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send revision received notification email for resubmission review
	go email.SendRevisionReceivedEmail(app.ApplicantEmail, app.ApplicantName, app.RegistrationNumber)

	// Trigger admin email notification in background
	adminEmail := os.Getenv("ADMIN_NOTIFICATION_EMAIL")
	if adminEmail == "" {
		adminEmail = "siwarisgelora@gmail.com"
	}
	go email.SendAdminRevisionSubmittedEmail(adminEmail, app.ApplicantName, app.RegistrationNumber)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Dokumen perbaikan berhasil diperbarui. Status permohonan dikembalikan ke Menunggu Verifikasi.",
	})
}
