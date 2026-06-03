package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"task-manager-go/config"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDB() {
	cfg := config.AppConfig

	var dsnWithoutDB, dsn string
	if strings.HasPrefix(cfg.DBHost, "/") {
		// Unix socket connection (useful for Google Cloud SQL on Cloud Run)
		dsnWithoutDB = fmt.Sprintf("%s:%s@unix(%s)/?parseTime=true",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost)
		dsn = fmt.Sprintf("%s:%s@unix(%s)/%s?parseTime=true",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBName)
	} else {
		// Standard TCP connection (local and VM development)
		dsnWithoutDB = fmt.Sprintf("%s:%s@tcp(%s:%s)/?parseTime=true",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort)
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	}

	// 1. Establish connection to MySQL server without database to create it if it doesn't exist
	tempDB, err := sql.Open("mysql", dsnWithoutDB)
	if err != nil {
		log.Printf("Warning: Error opening connection to MySQL server: %v (continuing...)\n", err)
	} else {
		defer tempDB.Close()
		// Verify MySQL server is reachable
		err = tempDB.Ping()
		if err != nil {
			log.Printf("Warning: Error pinging MySQL server: %v (continuing...)\n", err)
		} else {
			// Create database if it doesn't exist
			_, err = tempDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", cfg.DBName))
			if err != nil {
				log.Printf("Warning: Error creating database %s: %v (continuing...)\n", cfg.DBName, err)
			} else {
				log.Printf("Database '%s' verified/created successfully\n", cfg.DBName)
			}
		}
	}

	// 2. Establish connection to the target database
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("Warning: Error opening connection to database %s: %v (continuing...)\n", cfg.DBName, err)
		return
	}

	// Connection pool settings
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	err = DB.Ping()
	if err != nil {
		log.Printf("Warning: Error pinging database %s: %v. Database might not be ready yet. Skipping migrations.\n", cfg.DBName, err)
		return
	}

	log.Println("Database connection pool established successfully")

	// 3. Run auto-migrations
	runMigrations()
}

func runMigrations() {
	// Create Users table
	usersTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		color VARCHAR(20) NOT NULL DEFAULT '#4f46e5',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB;`

	_, err := DB.Exec(usersTableQuery)
	if err != nil {
		log.Fatalf("Error creating users table: %v", err)
	}
	log.Println("Table 'users' verified/created")

	// Create Clients table
	clientsTableQuery := `
	CREATE TABLE IF NOT EXISTS clients (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(150) NOT NULL,
		short_name VARCHAR(50),
		email VARCHAR(100),
		phone VARCHAR(20),
		pic_name VARCHAR(150),
		price_package VARCHAR(100) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB;`

	_, err = DB.Exec(clientsTableQuery)
	if err != nil {
		log.Fatalf("Error creating clients table: %v", err)
	}
	log.Println("Table 'clients' verified/created")

	// Create Tasks table (includes source field as requested)
	tasksTableQuery := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id INT NOT NULL,
		title VARCHAR(255) NOT NULL,
		description TEXT,
		category VARCHAR(50) NOT NULL,
		source VARCHAR(100) NOT NULL DEFAULT 'WA Supp',
		status VARCHAR(20) NOT NULL DEFAULT 'Pending',
		due_date DATE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	) ENGINE=InnoDB;`

	_, err = DB.Exec(tasksTableQuery)
	if err != nil {
		log.Fatalf("Error creating tasks table: %v", err)
	}
	log.Println("Table 'tasks' verified/created")

	// Set default source value to 'WA Supp' for existing table
	_, err = DB.Exec("ALTER TABLE tasks ALTER COLUMN source SET DEFAULT 'WA Supp'")
	if err != nil {
		log.Printf("Warning: failed to alter source column default value: %v", err)
	}

	// Check if client_id column exists in tasks table
	var columnCount int
	checkQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
	               WHERE TABLE_SCHEMA = DATABASE() 
	               AND TABLE_NAME = 'tasks' 
	               AND COLUMN_NAME = 'client_id'`
	err = DB.QueryRow(checkQuery).Scan(&columnCount)
	if err == nil && columnCount == 0 {
		log.Println("Migrating tasks table: adding client_id column")
		_, err = DB.Exec("ALTER TABLE tasks ADD COLUMN client_id INT NULL")
		if err != nil {
			log.Fatalf("Error adding client_id column to tasks table: %v", err)
		}
		_, err = DB.Exec("ALTER TABLE tasks ADD CONSTRAINT fk_tasks_client_id FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE SET NULL")
		if err != nil {
			log.Fatalf("Error adding foreign key constraint to tasks table: %v", err)
		}
	}

	// Check if province column exists in clients table
	var provinceColumnCount int
	checkProvinceQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
	                      WHERE TABLE_SCHEMA = DATABASE() 
	                      AND TABLE_NAME = 'clients' 
	                      AND COLUMN_NAME = 'province'`
	err = DB.QueryRow(checkProvinceQuery).Scan(&provinceColumnCount)
	if err == nil && provinceColumnCount == 0 {
		log.Println("Migrating clients table: adding province column")
		_, err = DB.Exec("ALTER TABLE clients ADD COLUMN province VARCHAR(100) NOT NULL DEFAULT 'DKI Jakarta'")
		if err != nil {
			log.Fatalf("Error adding province column to clients table: %v", err)
		}
		log.Println("Database migration completed: province column added to clients table")
	}

	// Create Training Schedules table (includes trainer field)
	trainingSchedulesTableQuery := `
	CREATE TABLE IF NOT EXISTS training_schedules (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id INT NOT NULL,
		client_id INT NOT NULL,
		title VARCHAR(255) NOT NULL,
		description TEXT,
		training_date DATETIME NOT NULL,
		location VARCHAR(255),
		trainer VARCHAR(100),
		training_type VARCHAR(50) NOT NULL DEFAULT 'Online',
		status VARCHAR(50) NOT NULL DEFAULT 'Scheduled',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE
	) ENGINE=InnoDB;`

	_, err = DB.Exec(trainingSchedulesTableQuery)
	if err != nil {
		log.Fatalf("Error creating training_schedules table: %v", err)
	}
	log.Println("Table 'training_schedules' verified/created")

	// Check if trainer column exists in training_schedules table
	var trainerColumnExists int
	checkTrainerQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
	                      WHERE TABLE_SCHEMA = DATABASE() 
	                      AND TABLE_NAME = 'training_schedules' 
	                      AND COLUMN_NAME = 'trainer'`
	err = DB.QueryRow(checkTrainerQuery).Scan(&trainerColumnExists)
	if err == nil && trainerColumnExists == 0 {
		log.Println("Migrating training_schedules table: adding trainer column")
		_, err = DB.Exec("ALTER TABLE training_schedules ADD COLUMN trainer VARCHAR(100) NULL")
		if err != nil {
			log.Fatalf("Error adding trainer column to training_schedules table: %v", err)
		}
		log.Println("Database migration completed: trainer column added to training_schedules table")
	}

	// Check if training_type column exists in training_schedules table
	var trainingTypeColumnExists int
	checkTrainingTypeQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
	                           WHERE TABLE_SCHEMA = DATABASE() 
	                           AND TABLE_NAME = 'training_schedules' 
	                           AND COLUMN_NAME = 'training_type'`
	err = DB.QueryRow(checkTrainingTypeQuery).Scan(&trainingTypeColumnExists)
	if err == nil && trainingTypeColumnExists == 0 {
		log.Println("Migrating training_schedules table: adding training_type column")
		_, err = DB.Exec("ALTER TABLE training_schedules ADD COLUMN training_type VARCHAR(50) NOT NULL DEFAULT 'Online'")
		if err != nil {
			log.Fatalf("Error adding training_type column to training_schedules table: %v", err)
		}
		log.Println("Database migration completed: training_type column added to training_schedules table")
	}

	// Check if user_id column exists in clients table to migrate it
	var clientUserIdCount int
	checkClientUserIdQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
	                          WHERE TABLE_SCHEMA = DATABASE() 
	                          AND TABLE_NAME = 'clients' 
	                          AND COLUMN_NAME = 'user_id'`
	err = DB.QueryRow(checkClientUserIdQuery).Scan(&clientUserIdCount)
	if err == nil && clientUserIdCount > 0 {
		log.Println("Migrating clients table: dropping user_id column and foreign key")
		
		// Find foreign key constraint names
		findFkQuery := `SELECT CONSTRAINT_NAME FROM information_schema.TABLE_CONSTRAINTS
		                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'clients' AND CONSTRAINT_TYPE = 'FOREIGN KEY'`
		rows, err := DB.Query(findFkQuery)
		if err == nil {
			var fkNames []string
			for rows.Next() {
				var fkName string
				if err := rows.Scan(&fkName); err == nil {
					fkNames = append(fkNames, fkName)
				}
			}
			rows.Close()
			for _, fkName := range fkNames {
				_, err = DB.Exec(fmt.Sprintf("ALTER TABLE clients DROP FOREIGN KEY `%s`", fkName))
				if err != nil {
					log.Printf("Warning: failed to drop foreign key constraint %s: %v", fkName, err)
				}
			}
		}
		
		// Drop the user_id column
		_, err = DB.Exec("ALTER TABLE clients DROP COLUMN user_id")
		if err != nil {
			log.Fatalf("Error dropping user_id column from clients table: %v", err)
		}
		log.Println("Database migration completed: user_id column removed from clients table")
	}

	// Check if pic_name column exists in clients table
	var picNameCount int
	checkPicNameQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
	                      WHERE TABLE_SCHEMA = DATABASE() 
	                      AND TABLE_NAME = 'clients' 
	                      AND COLUMN_NAME = 'pic_name'`
	err = DB.QueryRow(checkPicNameQuery).Scan(&picNameCount)
	if err == nil && picNameCount == 0 {
		log.Println("Migrating clients table: adding pic_name column")
		_, err = DB.Exec("ALTER TABLE clients ADD COLUMN pic_name VARCHAR(150) NULL")
		if err != nil {
			log.Fatalf("Error adding pic_name column to clients table: %v", err)
		}
		log.Println("Database migration completed: pic_name column added to clients table")
	}

	// Check if short_name column exists in clients table
	var shortNameCount int
	checkShortNameQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
	                        WHERE TABLE_SCHEMA = DATABASE() 
	                        AND TABLE_NAME = 'clients' 
	                        AND COLUMN_NAME = 'short_name'`
	err = DB.QueryRow(checkShortNameQuery).Scan(&shortNameCount)
	if err == nil && shortNameCount == 0 {
		log.Println("Migrating clients table: adding short_name column")
		_, err = DB.Exec("ALTER TABLE clients ADD COLUMN short_name VARCHAR(50) NULL")
		if err != nil {
			log.Fatalf("Error adding short_name column to clients table: %v", err)
		}
		log.Println("Database migration completed: short_name column added to clients table")
	}

	// Check if logo column exists in clients table
	var logoCount int
	checkLogoQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
	                   WHERE TABLE_SCHEMA = DATABASE() 
	                   AND TABLE_NAME = 'clients' 
	                   AND COLUMN_NAME = 'logo'`
	err = DB.QueryRow(checkLogoQuery).Scan(&logoCount)
	if err == nil && logoCount == 0 {
		log.Println("Migrating clients table: adding logo column")
		_, err = DB.Exec("ALTER TABLE clients ADD COLUMN logo VARCHAR(255) NULL")
		if err != nil {
			log.Fatalf("Error adding logo column to clients table: %v", err)
		}
		log.Println("Database migration completed: logo column added to clients table")
	}

	// Ensure uploads directory exists for logos
	err = os.MkdirAll("static/uploads/logos", 0755)
	if err != nil {
		log.Fatalf("Error creating logo uploads directory: %v", err)
	}
	log.Println("Directory 'static/uploads/logos' verified/created")

	// Create app_performances table
	appPerformancesTableQuery := `
	CREATE TABLE IF NOT EXISTS app_performances (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id INT NOT NULL,
		client_id INT NOT NULL,
		bulan VARCHAR(7) NOT NULL,
		total_klien INT NOT NULL,
		total_project INT NOT NULL,
		total_user INT NOT NULL,
		total_user_aktif INT NOT NULL,
		total_absen INT NOT NULL,
		total_telat INT NOT NULL,
		tepat_waktu INT NOT NULL,
		fitur_absensi TINYINT(1) NOT NULL DEFAULT 0,
		fitur_laporan TINYINT(1) NOT NULL DEFAULT 0,
		fitur_payroll TINYINT(1) NOT NULL DEFAULT 0,
		fitur_monitoring TINYINT(1) NOT NULL DEFAULT 0,
		fitur_payontime TINYINT(1) NOT NULL DEFAULT 0,
		fitur_paynow TINYINT(1) NOT NULL DEFAULT 0,
		catatan TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE
	) ENGINE=InnoDB;`

	_, err = DB.Exec(appPerformancesTableQuery)
	if err != nil {
		log.Fatalf("Error creating app_performances table: %v", err)
	}
	log.Println("Table 'app_performances' verified/created")

	// Create Tickets table
	ticketsTableQuery := `
	CREATE TABLE IF NOT EXISTS tickets (
		id INT AUTO_INCREMENT PRIMARY KEY,
		client_id INT NOT NULL,
		title VARCHAR(255) NOT NULL,
		description TEXT,
		user_id INT NOT NULL,
		file_path VARCHAR(255) NULL,
		issue_date DATE NOT NULL,
		category VARCHAR(50) NOT NULL,
		ticket_link VARCHAR(255) NULL,
		status VARCHAR(50) NOT NULL DEFAULT 'Pending',
		finished_date DATE NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	) ENGINE=InnoDB;`

	_, err = DB.Exec(ticketsTableQuery)
	if err != nil {
		log.Fatalf("Error creating tickets table: %v", err)
	}
	log.Println("Table 'tickets' verified/created")

	// Ensure uploads directory exists for tickets
	err = os.MkdirAll("static/uploads/tickets", 0755)
	if err != nil {
		log.Fatalf("Error creating ticket uploads directory: %v", err)
	}
	log.Println("Directory 'static/uploads/tickets' verified/created")

	// Create Ticket Messages table for discussions
	ticketMessagesTableQuery := `
	CREATE TABLE IF NOT EXISTS ticket_messages (
		id INT AUTO_INCREMENT PRIMARY KEY,
		ticket_id INT NOT NULL,
		user_id INT NOT NULL,
		message TEXT,
		file_path VARCHAR(255) NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (ticket_id) REFERENCES tickets(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	) ENGINE=InnoDB;`

	_, err = DB.Exec(ticketMessagesTableQuery)
	if err != nil {
		log.Fatalf("Error creating ticket_messages table: %v", err)
	}
	log.Println("Table 'ticket_messages' verified/created")

	// Create User Sessions table
	userSessionsTableQuery := `
	CREATE TABLE IF NOT EXISTS user_sessions (
		token VARCHAR(255) PRIMARY KEY,
		user_id INT NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	) ENGINE=InnoDB;`

	_, err = DB.Exec(userSessionsTableQuery)
	if err != nil {
		log.Fatalf("Error creating user_sessions table: %v", err)
	}
	log.Println("Table 'user_sessions' verified/created")

	// Check if color column exists in users table
	var userColorExists int
	checkUserColorQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
	                        WHERE TABLE_SCHEMA = DATABASE() 
	                        AND TABLE_NAME = 'users' 
	                        AND COLUMN_NAME = 'color'`
	err = DB.QueryRow(checkUserColorQuery).Scan(&userColorExists)
	if err == nil && userColorExists == 0 {
		log.Println("Migrating users table: adding color column")
		_, err = DB.Exec("ALTER TABLE users ADD COLUMN color VARCHAR(20) NOT NULL DEFAULT '#4f46e5'")
		if err != nil {
			log.Fatalf("Error adding color column to users table: %v", err)
		}
		log.Println("Database migration completed: color column added to users table")
	}

	// Create Ticket Assignees table
	ticketAssigneesTableQuery := `
	CREATE TABLE IF NOT EXISTS ticket_assignees (
		ticket_id INT NOT NULL,
		user_id INT NOT NULL,
		PRIMARY KEY (ticket_id, user_id),
		FOREIGN KEY (ticket_id) REFERENCES tickets(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	) ENGINE=InnoDB;`

	_, err = DB.Exec(ticketAssigneesTableQuery)
	if err != nil {
		log.Fatalf("Error creating ticket_assignees table: %v", err)
	}
	log.Println("Table 'ticket_assignees' verified/created")

	// Create Leads table
	leadsTableQuery := `
	CREATE TABLE IF NOT EXISTS leads (
		id INT AUTO_INCREMENT PRIMARY KEY,
		source VARCHAR(50) NOT NULL,
		company_name VARCHAR(150) NOT NULL,
		contact_name VARCHAR(150) NOT NULL,
		phone VARCHAR(20),
		email VARCHAR(100),
		employee_count INT DEFAULT 0,
		status VARCHAR(50) NOT NULL DEFAULT 'Reachout',
		sales_id INT NOT NULL,
		follow_up_history TEXT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (sales_id) REFERENCES users(id) ON DELETE CASCADE
	) ENGINE=InnoDB;`

	_, err = DB.Exec(leadsTableQuery)
	if err != nil {
		log.Fatalf("Error creating leads table: %v", err)
	}
	log.Println("Table 'leads' verified/created")

	// Check if follow_up_history column exists in leads table
	var followUpHistoryExists int
	checkFollowUpQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
	                       WHERE TABLE_SCHEMA = DATABASE() 
	                       AND TABLE_NAME = 'leads' 
	                       AND COLUMN_NAME = 'follow_up_history'`
	err = DB.QueryRow(checkFollowUpQuery).Scan(&followUpHistoryExists)
	if err == nil && followUpHistoryExists == 0 {
		log.Println("Migrating leads table: adding follow_up_history column")
		_, err = DB.Exec("ALTER TABLE leads ADD COLUMN follow_up_history TEXT NULL")
		if err != nil {
			log.Fatalf("Error adding follow_up_history column to leads table: %v", err)
		}
		log.Println("Database migration completed: follow_up_history column added to leads table")
	}

	// Check if role column exists in users table
	var roleColumnExists int
	checkRoleQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
	                   WHERE TABLE_SCHEMA = DATABASE() 
	                   AND TABLE_NAME = 'users' 
	                   AND COLUMN_NAME = 'role'`
	err = DB.QueryRow(checkRoleQuery).Scan(&roleColumnExists)
	if err == nil && roleColumnExists == 0 {
		log.Println("Migrating users table: adding role column")
		_, err = DB.Exec("ALTER TABLE users ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'user'")
		if err != nil {
			log.Fatalf("Error adding role column to users table: %v", err)
		}
		// Set all existing users to 'admin' so they don't get locked out initially
		_, err = DB.Exec("UPDATE users SET role = 'admin'")
		if err != nil {
			log.Printf("Warning: failed to set existing users to admin: %v", err)
		}
		log.Println("Database migration completed: role column added to users table and existing users set to admin")
	}

	// Check if nip column exists in users table
	var nipColumnExists int
	checkNipQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
	                  WHERE TABLE_SCHEMA = DATABASE() 
	                  AND TABLE_NAME = 'users' 
	                  AND COLUMN_NAME = 'nip'`
	err = DB.QueryRow(checkNipQuery).Scan(&nipColumnExists)
	if err == nil && nipColumnExists == 0 {
		log.Println("Migrating users table: adding nip column")
		_, err = DB.Exec("ALTER TABLE users ADD COLUMN nip VARCHAR(30) UNIQUE NULL")
		if err != nil {
			log.Fatalf("Error adding nip column to users table: %v", err)
		}
		log.Println("Database migration completed: nip column added to users table")
	}

	// Migrate existing 'Undefined' categories to 'Lain-lain' in tasks and tickets
	_, _ = DB.Exec("UPDATE tasks SET category = 'Lain-lain' WHERE category = 'Undefined'")
	_, _ = DB.Exec("UPDATE tickets SET category = 'Lain-lain' WHERE category = 'Undefined'")
	log.Println("Database migration completed: updated 'Undefined' categories to 'Lain-lain'")
}
