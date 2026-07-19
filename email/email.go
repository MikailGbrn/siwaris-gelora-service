package email

import (
	"log"
)

// SendSubmissionEmail mocks sending confirmation upon new application by logging it to the console
func SendSubmissionEmail(to string, name string, regNum string) {
	log.Printf("\n--- [MOCK EMAIL SENT] ---\nTo: %s\nSubject: [SIWARIS GELORA] Pendaftaran Berhasil - %s\nBody:\nHallo %s, permohonan Anda dengan nomor registrasi %s telah kami terima.\n-------------------------\n", to, regNum, name, regNum)
}

// SendStatusUpdateEmail mocks sending status update notification by logging it to the console
func SendStatusUpdateEmail(to string, name string, regNum string, status string, notes string) {
	log.Printf("\n--- [MOCK EMAIL SENT] ---\nTo: %s\nSubject: [SIWARIS GELORA] Pembaruan Status - %s\nBody:\nHalo %s, status permohonan %s diubah menjadi: %s. Catatan: %s\n-------------------------\n", to, regNum, name, regNum, status, notes)
}
