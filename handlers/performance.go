package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"task-manager-go/middleware"
	"task-manager-go/models"
)

type PerformancePageData struct {
	Performances []models.AppPerformance
	Clients      []models.Client
}

// ListPerformances renders the list of app performances
func ListPerformances(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetSessionUser(r)

	performances, err := models.GetAllPerformancesByUserID(userID)
	if err != nil {
		log.Printf("Error fetching app performances: %v", err)
		RenderTemplate(w, r, "performance.html", "Peforma App", "performance", nil, "Gagal memuat data peforma app.", "")
		return
	}

	clients, err := models.GetAllClients()
	if err != nil {
		log.Printf("Error fetching clients for performance: %v", err)
		clients = []models.Client{}
	}

	data := PerformancePageData{
		Performances: performances,
		Clients:      clients,
	}

	successMsg := r.URL.Query().Get("success")
	errorMsg := r.URL.Query().Get("error")

	RenderTemplate(w, r, "performance.html", "Peforma App", "performance", data, errorMsg, successMsg)
}

// CreatePerformance processes performance record creation
func CreatePerformance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/performance", http.StatusSeeOther)
		return
	}

	userID, _ := middleware.GetSessionUser(r)
	clientIDStr := strings.TrimSpace(r.FormValue("client_id"))
	bulan := strings.TrimSpace(r.FormValue("bulan")) // YYYY-MM
	totalKlienStr := strings.TrimSpace(r.FormValue("total_klien"))
	totalProjectStr := strings.TrimSpace(r.FormValue("total_project"))
	totalUserStr := strings.TrimSpace(r.FormValue("total_user"))
	totalUserAktifStr := strings.TrimSpace(r.FormValue("total_user_aktif"))
	totalAbsenStr := strings.TrimSpace(r.FormValue("total_absen"))
	totalTelatStr := strings.TrimSpace(r.FormValue("total_telat"))
	tepatWaktuStr := strings.TrimSpace(r.FormValue("tepat_waktu"))
	
	fiturAbsensi := r.FormValue("fitur_absensi") == "on"
	fiturLaporan := r.FormValue("fitur_laporan") == "on"
	fiturPayroll := r.FormValue("fitur_payroll") == "on"
	fiturMonitoring := r.FormValue("fitur_monitoring") == "on"
	fiturPayontime := r.FormValue("fitur_payontime") == "on"
	fiturPaynow := r.FormValue("fitur_paynow") == "on"
	catatan := strings.TrimSpace(r.FormValue("catatan"))

	if clientIDStr == "" || bulan == "" || totalUserStr == "" {
		http.Redirect(w, r, "/performance?error=Klien,+Bulan,+dan+Total+User+wajib+diisi", http.StatusSeeOther)
		return
	}

	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Redirect(w, r, "/performance?error=Klien+tidak+valid", http.StatusSeeOther)
		return
	}

	totalKlien, _ := strconv.Atoi(totalKlienStr)
	totalProject, _ := strconv.Atoi(totalProjectStr)
	totalUser, _ := strconv.Atoi(totalUserStr)
	totalUserAktif, _ := strconv.Atoi(totalUserAktifStr)
	totalAbsen, _ := strconv.Atoi(totalAbsenStr)
	totalTelat, _ := strconv.Atoi(totalTelatStr)
	tepatWaktu, _ := strconv.Atoi(tepatWaktuStr)

	err = models.CreatePerformance(userID, clientID, bulan, totalKlien, totalProject, totalUser, totalUserAktif, totalAbsen, totalTelat, tepatWaktu, fiturAbsensi, fiturLaporan, fiturPayroll, fiturMonitoring, fiturPayontime, fiturPaynow, catatan)
	if err != nil {
		log.Printf("Error creating performance record: %v", err)
		http.Redirect(w, r, "/performance?error=Gagal+menyimpan+data+peforma+app", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/performance?success=Data+peforma+app+berhasil+disimpan", http.StatusSeeOther)
}

// UpdatePerformance processes performance record updates
func UpdatePerformance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/performance", http.StatusSeeOther)
		return
	}

	userID, _ := middleware.GetSessionUser(r)
	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/performance?error=ID+tidak+valid", http.StatusSeeOther)
		return
	}

	clientIDStr := strings.TrimSpace(r.FormValue("client_id"))
	bulan := strings.TrimSpace(r.FormValue("bulan")) // YYYY-MM
	totalKlienStr := strings.TrimSpace(r.FormValue("total_klien"))
	totalProjectStr := strings.TrimSpace(r.FormValue("total_project"))
	totalUserStr := strings.TrimSpace(r.FormValue("total_user"))
	totalUserAktifStr := strings.TrimSpace(r.FormValue("total_user_aktif"))
	totalAbsenStr := strings.TrimSpace(r.FormValue("total_absen"))
	totalTelatStr := strings.TrimSpace(r.FormValue("total_telat"))
	tepatWaktuStr := strings.TrimSpace(r.FormValue("tepat_waktu"))
	
	fiturAbsensi := r.FormValue("fitur_absensi") == "on"
	fiturLaporan := r.FormValue("fitur_laporan") == "on"
	fiturPayroll := r.FormValue("fitur_payroll") == "on"
	fiturMonitoring := r.FormValue("fitur_monitoring") == "on"
	fiturPayontime := r.FormValue("fitur_payontime") == "on"
	fiturPaynow := r.FormValue("fitur_paynow") == "on"
	catatan := strings.TrimSpace(r.FormValue("catatan"))

	if clientIDStr == "" || bulan == "" || totalUserStr == "" {
		http.Redirect(w, r, "/performance?error=Klien,+Bulan,+dan+Total+User+wajib+diisi", http.StatusSeeOther)
		return
	}

	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Redirect(w, r, "/performance?error=Klien+tidak+valid", http.StatusSeeOther)
		return
	}

	totalKlien, _ := strconv.Atoi(totalKlienStr)
	totalProject, _ := strconv.Atoi(totalProjectStr)
	totalUser, _ := strconv.Atoi(totalUserStr)
	totalUserAktif, _ := strconv.Atoi(totalUserAktifStr)
	totalAbsen, _ := strconv.Atoi(totalAbsenStr)
	totalTelat, _ := strconv.Atoi(totalTelatStr)
	tepatWaktu, _ := strconv.Atoi(tepatWaktuStr)

	err = models.UpdatePerformance(id, userID, clientID, bulan, totalKlien, totalProject, totalUser, totalUserAktif, totalAbsen, totalTelat, tepatWaktu, fiturAbsensi, fiturLaporan, fiturPayroll, fiturMonitoring, fiturPayontime, fiturPaynow, catatan)
	if err != nil {
		log.Printf("Error updating performance record: %v", err)
		http.Redirect(w, r, "/performance?error=Gagal+memperbarui+data+peforma+app", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/performance?success=Data+peforma+app+berhasil+diperbarui", http.StatusSeeOther)
}

// DeletePerformance processes performance record deletion
func DeletePerformance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/performance", http.StatusSeeOther)
		return
	}

	userID, _ := middleware.GetSessionUser(r)
	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/performance?error=ID+tidak+valid", http.StatusSeeOther)
		return
	}

	err = models.DeletePerformance(id, userID)
	if err != nil {
		log.Printf("Error deleting performance record: %v", err)
		http.Redirect(w, r, "/performance?error=Gagal+menghapus+data+peforma+app", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/performance?success=Data+peforma+app+berhasil+dihapus", http.StatusSeeOther)
}
