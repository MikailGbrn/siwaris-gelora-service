package adminhandlers

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/jung-kurt/gofpdf/v2"
	"siwaris-gelora-service/db"
)

// GenerateApplicationPDF creates a PDF dossier for the application and writes it to w
func GenerateApplicationPDF(w io.Writer, app *db.Application) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Header / Kop Surat
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(0, 7, "PEMERINTAH PROVINSI DKI JAKARTA", "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 7, "KECAMATAN TANAH ABANG - KELURAHAN GELORA", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 5, "Jl. Gelora No.1, Jakarta Pusat", "", 1, "C", false, 0, "")
	
	// Divider Line
	pdf.SetLineWidth(0.8)
	pdf.Line(10, 32, 200, 32)
	pdf.Ln(10)

	// Document Title
	pdf.SetFont("Arial", "BU", 14)
	pdf.CellFormat(0, 7, "SURAT PERNYATAAN AHLI WARIS", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 5, fmt.Sprintf("Nomor Registrasi: %s", app.RegistrationNumber), "", 1, "C", false, 0, "")
	pdf.Ln(10)

	// Section 1: Data Pemohon
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(0, 6, "I. DATA PEMOHON", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	
	writeField(pdf, "Nama Lengkap", app.ApplicantName)
	writeField(pdf, "NIK", app.ApplicantNik)
	writeField(pdf, "Nomor KK", app.ApplicantKk)
	writeField(pdf, "Alamat", app.ApplicantAddress)
	writeField(pdf, "Nomor HP", app.ApplicantPhone)
	writeField(pdf, "Email", app.ApplicantEmail)
	pdf.Ln(5)

	// Section 2: Data Pewaris
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(0, 6, "II. DATA PEWARIS", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	
	writeField(pdf, "Nama Pewaris (Almarhum/ah)", app.HeirName)
	writeField(pdf, "Tanggal Meninggal", app.DeathDate)
	writeField(pdf, "Hubungan Keluarga", app.Relationship)
	if app.Relationship == "Orang Tua" || app.Relationship == "Istri / Suami" {
		writeField(pdf, "Status Cerai Pewaris", app.IsDivorced)
	}
	pdf.Ln(5)

	// Section 3: Status Administrasi
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(0, 6, "III. STATUS PERMOHONAN", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	writeField(pdf, "Status Saat Ini", app.Status)
	writeField(pdf, "Estimasi Selesai", app.EstimatedCompletion)
	if app.AdminNotes != "" {
		writeField(pdf, "Catatan Petugas", app.AdminNotes)
	}
	pdf.Ln(15)

	// Signatures
	pdf.SetFont("Arial", "", 10)
	leftCol := 60.0
	rightCol := 60.0
	spacing := 20.0
	
	currentY := pdf.GetY()
	
	pdf.SetXY(15, currentY)
	pdf.CellFormat(leftCol, 5, "Pemohon,", "", 0, "C", false, 0, "")
	
	pdf.SetXY(135, currentY)
	pdf.CellFormat(rightCol, 5, "Lurah Kelurahan Gelora,", "", 1, "C", false, 0, "")
	
	pdf.Ln(spacing)
	currentY = pdf.GetY()
	
	pdf.SetXY(15, currentY)
	pdf.SetFont("Arial", "U", 10)
	pdf.CellFormat(leftCol, 5, app.ApplicantName, "", 0, "C", false, 0, "")
	
	pdf.SetXY(135, currentY)
	pdf.SetFont("Arial", "U", 10)
	pdf.CellFormat(rightCol, 5, ".........................................", "", 1, "C", false, 0, "")

	return pdf.Output(w)
}

func writeField(pdf *gofpdf.Fpdf, label, value string) {
	if value == "" {
		value = "-"
	}
	pdf.CellFormat(60, 6, "  "+label, "", 0, "L", false, 0, "")
	pdf.CellFormat(5, 6, ":", "", 0, "C", false, 0, "")
	pdf.MultiCell(0, 6, value, "", "L", false)
}

// AdminDownloadPDFHandler handles request to generate and download the PDF dossier
func AdminDownloadPDFHandler(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=Dossier_%s.pdf", app.RegistrationNumber))
	w.Header().Set("Content-Type", "application/pdf")

	err = GenerateApplicationPDF(w, &app)
	if err != nil {
		log.Println("Error generating PDF:", err)
	}
}
