package db

import (
	"database/sql"
	"fmt"
	"log"
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
		log.Fatalf("Error opening connection to MySQL server: %v", err)
	}
	defer tempDB.Close()

	// Verify MySQL server is reachable
	err = tempDB.Ping()
	if err != nil {
		log.Fatalf("Error pinging MySQL server (please verify MySQL is running and credentials in .env are correct): %v", err)
	}

	// Create database if it doesn't exist
	_, err = tempDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", cfg.DBName))
	if err != nil {
		log.Fatalf("Error creating database %s: %v", cfg.DBName, err)
	}
	log.Printf("Database '%s' verified/created successfully\n", cfg.DBName)

	// 2. Establish connection to the target database
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening connection to database %s: %v", cfg.DBName, err)
	}

	// Connection pool settings
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	err = DB.Ping()
	if err != nil {
		log.Fatalf("Error pinging database %s: %v", cfg.DBName, err)
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
		log.Println("Database migration completed: client_id column added to tasks table")
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
}
