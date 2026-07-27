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
	          relationship, is_divorced, file_permohonan, file_pengantar_rt_rw, file_pernyataan_kebenaran,
	          file_sptjm, file_ktp_pewaris, file_ktp_ahli_waris, file_kk_ahli_waris,
	          file_akta_lahir_ahli_waris, file_ktp_saksi, file_kematian_ahli_waris_wafat_lebih_dulu,
	          file_pendukung_lainnya, file_surat_nikah_pewaris, file_ktp_suami, file_ktp_istri,
	          file_akta_cerai_pewaris, admin_notes, estimated_completion, created_at, updated_at 
	          FROM applications WHERE registration_number = ? AND applicant_nik = ?`

	err := db.DB.QueryRow(query, regNum, nik).Scan(
		&app.ID, &app.RegistrationNumber, &app.Status, &app.ApplicantName, &app.ApplicantNik, &app.ApplicantKk,
		&app.ApplicantAddress, &app.ApplicantPhone, &app.ApplicantEmail, &app.HeirName, &app.DeathDate,
		&app.Relationship, &app.IsDivorced, &app.FilePermohonan, &app.FilePengantarRtRw, &app.FilePernyataanKebenaran,
		&app.FileSptjm, &app.FileKtpPewaris, &app.FileKtpAhliWaris, &app.FileKkAhliWaris,
		&app.FileAktaLahirAhliWaris, &app.FileKtpSaksi, &app.FileKematianAhliWarisWafatLebihDulu,
		&app.FilePendukungLainnya, &app.FileSuratNikahPewaris, &app.FileKtpSuami, &app.FileKtpIstri,
		&app.FileAktaCeraiPewaris, &app.AdminNotes, &app.EstimatedCompletion, &app.CreatedAt, &app.UpdatedAt,
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
