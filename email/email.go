package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type ResendPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func sendEmail(to string, subject string, htmlContent string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		log.Println("WARNING: RESEND_API_KEY environment variable is not set. Falling back to mock logging.")
		return fmt.Errorf("RESEND_API_KEY not set")
	}

	payload := ResendPayload{
		From:    "SIWARIS Gelora <onboarding@resend.dev>",
		To:      []string{to},
		Subject: subject,
		HTML:    htmlContent,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("ERROR: Resend API request failed: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		errMsg := string(bodyBytes)
		log.Printf("ERROR: Resend API returned status %s: %s\n", resp.Status, errMsg)
		if resp.StatusCode == http.StatusBadRequest && strings.Contains(strings.ToLower(errMsg), "sandbox") {
			log.Printf("\n--- [RESEND SANDBOX LIMITATION] ---\nResend API rejected the email because you are in sandbox mode and trying to send to '%s'.\nTo fix this:\n1. Send emails to the address you used to register on Resend.\n2. Or verify a custom domain on the Resend dashboard.\nError details: %s\n-----------------------------------\n", to, errMsg)
		}
		return fmt.Errorf("failed to send email: status %s, response: %s", resp.Status, errMsg)
	}

	log.Printf("Email successfully sent via Resend API to: %s\n", to)
	return nil
}

// SendSubmissionEmail sends a submission confirmation email to the citizen
func SendSubmissionEmail(to string, name string, regNum string) {
	subject := fmt.Sprintf("[SIWARIS GELORA] Pendaftaran Berhasil - %s", regNum)

	htmlContent := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="utf-8">
		<title>Pendaftaran Berhasil</title>
		<style>
			body { font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; line-height: 1.6; color: #333333; margin: 0; padding: 0; background-color: #f4f7f6; }
			.container { max-width: 600px; margin: 20px auto; padding: 30px; background-color: #ffffff; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.05); }
			.header { text-align: center; border-bottom: 2px solid #e2e8f0; padding-bottom: 20px; margin-bottom: 25px; }
			.header h1 { color: #0284c7; margin: 0; font-size: 24px; }
			.content { font-size: 16px; }
			.registration-box { background-color: #f0f9ff; border: 1px solid #bae6fd; border-radius: 6px; padding: 20px; margin: 20px 0; text-align: center; }
			.registration-number { font-size: 24px; font-weight: bold; color: #0369a1; letter-spacing: 1px; margin: 10px 0; }
			.footer { text-align: center; font-size: 12px; color: #94a3b8; border-top: 1px solid #e2e8f0; padding-top: 20px; margin-top: 30px; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h1>SIWARIS GELORA</h1>
				<p style="margin: 5px 0 0 0; color: #64748b; font-size: 14px;">Kelurahan Gelora - Tanah Abang, Jakarta Pusat</p>
			</div>
			<div class="content">
				<p>Halo <strong>%s</strong>,</p>
				<p>Terima kasih telah menggunakan layanan SIWARIS Gelora. Permohonan Surat Pernyataan Ahli Waris Anda telah kami terima dengan sukses.</p>
				
				<div class="registration-box">
					<div style="font-size: 14px; color: #0369a1; text-transform: uppercase; font-weight: 500;">Nomor Registrasi Anda</div>
					<div class="registration-number">%s</div>
					<div style="font-size: 13px; color: #64748b;">Simpan nomor ini untuk melakukan pengecekan status permohonan Anda.</div>
				</div>

				<p><strong>Langkah Selanjutnya:</strong></p>
				<ol>
					<li>Petugas loket kami akan memverifikasi kelengkapan berkas fisik yang Anda unggah.</li>
					<li>Anda dapat memantau progres layanan ini di halaman <strong>Cek Status</strong> dengan menggunakan Nomor Registrasi di atas dan NIK Anda.</li>
					<li>Jika terdapat kekurangan atau kesalahan dokumen, status akan diubah menjadi <em>"Perlu Perbaikan"</em> dan Anda dapat mengunggah ulang dokumen revisi secara langsung melalui halaman <strong>Cek Status</strong>.</li>
				</ol>

				<p>Jika ada pertanyaan, silakan hubungi pusat pelayanan kami di Kantor Kelurahan Gelora.</p>
			</div>
			<div class="footer">
				<p>&copy; 2026 Pemerintah Provinsi DKI Jakarta | Kelurahan Gelora</p>
				<p>Email ini dikirimkan secara otomatis oleh sistem pelayanan digital SIWARIS GELORA.</p>
			</div>
		</div>
	</body>
	</html>
	`, name, regNum)

	err := sendEmail(to, subject, htmlContent)
	if err != nil {
		log.Printf("\n--- [FALLBACK MOCK EMAIL SENT] ---\nTo: %s\nSubject: %s\nBody:\nHallo %s, permohonan Anda dengan nomor registrasi %s telah kami terima.\n-------------------------\n", to, subject, name, regNum)
	}
}

// SendRevisionReceivedEmail sends an email to the citizen confirming their revision has been received
func SendRevisionReceivedEmail(to string, name string, regNum string) {
	subject := fmt.Sprintf("[SIWARIS GELORA] Dokumen Perbaikan Diterima - %s", regNum)

	htmlContent := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="utf-8">
		<title>Dokumen Perbaikan Diterima</title>
		<style>
			body { font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; line-height: 1.6; color: #333333; margin: 0; padding: 0; background-color: #f4f7f6; }
			.container { max-width: 600px; margin: 20px auto; padding: 30px; background-color: #ffffff; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.05); }
			.header { text-align: center; border-bottom: 2px solid #e2e8f0; padding-bottom: 20px; margin-bottom: 25px; }
			.header h1 { color: #0284c7; margin: 0; font-size: 24px; }
			.content { font-size: 16px; }
			.info-box { background-color: #f0fdf4; border: 1px solid #bbf7d0; border-radius: 6px; padding: 20px; margin: 20px 0; text-align: center; color: #166534; }
			.registration-number { font-size: 22px; font-weight: bold; letter-spacing: 1px; margin: 10px 0; }
			.footer { text-align: center; font-size: 12px; color: #94a3b8; border-top: 1px solid #e2e8f0; padding-top: 20px; margin-top: 30px; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h1>SIWARIS GELORA</h1>
				<p style="margin: 5px 0 0 0; color: #64748b; font-size: 14px;">Kelurahan Gelora - Tanah Abang, Jakarta Pusat</p>
			</div>
			<div class="content">
				<p>Halo <strong>%s</strong>,</p>
				<p>Dokumen perbaikan (revisi) untuk permohonan Surat Pernyataan Ahli Waris Anda telah berhasil kami terima.</p>
				
				<div class="info-box">
					<div style="font-size: 14px; text-transform: uppercase; font-weight: 500;">Nomor Registrasi</div>
					<div class="registration-number">%s</div>
					<div style="font-size: 13px;">Status permohonan Anda kini kembali menjadi <strong>"Menunggu Verifikasi"</strong>.</div>
				</div>

				<p>Petugas kami akan segera meninjau ulang dokumen perbaikan yang baru saja Anda unggah. Anda dapat terus memantau perkembangannya melalui halaman <strong>Cek Status</strong> pada portal SIWARIS Gelora.</p>
			</div>
			<div class="footer">
				<p>&copy; 2026 Pemerintah Provinsi DKI Jakarta | Kelurahan Gelora</p>
				<p>Email ini dikirimkan secara otomatis oleh sistem pelayanan digital SIWARIS GELORA.</p>
			</div>
		</div>
	</body>
	</html>
	`, name, regNum)

	err := sendEmail(to, subject, htmlContent)
	if err != nil {
		log.Printf("\n--- [FALLBACK MOCK EMAIL SENT] ---\nTo: %s\nSubject: %s\nBody:\nHalo %s, dokumen perbaikan untuk permohonan %s telah kami terima.\n-------------------------\n", to, subject, name, regNum)
	}
}

// SendStatusUpdateEmail sends status change notification email to the citizen
func SendStatusUpdateEmail(to string, name string, regNum string, status string, notes string) {
	subject := fmt.Sprintf("[SIWARIS GELORA] Status Baru: %s - %s", status, regNum)

	statusColor := "#64748b" // default gray
	statusDesc := "Permohonan Anda sedang ditinjau."

	switch status {
	case "Menunggu Verifikasi":
		statusColor = "#d97706" // Orange
		statusDesc = "Berkas perbaikan Anda sudah kami terima dan saat ini menunggu verifikasi ulang oleh petugas."
	case "Perlu Perbaikan":
		statusColor = "#ef4444" // Red
		statusDesc = "Petugas menemukan dokumen yang kurang sesuai. Mohon segera lakukan perbaikan berkas."
	case "Sedang Diproses":
		statusColor = "#1d4ed8" // Blue
		statusDesc = "Draf surat pernyataan ahli waris Anda sedang diproses dan disusun oleh kelurahan."
	case "Menunggu TTD":
		statusColor = "#7c3aed" // Purple
		statusDesc = "Draf surat selesai dibuat dan saat ini sedang menunggu tanda tangan dari Lurah."
	case "Selesai":
		statusColor = "#22c55e" // Green
		statusDesc = "Selamat! Surat Pernyataan Ahli Waris Anda telah selesai diterbitkan dan siap diambil."
	}

	htmlContent := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="utf-8">
		<title>Pembaruan Status Permohonan</title>
		<style>
			body { font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; line-height: 1.6; color: #333333; margin: 0; padding: 0; background-color: #f4f7f6; }
			.container { max-width: 600px; margin: 20px auto; padding: 30px; background-color: #ffffff; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.05); }
			.header { text-align: center; border-bottom: 2px solid #e2e8f0; padding-bottom: 20px; margin-bottom: 25px; }
			.header h1 { color: #0284c7; margin: 0; font-size: 24px; }
			.content { font-size: 16px; }
			.status-box { background-color: #fafafa; border: 1px solid #e2e8f0; border-radius: 6px; padding: 20px; margin: 20px 0; }
			.status-badge { display: inline-block; padding: 6px 16px; font-weight: bold; color: #ffffff; border-radius: 20px; font-size: 14px; text-transform: uppercase; margin-bottom: 10px; }
			.notes-box { background-color: #fffbeb; border-left: 4px solid #f59e0b; padding: 15px; border-radius: 4px; margin-top: 15px; font-style: italic; }
			.footer { text-align: center; font-size: 12px; color: #94a3b8; border-top: 1px solid #e2e8f0; padding-top: 20px; margin-top: 30px; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h1>SIWARIS GELORA</h1>
				<p style="margin: 5px 0 0 0; color: #64748b; font-size: 14px;">Kelurahan Gelora - Tanah Abang, Jakarta Pusat</p>
			</div>
			<div class="content">
				<p>Halo <strong>%s</strong>,</p>
				<p>Terdapat pembaruan status pada permohonan Surat Pernyataan Ahli Waris Anda untuk nomor registrasi <strong>%s</strong>.</p>
				
				<div class="status-box">
					<div class="status-badge" style="background-color: %s;">%s</div>
					<div style="font-weight: 500; font-size: 15px; margin-bottom: 10px;">%s</div>
					
					%s
				</div>

				<p>Silakan kunjungi halaman <strong>Cek Status</strong> pada portal SIWARIS Gelora untuk melacak riwayat pelayanan atau mengunggah berkas perbaikan jika diperlukan.</p>
			</div>
			<div class="footer">
				<p>&copy; 2026 Pemerintah Provinsi DKI Jakarta | Kelurahan Gelora</p>
				<p>Email ini dikirimkan secara otomatis oleh sistem pelayanan digital SIWARIS GELORA.</p>
				<p style="font-size: 9px; color: #cbd5e1; margin-top: 5px;">Ref ID: %d</p>
			</div>
		</div>
	</body>
	</html>
	`, name, regNum, statusColor, status, statusDesc, func() string {
		if notes != "" {
			return fmt.Sprintf(`<div class="notes-box"><strong>Catatan Petugas:</strong> "%s"</div>`, notes)
		}
		return ""
	}(), time.Now().UnixNano())

	err := sendEmail(to, subject, htmlContent)
	if err != nil {
		log.Printf("\n--- [FALLBACK MOCK EMAIL SENT] ---\nTo: %s\nSubject: %s\nBody:\nHalo %s, status permohonan %s diubah menjadi: %s. Catatan: %s\n-------------------------\n", to, subject, name, regNum, status, notes)
	}
}

// SendAdminNewSubmissionEmail notifies the admin when a new application is submitted
func SendAdminNewSubmissionEmail(to string, applicantName string, regNum string, relationship string) {
	subject := fmt.Sprintf("[SIWARIS GELORA] PERMOHONAN BARU MASUK - %s", regNum)

	panelURL := os.Getenv("ADMIN_PANEL_URL")
	if panelURL == "" {
		panelURL = "http://localhost:5173/login"
	}

	htmlContent := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="utf-8">
		<title>Permohonan Baru Masuk</title>
		<style>
			body { font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; line-height: 1.6; color: #333333; margin: 0; padding: 0; background-color: #f4f7f6; }
			.container { max-width: 600px; margin: 20px auto; padding: 30px; background-color: #ffffff; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.05); }
			.header { text-align: center; border-bottom: 2px solid #e2e8f0; padding-bottom: 20px; margin-bottom: 25px; }
			.header h1 { color: #1e3a8a; margin: 0; font-size: 24px; }
			.content { font-size: 16px; }
			.info-box { background-color: #f8fafc; border: 1px solid #e2e8f0; border-radius: 6px; padding: 20px; margin: 20px 0; }
			.btn-link { display: inline-block; padding: 12px 24px; color: #ffffff; background-color: #1e3a8a; text-decoration: none; border-radius: 6px; font-weight: bold; margin-top: 15px; text-align: center; }
			.footer { text-align: center; font-size: 12px; color: #94a3b8; border-top: 1px solid #e2e8f0; padding-top: 20px; margin-top: 30px; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h1>SIWARIS GELORA - NOTIFIKASI ADMIN</h1>
				<p style="margin: 5px 0 0 0; color: #64748b; font-size: 14px;">Kelurahan Gelora - Tanah Abang, Jakarta Pusat</p>
			</div>
			<div class="content">
				<p>Halo Petugas Administrator,</p>
				<p>Sebuah permohonan Surat Pernyataan Ahli Waris baru telah berhasil diajukan oleh pemohon dan membutuhkan verifikasi berkas dari pihak Kelurahan.</p>
				
				<div class="info-box">
					<table style="width: 100%; border-collapse: collapse; font-size: 15px;">
						<tr>
							<td style="padding: 6px 0; color: #64748b; width: 40%;">Nomor Registrasi:</td>
							<td style="padding: 6px 0; font-weight: bold; color: #1e3a8a;">%s</td>
						</tr>
						<tr>
							<td style="padding: 6px 0; color: #64748b;">Nama Pemohon:</td>
							<td style="padding: 6px 0; font-weight: bold;">%s</td>
						</tr>
						<tr>
							<td style="padding: 6px 0; color: #64748b;">Hubungan Pewaris:</td>
							<td style="padding: 6px 0;">%s</td>
						</tr>
						<tr>
							<td style="padding: 6px 0; color: #64748b;">Waktu Pengajuan:</td>
							<td style="padding: 6px 0; color: #64748b;">%s</td>
						</tr>
					</table>
				</div>

				<p>Mohon segera masuk ke Admin Dashboard untuk memeriksa kelengkapan dokumen pendukung yang diunggah pemohon.</p>
				
				<div style="text-align: center;">
					<a href="%s" class="btn-link" style="color: #ffffff;">Masuk ke Panel Admin</a>
				</div>
			</div>
			<div class="footer">
				<p>&copy; 2026 Pemerintah Provinsi DKI Jakarta | Kelurahan Gelora</p>
				<p>Email ini dikirimkan secara otomatis oleh sistem SIWARIS GELORA untuk Petugas Administrator.</p>
			</div>
		</div>
	</body>
	</html>
	`, regNum, applicantName, relationship, time.Now().Format("02-01-2006 15:04 MST"), panelURL)

	err := sendEmail(to, subject, htmlContent)
	if err != nil {
		log.Printf("\n--- [FALLBACK MOCK ADMIN EMAIL SENT] ---\nTo: %s\nSubject: %s\nBody:\nNotifikasi Baru: Permohonan baru %s oleh %s (%s) membutuhkan verifikasi.\n-------------------------\n", to, subject, regNum, applicantName, relationship)
	}
}

// SendAdminRevisionSubmittedEmail notifies the admin when a citizen submits revision files
func SendAdminRevisionSubmittedEmail(to string, applicantName string, regNum string) {
	subject := fmt.Sprintf("[SIWARIS GELORA] PERBAIKAN BERKAS DIUNGGAH - %s", regNum)

	panelURL := os.Getenv("ADMIN_PANEL_URL")
	if panelURL == "" {
		panelURL = "http://localhost:5173/login"
	}

	htmlContent := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="utf-8">
		<title>Perbaikan Berkas Diunggah</title>
		<style>
			body { font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; line-height: 1.6; color: #333333; margin: 0; padding: 0; background-color: #f4f7f6; }
			.container { max-width: 600px; margin: 20px auto; padding: 30px; background-color: #ffffff; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.05); }
			.header { text-align: center; border-bottom: 2px solid #e2e8f0; padding-bottom: 20px; margin-bottom: 25px; }
			.header h1 { color: #b91c1c; margin: 0; font-size: 24px; }
			.content { font-size: 16px; }
			.info-box { background-color: #fffbeb; border: 1px solid #fef3c7; border-radius: 6px; padding: 20px; margin: 20px 0; }
			.btn-link { display: inline-block; padding: 12px 24px; color: #ffffff; background-color: #b91c1c; text-decoration: none; border-radius: 6px; font-weight: bold; margin-top: 15px; text-align: center; }
			.footer { text-align: center; font-size: 12px; color: #94a3b8; border-top: 1px solid #e2e8f0; padding-top: 20px; margin-top: 30px; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h1>SIWARIS GELORA - NOTIFIKASI ADMIN</h1>
				<p style="margin: 5px 0 0 0; color: #64748b; font-size: 14px;">Kelurahan Gelora - Tanah Abang, Jakarta Pusat</p>
			</div>
			<div class="content">
				<p>Halo Petugas Administrator,</p>
				<p>Pemohon telah mengunggah berkas perbaikan (revisi) untuk permohonan yang sebelumnya ditandai <em>"Perlu Perbaikan"</em>.</p>
				
				<div class="info-box">
					<table style="width: 100%; border-collapse: collapse; font-size: 15px;">
						<tr>
							<td style="padding: 6px 0; color: #64748b; width: 40%;">Nomor Registrasi:</td>
							<td style="padding: 6px 0; font-weight: bold; color: #b91c1c;">%s</td>
						</tr>
						<tr>
							<td style="padding: 6px 0; color: #64748b;">Nama Pemohon:</td>
							<td style="padding: 6px 0; font-weight: bold;">%s</td>
						</tr>
						<tr>
							<td style="padding: 6px 0; color: #64748b;">Waktu Unggah:</td>
							<td style="padding: 6px 0; color: #64748b;">%s</td>
						</tr>
					</table>
				</div>

				<p>Mohon segera masuk ke Admin Dashboard untuk memeriksa kembali berkas perbaikan tersebut dan melanjutkan proses pelayanan.</p>
				
				<div style="text-align: center;">
					<a href="%s" class="btn-link" style="color: #ffffff;">Masuk ke Panel Admin</a>
				</div>
			</div>
			<div class="footer">
				<p>&copy; 2026 Pemerintah Provinsi DKI Jakarta | Kelurahan Gelora</p>
				<p>Email ini dikirimkan secara otomatis oleh sistem SIWARIS GELORA untuk Petugas Administrator.</p>
			</div>
		</div>
	</body>
	</html>
	`, regNum, applicantName, time.Now().Format("02-01-2006 15:04 MST"), panelURL)

	err := sendEmail(to, subject, htmlContent)
	if err != nil {
		log.Printf("\n--- [FALLBACK MOCK ADMIN EMAIL SENT] ---\nTo: %s\nSubject: %s\nBody:\nNotifikasi Baru: Perbaikan berkas diunggah untuk %s oleh %s.\n-------------------------\n", to, subject, regNum, applicantName)
	}
}
