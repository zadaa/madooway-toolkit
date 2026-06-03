package handlers

import (
	"log"
	"net/http"
	"task-manager-go/models"
)

// ShowLeaderboard renders the leaderboard page (Admin only)
func ShowLeaderboard(w http.ResponseWriter, r *http.Request) {
	entries, err := models.GetLeaderboard()
	if err != nil {
		log.Printf("Error fetching leaderboard: %v", err)
		RenderTemplate(w, r, "leaderboard.html", "Leaderboard Petugas", "leaderboard", nil, "Gagal memuat leaderboard.", "")
		return
	}

	RenderTemplate(w, r, "leaderboard.html", "Leaderboard Petugas", "leaderboard", entries, "", "")
}
