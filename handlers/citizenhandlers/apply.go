package citizenhandlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
			log.Println("Error creating uploads folder:", err)
		}
	}
}

func uploadToSupabaseStorage(file multipart.File, filename string) (string, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	bucketName := "siwaris-uploads"

	// Read file bytes
	var buf bytes.Buffer
	_, err := io.Copy(&buf, file)
	if err != nil {
		return "", err
	}
	fileBytes := buf.Bytes()

	// Detect content type safely
	detectSize := 512
	if len(fileBytes) < detectSize {
		detectSize = len(fileBytes)
	}
	contentType := http.DetectContentType(fileBytes[:detectSize])

	// Build target URL
	cleanURL := strings.TrimSuffix(supabaseURL, "/")
	targetURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", cleanURL, bucketName, filename)

	req, err := http.NewRequest("POST", targetURL, bytes.NewReader(fileBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Content-Type", contentType)

	// Set a 30-second connection timeout to prevent hanging indefinitely on network issues
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("ERROR: Supabase Storage upload request failed: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		errMsg := string(bodyBytes)
		log.Printf("ERROR: Supabase Storage returned status %s: %s\n", resp.Status, errMsg)
		return "", fmt.Errorf("failed to upload to supabase storage: status %s, response: %s", resp.Status, errMsg)
	}

	// Return the public CDN URL to access the file
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", cleanURL, bucketName, filename)
	return publicURL, nil
}

// saveUploadedFile saves a file to either Supabase Storage or local storage depending on config
func saveUploadedFile(file multipart.File, originalFilename string) (string, error) {
	// Generate unique name
	filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), originalFilename)

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	if supabaseURL != "" && supabaseKey != "" {
		log.Println("SUPABASE_URL detected. Uploading file to Supabase Storage...")
		_, _ = file.Seek(0, io.SeekStart)
		return uploadToSupabaseStorage(file, filename)
	}

	EnsureUploadDirExists()
	destPath := filepath.Join(UploadDir, filename)
	out, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, _ = file.Seek(0, io.SeekStart)
	_, err = io.Copy(out, file)
	if err != nil {
		return "", err
	}
	return "/uploads/" + filename, nil
}

// ApplyHandler handles new application submission with file uploads
func ApplyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload size (5MB per file, max 75MB total)
	err := r.ParseMultipartForm(75 << 20)
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
	isDivorced := r.FormValue("is_divorced")

	if applicantName == "" || applicantNik == "" || applicantKk == "" || applicantEmail == "" || heirName == "" {
		http.Error(w, "Semua field utama wajib diisi", http.StatusBadRequest)
		return
	}

	filePaths := make(map[string]string)

	// 1. Global mandatory files
	globalMandatoryKeys := []string{
		"file_permohonan", "file_pengantar_rt_rw", "file_pernyataan_kebenaran", "file_sptjm",
		"file_ktp_pewaris", "file_ktp_ahli_waris", "file_kk_ahli_waris", "file_akta_lahir_ahli_waris",
		"file_ktp_saksi",
	}
	for _, key := range globalMandatoryKeys {
		file, handler, err := r.FormFile(key)
		if err != nil {
			http.Error(w, fmt.Sprintf("Berkas '%s' wajib diunggah", key), http.StatusBadRequest)
			return
		}
		defer file.Close()
		path, err := saveUploadedFile(file, handler.Filename)
		if err != nil {
			http.Error(w, "Gagal menulis berkas", http.StatusInternalServerError)
			return
		}
		filePaths[key] = path
	}

	// 2. Global optional files
	globalOptionalKeys := []string{"file_kematian_ahli_waris_wafat_lebih_dulu", "file_pendukung_lainnya"}
	for _, key := range globalOptionalKeys {
		file, handler, err := r.FormFile(key)
		if err != nil {
			filePaths[key] = ""
			continue
		}
		defer file.Close()
		path, err := saveUploadedFile(file, handler.Filename)
		if err != nil {
			http.Error(w, "Gagal menulis berkas", http.StatusInternalServerError)
			return
		}
		filePaths[key] = path
	}

	// 3. Relationship-based conditional files
	hasRelDocs := relationship == "Orang Tua" || relationship == "Istri / Suami"
	relKeys := []string{"file_surat_nikah_pewaris", "file_ktp_suami", "file_ktp_istri"}
	for _, key := range relKeys {
		if hasRelDocs {
			file, handler, err := r.FormFile(key)
			if err != nil {
				http.Error(w, fmt.Sprintf("Berkas '%s' wajib diunggah untuk hubungan %s", key, relationship), http.StatusBadRequest)
				return
			}
			defer file.Close()
			path, err := saveUploadedFile(file, handler.Filename)
			if err != nil {
				http.Error(w, "Gagal menulis berkas", http.StatusInternalServerError)
				return
			}
			filePaths[key] = path
		} else {
			filePaths[key] = ""
		}
	}

	// 4. Divorce-based conditional files
	needCeraiDoc := hasRelDocs && isDivorced == "Ya"
	if needCeraiDoc {
		file, handler, err := r.FormFile("file_akta_cerai_pewaris")
		if err != nil {
			http.Error(w, "Berkas 'file_akta_cerai_pewaris' wajib diunggah karena status cerai", http.StatusBadRequest)
			return
		}
		defer file.Close()
		path, err := saveUploadedFile(file, handler.Filename)
		if err != nil {
			http.Error(w, "Gagal menulis berkas", http.StatusInternalServerError)
			return
		}
		filePaths["file_akta_cerai_pewaris"] = path
	} else {
		filePaths["file_akta_cerai_pewaris"] = ""
	}

	// Start transaction
	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert application record based on active database engine
	var lastID int64
	if db.DbType == "postgres" {
		insertSQL := `
			INSERT INTO applications (
				registration_number, status, applicant_name, applicant_nik, applicant_kk, 
				applicant_address, applicant_phone, applicant_email, heir_name, death_date, 
				relationship, is_divorced, file_permohonan, file_pengantar_rt_rw, file_pernyataan_kebenaran,
				file_sptjm, file_ktp_pewaris, file_ktp_ahli_waris, file_kk_ahli_waris,
				file_akta_lahir_ahli_waris, file_ktp_saksi, file_kematian_ahli_waris_wafat_lebih_dulu,
				file_pendukung_lainnya, file_surat_nikah_pewaris, file_ktp_suami, file_ktp_istri,
				file_akta_cerai_pewaris, admin_notes, estimated_completion, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			RETURNING id
		`
		err = tx.QueryRow(db.Rebind(insertSQL), "TEMP", "Menunggu Verifikasi", applicantName, applicantNik, applicantKk,
			applicantAddress, applicantPhone, applicantEmail, heirName, deathDate, relationship, isDivorced,
			filePaths["file_permohonan"], filePaths["file_pengantar_rt_rw"], filePaths["file_pernyataan_kebenaran"],
			filePaths["file_sptjm"], filePaths["file_ktp_pewaris"], filePaths["file_ktp_ahli_waris"], filePaths["file_kk_ahli_waris"],
			filePaths["file_akta_lahir_ahli_waris"], filePaths["file_ktp_saksi"], filePaths["file_kematian_ahli_waris_wafat_lebih_dulu"],
			filePaths["file_pendukung_lainnya"], filePaths["file_surat_nikah_pewaris"], filePaths["file_ktp_suami"], filePaths["file_ktp_istri"],
			filePaths["file_akta_cerai_pewaris"], "", "", time.Now(), time.Now()).Scan(&lastID)
		if err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		insertSQL := `
			INSERT INTO applications (
				registration_number, status, applicant_name, applicant_nik, applicant_kk, 
				applicant_address, applicant_phone, applicant_email, heir_name, death_date, 
				relationship, is_divorced, file_permohonan, file_pengantar_rt_rw, file_pernyataan_kebenaran,
				file_sptjm, file_ktp_pewaris, file_ktp_ahli_waris, file_kk_ahli_waris,
				file_akta_lahir_ahli_waris, file_ktp_saksi, file_kematian_ahli_waris_wafat_lebih_dulu,
				file_pendukung_lainnya, file_surat_nikah_pewaris, file_ktp_suami, file_ktp_istri,
				file_akta_cerai_pewaris, admin_notes, estimated_completion, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		res, err := tx.Exec(db.Rebind(insertSQL), "TEMP", "Menunggu Verifikasi", applicantName, applicantNik, applicantKk,
			applicantAddress, applicantPhone, applicantEmail, heirName, deathDate, relationship, isDivorced,
			filePaths["file_permohonan"], filePaths["file_pengantar_rt_rw"], filePaths["file_pernyataan_kebenaran"],
			filePaths["file_sptjm"], filePaths["file_ktp_pewaris"], filePaths["file_ktp_ahli_waris"], filePaths["file_kk_ahli_waris"],
			filePaths["file_akta_lahir_ahli_waris"], filePaths["file_ktp_saksi"], filePaths["file_kematian_ahli_waris_wafat_lebih_dulu"],
			filePaths["file_pendukung_lainnya"], filePaths["file_surat_nikah_pewaris"], filePaths["file_ktp_suami"], filePaths["file_ktp_istri"],
			filePaths["file_akta_cerai_pewaris"], "", "", time.Now(), time.Now())
		if err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lastID, err = res.LastInsertId()
		if err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Format permanent registration number: SWG-[YEAR]-[4 DIGIT ID] (e.g. SWG-2026-0012)
	regNum := fmt.Sprintf("SWG-%d-%04d", time.Now().Year(), lastID)

	_, err = tx.Exec(db.Rebind("UPDATE applications SET registration_number = ? WHERE id = ?"), regNum, lastID)
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
