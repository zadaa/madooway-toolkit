package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"task-manager-go/db"
)

type AppPerformance struct {
	ID              int
	UserID          int
	ClientID        int
	ClientName      string
	ClientLogo      string
	Bulan           string // YYYY-MM
	TotalKlien      int
	TotalProject    int
	TotalUser       int
	TotalUserAktif  int
	TotalAbsen      int
	TotalTelat      int
	TepatWaktu      int
	FiturAbsensi    bool
	FiturLaporan    bool
	FiturPayroll    bool
	FiturMonitoring bool
	FiturPayontime  bool
	FiturPaynow     bool
	Catatan         string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// EngagementRate calculates (total_user_aktif/total_user * 100)
func (a *AppPerformance) EngagementRate() float64 {
	if a.TotalUser <= 0 {
		return 0.0
	}
	return (float64(a.TotalUserAktif) / float64(a.TotalUser)) * 100.0
}

// LateRate calculates (total_telat/total_absen)
func (a *AppPerformance) LateRate() float64 {
	if a.TotalAbsen <= 0 {
		return 0.0
	}
	return float64(a.TotalTelat) / float64(a.TotalAbsen)
}

// FormattedEngagementRate formats EngagementRate with percentage sign
func (a *AppPerformance) FormattedEngagementRate() string {
	return fmt.Sprintf("%.2f%%", a.EngagementRate())
}

// FormattedLateRate formats LateRate as a percentage or ratio. We show as percentage.
func (a *AppPerformance) FormattedLateRate() string {
	return fmt.Sprintf("%.2f%%", a.LateRate()*100.0)
}

// FormattedBulan formats YYYY-MM to Indonesian Month Name and Year (e.g. Mei 2026)
func (a *AppPerformance) FormattedBulan() string {
	t, err := time.Parse("2006-01", a.Bulan)
	if err != nil {
		return a.Bulan
	}
	months := []string{
		"", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	monthIdx := int(t.Month())
	if monthIdx < 1 || monthIdx > 12 {
		return a.Bulan
	}
	return fmt.Sprintf("%s %d", months[monthIdx], t.Year())
}

// CreatePerformance inserts a new performance record
func CreatePerformance(userID, clientID int, bulan string, totalKlien, totalProject, totalUser, totalUserAktif, totalAbsen, totalTelat, tepatWaktu int, fiturAbsensi, fiturLaporan, fiturPayroll, fiturMonitoring, fiturPayontime, fiturPaynow bool, catatan string) error {
	query := `INSERT INTO app_performances (user_id, client_id, bulan, total_klien, total_project, total_user, total_user_aktif, total_absen, total_telat, tepat_waktu, fitur_absensi, fitur_laporan, fitur_payroll, fitur_monitoring, fitur_payontime, fitur_paynow, catatan)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.DB.Exec(query, userID, clientID, bulan, totalKlien, totalProject, totalUser, totalUserAktif, totalAbsen, totalTelat, tepatWaktu, fiturAbsensi, fiturLaporan, fiturPayroll, fiturMonitoring, fiturPayontime, fiturPaynow, catatan)
	return err
}

// UpdatePerformance updates an existing performance record
func UpdatePerformance(id, userID, clientID int, bulan string, totalKlien, totalProject, totalUser, totalUserAktif, totalAbsen, totalTelat, tepatWaktu int, fiturAbsensi, fiturLaporan, fiturPayroll, fiturMonitoring, fiturPayontime, fiturPaynow bool, catatan string) error {
	query := `UPDATE app_performances 
	          SET client_id = ?, bulan = ?, total_klien = ?, total_project = ?, total_user = ?, total_user_aktif = ?, total_absen = ?, total_telat = ?, tepat_waktu = ?, fitur_absensi = ?, fitur_laporan = ?, fitur_payroll = ?, fitur_monitoring = ?, fitur_payontime = ?, fitur_paynow = ?, catatan = ?
	          WHERE id = ? AND user_id = ?`
	_, err := db.DB.Exec(query, clientID, bulan, totalKlien, totalProject, totalUser, totalUserAktif, totalAbsen, totalTelat, tepatWaktu, fiturAbsensi, fiturLaporan, fiturPayroll, fiturMonitoring, fiturPayontime, fiturPaynow, catatan, id, userID)
	return err
}

// DeletePerformance deletes a performance record
func DeletePerformance(id, userID int) error {
	query := `DELETE FROM app_performances WHERE id = ? AND user_id = ?`
	_, err := db.DB.Exec(query, id, userID)
	return err
}

// GetPerformanceByID fetches a single performance record by ID and user ID
func GetPerformanceByID(id, userID int) (*AppPerformance, error) {
	query := `SELECT ap.id, ap.user_id, ap.client_id, c.name as client_name, COALESCE(c.logo, '') as client_logo, ap.bulan, ap.total_klien, ap.total_project, ap.total_user, ap.total_user_aktif, ap.total_absen, ap.total_telat, ap.tepat_waktu, ap.fitur_absensi, ap.fitur_laporan, ap.fitur_payroll, ap.fitur_monitoring, ap.fitur_payontime, ap.fitur_paynow, COALESCE(ap.catatan, ''), ap.created_at, ap.updated_at 
	          FROM app_performances ap
	          INNER JOIN clients c ON ap.client_id = c.id
	          WHERE ap.id = ? AND ap.user_id = ?`
	row := db.DB.QueryRow(query, id, userID)

	var ap AppPerformance
	err := row.Scan(&ap.ID, &ap.UserID, &ap.ClientID, &ap.ClientName, &ap.ClientLogo, &ap.Bulan, &ap.TotalKlien, &ap.TotalProject, &ap.TotalUser, &ap.TotalUserAktif, &ap.TotalAbsen, &ap.TotalTelat, &ap.TepatWaktu, &ap.FiturAbsensi, &ap.FiturLaporan, &ap.FiturPayroll, &ap.FiturMonitoring, &ap.FiturPayontime, &ap.FiturPaynow, &ap.Catatan, &ap.CreatedAt, &ap.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &ap, nil
}

// GetAllPerformancesByUserID lists all performance records for a user
func GetAllPerformancesByUserID(userID int) ([]AppPerformance, error) {
	query := `SELECT ap.id, ap.user_id, ap.client_id, c.name as client_name, COALESCE(c.logo, '') as client_logo, ap.bulan, ap.total_klien, ap.total_project, ap.total_user, ap.total_user_aktif, ap.total_absen, ap.total_telat, ap.tepat_waktu, ap.fitur_absensi, ap.fitur_laporan, ap.fitur_payroll, ap.fitur_monitoring, ap.fitur_payontime, ap.fitur_paynow, COALESCE(ap.catatan, ''), ap.created_at, ap.updated_at 
	          FROM app_performances ap
	          INNER JOIN clients c ON ap.client_id = c.id
	          WHERE ap.user_id = ?
	          ORDER BY ap.bulan DESC, c.name ASC`
	rows, err := db.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []AppPerformance
	for rows.Next() {
		var ap AppPerformance
		err := rows.Scan(&ap.ID, &ap.UserID, &ap.ClientID, &ap.ClientName, &ap.ClientLogo, &ap.Bulan, &ap.TotalKlien, &ap.TotalProject, &ap.TotalUser, &ap.TotalUserAktif, &ap.TotalAbsen, &ap.TotalTelat, &ap.TepatWaktu, &ap.FiturAbsensi, &ap.FiturLaporan, &ap.FiturPayroll, &ap.FiturMonitoring, &ap.FiturPayontime, &ap.FiturPaynow, &ap.Catatan, &ap.CreatedAt, &ap.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, ap)
	}
	return list, nil
}
