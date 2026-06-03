package handlers

import (
	"html/template"
	"net/http"
	"task-manager-go/db"
	"task-manager-go/middleware"
	"task-manager-go/models"
)

type PageData struct {
	Title      string
	ActiveTab  string
	User       *models.User
	ErrorMsg   string
	SuccessMsg string
	Data       interface{}
}

// RenderTemplate renders the main layout enclosing the specified template file
func RenderTemplate(w http.ResponseWriter, r *http.Request, templateName string, title string, activeTab string, data interface{}, errorMsg string, successMsg string) {
	var user *models.User
	userID, loggedIn := middleware.GetSessionUser(r)
	if loggedIn {
		user = &models.User{ID: userID}
		err := db.DB.QueryRow("SELECT username, email, role FROM users WHERE id = ?", userID).Scan(&user.Username, &user.Email, &user.Role)
		if err != nil {
			user.Username = "User"
		}
	}

	files := []string{
		"templates/layout.html",
		"templates/" + templateName,
	}

	tmpl := template.New("layout").Funcs(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"firstLetter": func(s string) string {
			if len(s) == 0 {
				return ""
			}
			return string([]rune(s)[0])
		},
	})
	
	tmpl, err := tmpl.ParseFiles(files...)
	if err != nil {
		http.Error(w, "Error parsing templates: "+err.Error(), http.StatusInternalServerError)
		return
	}

	pd := PageData{
		Title:      title,
		ActiveTab:  activeTab,
		User:       user,
		ErrorMsg:   errorMsg,
		SuccessMsg: successMsg,
		Data:       data,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tmpl.ExecuteTemplate(w, "layout", pd)
	if err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
	}
}
