package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type RegRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Bid struct {
	ID           int         `json:"bid_id"`
	EmployeeName string      `json:"employee_name"`
	Age          int         `json:"age"`
	OverallExp   int         `json:"overall_experience"`
	SPExp        int         `json:"s_p_experience"`
	IsRead       bool        `json:"is_read"`
	JobTitle     string      `json:"job_title"`
	Subdivision  string      `json:"subdivision"`
	Educations   []Education `json:"educations"`
	Languages    []Language  `json:"languages"`
}

type Language struct {
	Name  string `json:"language"`
	Level string `json:"proficiency"`
}

type Education struct {
	Name  string `json:"name"`
	Place string `json:"place"`
}

var db *sql.DB

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("SSL_MODE"),
	)

	fmt.Println("Connection string:", connStr)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Ошибка при подключении к базе данных: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Ошибка Ping: %v", err)
	}

	fmt.Println("Подключение к PostgreSQL успешно!")

	r := gin.Default()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET не найден в .env")
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8081"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Range"},
		ExposeHeaders:    []string{"Content-Length", "Content-Range"},
		AllowCredentials: true,
	}))

	r.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "-1")
		c.Next()
	})

	r.Static("/static", "../frontend")

	r.GET("/login", func(c *gin.Context) {
		c.File("../frontend/public/login.html")
	})
	r.GET("/register", func(c *gin.Context) {
		c.File("../frontend/public/register.html")
	})
	r.StaticFile("/favicon.ico", "./static/favicon.ico")

	r.GET("/employee", AuthMiddleware(jwtSecret), func(c *gin.Context) {
		userID := c.MustGet("userClaims").(jwt.MapClaims)["user_id"].(string)
		var (
			role string
		)
		err := db.QueryRow(`
			SELECT 
				role
			FROM users 
			WHERE id = $1`,
			userID,
		).Scan(&role)
		if err != nil {
			c.JSON(500, gin.H{"error": "DB error"})
			return
		}
		log.Print(role)
		if role == "employee" {
			c.File("../frontend/public/employee.html")
		} else if role == "admin" {
			c.File("../frontend/public/admin.html")
		} else if role == "user" {
			c.File("../frontend/public/user.html")
		} else {
			c.JSON(500, gin.H{"error": "Not this role"})
			return
		}
	})

	r.GET("/api/user", AuthMiddleware(jwtSecret), func(c *gin.Context) {
		userID := c.MustGet("userClaims").(jwt.MapClaims)["user_id"].(string)

		var (
			username string
			logoURL  string
		)
		err := db.QueryRow(`
			SELECT 
				username,
				COALESCE(NULLIF(logo_url, ''), 'https://i.imgur.com/k8NBJSm.jpg') as logo_url 
			FROM users 
			WHERE id = $1`,
			userID,
		).Scan(&username, &logoURL)

		if err != nil {
			c.JSON(500, gin.H{"error": "DB error"})
			return
		}

		c.JSON(200, gin.H{
			"user":    username,
			"logoURL": logoURL,
		})
	})

	r.GET("/api/messages", AuthMiddleware(jwtSecret), EmployeeMiddleware)

	r.GET("/api/messagesread", AuthMiddleware(jwtSecret), MessagesMiddleware)

	r.POST("/logout", func(c *gin.Context) {
		c.SetCookie("authToken", "", -1, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"message": "Вы успешно вышли"})
	})

	r.GET("/api/job-titles", func(c *gin.Context) {
		rows, err := db.Query("SELECT id, name FROM job_title")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer rows.Close()

		var jobTitles []map[string]interface{}
		for rows.Next() {
			var id int
			var name string
			if err := rows.Scan(&id, &name); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan job titles"})
				return
			}
			jobTitles = append(jobTitles, map[string]interface{}{
				"id":   id,
				"name": name,
			})
		}

		c.JSON(http.StatusOK, jobTitles)
	})

	r.GET("/api/subdivisions", func(c *gin.Context) {
		rows, err := db.Query("SELECT id, name FROM subdivision")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer rows.Close()

		var subdivisions []map[string]interface{}
		for rows.Next() {
			var id int
			var name string
			if err := rows.Scan(&id, &name); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan subdivisions"})
				return
			}
			subdivisions = append(subdivisions, map[string]interface{}{
				"id":   id,
				"name": name,
			})
		}
		c.JSON(http.StatusOK, subdivisions)
	})

	r.GET("/api/languages", func(c *gin.Context) {
		rows, err := db.Query("SELECT id, language FROM languages")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer rows.Close()

		var languages []map[string]interface{}
		for rows.Next() {
			var id int
			var language string
			if err := rows.Scan(&id, &language); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan languages"})
				return
			}
			languages = append(languages, map[string]interface{}{
				"id":       id,
				"language": language,
			})
		}

		c.JSON(http.StatusOK, languages)
	})

	r.GET("/api/educations", func(c *gin.Context) {
		rows, err := db.Query("SELECT id, name FROM education")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer rows.Close()

		var educations []map[string]interface{}
		for rows.Next() {
			var id int
			var name string
			if err := rows.Scan(&id, &name); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan educations"})
				return
			}
			educations = append(educations, map[string]interface{}{
				"id":   id,
				"name": name,
			})
		}

		c.JSON(http.StatusOK, educations)
	})

	r.POST("/api/submit-application", postRequest)

	r.GET("/api/employees/get", GetEmployees)

	r.GET("/api/job_title/get", GetJobTitles)

	r.GET("/api/subdivision/get", GetSubdivisions)

	r.POST("/register", RegisterHandler)

	r.POST("/login", AuthHandler)

	r.POST("/api/accept-application/:id", AuthMiddleware(jwtSecret), acceptRequest)

	r.DELETE("/api/reject-application/:id", AuthMiddleware(jwtSecret), denyRequest)

	r.GET("/api/employees/:id", GetEmployeeByID)

	r.PUT("/api/employees/:id", editEmployees)

	r.DELETE("/api/employees/:id", deleteEmployees)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	fmt.Printf("Сервер запущен на порту %s\n", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}

func RegisterHandler(c *gin.Context) {
	var req RegRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 OR username = $2)", req.Email, req.Username).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Email or username already exists"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password hashing failed"})
		return
	}

	ip := c.ClientIP()

	_, err = db.Exec(`INSERT INTO users (email, username, password_hash, role, registration_date, ip_address) VALUES ($1, $2, $3, 'user', NOW(), $4)`, req.Email, req.Username, string(hashedPassword), ip)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func AuthHandler(c *gin.Context) {
	var req AuthRequest
	if err := c.BindJSON(&req); err != nil {
		log.Printf("Ошибка привязки JSON: %v", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	log.Printf("Поиск пользователя: %s", req.Email)
	var userID string
	var passwordHash string
	var logoURL string
	var username string
	err := db.QueryRow("SELECT id, password_hash, username, COALESCE(NULLIF(logo_url, ''), 'https://i.imgur.com/k8NBJSm.jpg') as logo_url FROM users WHERE LOWER(email) = LOWER($1)", req.Email).Scan(&userID, &passwordHash, &username, &logoURL)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Пользователь не найден: %s", req.Email)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не найден"})
		} else {
			log.Printf("Ошибка базы данных: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных"})
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		log.Printf("Неверный пароль: %s", userID)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Неверные учетные данные"})
		return
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	token, err := GenerateJWT(userID, jwtSecret)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации токена"})
		return
	}

	_, err = db.Exec("UPDATE users SET last_login_date = NOW() WHERE id = $1", userID)
	if err != nil {
		log.Printf("Ошибка обновления last_login_date: %v", err)
	}

	c.SetCookie(
		"authToken",
		token,
		36000,
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   token,
		"user":    username,
		"logoURL": logoURL,
	})
}

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("authToken")
		if err != nil {
			log.Print("Токен отсутствует в куках")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Отсутствует токен"})
			return
		}

		log.Print("Токен из куки: ", tokenString)

		claims, err := validateToken(tokenString, jwtSecret)
		if err != nil {
			log.Printf("Ошибка валидации токена: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Недействительный токен"})
			return
		}
		c.Set("userClaims", claims)
		c.Next()
	}
}

func validateToken(tokenString, jwtSecret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("токен недействителен")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("неверный формат claims")
	}

	return claims, nil
}

func GenerateJWT(userID string, jwtSecret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().UTC().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(jwtSecret))
	return signedToken, err
}

func EmployeeMiddleware(c *gin.Context) {
	rows, err := db.Query(`
        SELECT
        eb.id AS bid_id,
        eb.fio AS employee_name,
        eb.age,
        eb.overall_experience,
        eb.s_p_experience,
        eb.read AS is_read,
        -- Должность
        jt.name AS job_title,
        -- Подразделение
        sd.name AS subdivision,
        -- Образования (массив объектов)
        COALESCE(
            json_agg(DISTINCT jsonb_build_object(
                'name', ed.name,
                'place', eeb.place
            )) FILTER (WHERE ed.name IS NOT NULL), 
            '[]') AS educations,
        -- Языки с уровнями (массив объектов)
        COALESCE(
            json_agg(DISTINCT jsonb_build_object(
                'language', lg.language,
                'proficiency', elb.proficiency
            )) FILTER (WHERE lg.language IS NOT NULL), 
            '[]'
        ) AS languages
        FROM employee_bid eb
        LEFT JOIN job_title jt ON eb.job_title_id = jt.id
        LEFT JOIN subdivision sd ON eb.subdivision_id = sd.id
        LEFT JOIN employee_education_bid eeb ON eb.id = eeb.employee_id
        LEFT JOIN education ed ON eeb.education_id = ed.id
        LEFT JOIN employee_languages_bid elb ON eb.id = elb.employee_id
        LEFT JOIN languages lg ON elb.language_id = lg.id
        GROUP BY 
        eb.id, 
        eb.fio, 
        eb.age, 
        eb.overall_experience, 
        eb.s_p_experience, 
        eb.read,
        jt.name,
        sd.name;
    `)
	if err != nil {
		c.JSON(500, gin.H{"error": "Database error: " + err.Error()})
		return
	}
	defer rows.Close()

	var bids []Bid
	for rows.Next() {
		var bid Bid
		var languagesStr string
		var educationsStr string
		if err := rows.Scan(
			&bid.ID,
			&bid.EmployeeName,
			&bid.Age,
			&bid.OverallExp,
			&bid.SPExp,
			&bid.IsRead,
			&bid.JobTitle,
			&bid.Subdivision,
			&educationsStr,
			&languagesStr,
		); err != nil {
			c.JSON(500, gin.H{"error": "Failed to parse playlists: " + err.Error()})
			return
		}
		if err := json.Unmarshal([]byte(educationsStr), &bid.Educations); err != nil {
			c.JSON(500, gin.H{"error": "Failed to parse educations: " + err.Error()})
			return
		}
		if err := json.Unmarshal([]byte(languagesStr), &bid.Languages); err != nil {
			c.JSON(500, gin.H{"error": "Failed to parse languages: " + err.Error()})
			return
		}
		bids = append(bids, bid)
	}
	if err := rows.Err(); err != nil {
		c.JSON(500, gin.H{"error": "Database error: " + err.Error()})
		return
	}
	if len(bids) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Заявки не найдены"})
		return
	}
	c.JSON(http.StatusOK, bids)
}

func MessagesMiddleware(c *gin.Context) {
	_, err := db.Exec("UPDATE employee_bid SET read = true")
	if err != nil {
		log.Printf("Ошибка выполнения запроса UPDATE: %v", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Rows updated successfully"})
}

func postRequest(c *gin.Context) {
	var req struct {
		FIO               string `json:"fio"`
		Age               int    `json:"age"`
		OverallExperience int    `json:"overall_experience"`
		SPExperience      int    `json:"s_p_experience"`
		JobTitleID        int    `json:"job_title_id"`
		SubdivisionID     int    `json:"subdivision_id"`
		Languages         []struct {
			LanguageID  int    `json:"language_id"`
			Proficiency string `json:"proficiency"`
		} `json:"languages"`
		Educations []struct {
			EducationID int    `json:"education_id"`
			Place       string `json:"place"`
		} `json:"educations"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database transaction error"})
		return
	}

	var employeeID int
	err = tx.QueryRow(`
        INSERT INTO employee_bid (
            fio, age, overall_experience, s_p_experience, job_title_id, subdivision_id
        ) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
    `, req.FIO, req.Age, req.OverallExperience, req.SPExperience, req.JobTitleID, req.SubdivisionID).Scan(&employeeID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert into employee_bid"})
		log.Printf("Failed to insert into employee_bid")
		return
	}

	for _, lang := range req.Languages {
		_, err := tx.Exec(`
            INSERT INTO employee_languages_bid (employee_id, language_id, proficiency)
            VALUES ($1, $2, $3)
        `, employeeID, lang.LanguageID, lang.Proficiency)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert into employee_languages_bid"})
			log.Printf("Failed to insert into employee_languages_bid")
			return
		}
	}

	for _, edu := range req.Educations {
		_, err := tx.Exec(`
            INSERT INTO employee_education_bid (employee_id, education_id, place)
            VALUES ($1, $2, $3)
        `, employeeID, edu.EducationID, edu.Place)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert into employee_education_bid"})
			log.Printf("Failed to insert into employee_education_bid")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Application submitted successfully"})
}

func acceptRequest(c *gin.Context) {
	bidID := c.Param("id")

	tx, err := db.Begin()
	if err != nil {
		log.Printf("Ошибка начала транзакции: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database transaction error"})
		return
	}

	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM employee_bid WHERE id = $1)", bidID).Scan(&exists)
	if err != nil || !exists {
		tx.Rollback()
		log.Printf("Заявка с ID %s не найдена", bidID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Заявка не найдена"})
		return
	}

	var employeeID int
	err = tx.QueryRow(`
        INSERT INTO employee (
            fio, age, overall_experience, s_p_experience, job_title_id, subdivision_id
        )
        SELECT 
            fio, age, overall_experience, s_p_experience, job_title_id, subdivision_id
        FROM employee_bid
        WHERE id = $1
        RETURNING id
    `, bidID).Scan(&employeeID)
	if err != nil {
		tx.Rollback()
		log.Printf("Ошибка копирования данных в таблицу employee: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to copy data into employee table"})
		return
	}

	_, err = tx.Exec(`
        INSERT INTO employee_languages (employee_id, language_id, proficiency)
        SELECT 
            $1, language_id, proficiency
        FROM employee_languages_bid
        WHERE employee_id = $2
    `, employeeID, bidID)
	if err != nil {
		tx.Rollback()
		log.Printf("Ошибка копирования данных в таблицу employee_languages: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to copy data into employee_languages table"})
		return
	}

	_, err = tx.Exec(`
        INSERT INTO employee_education (employee_id, education_id, place)
        SELECT 
            $1, education_id, place
        FROM employee_education_bid
        WHERE employee_id = $2
    `, employeeID, bidID)
	if err != nil {
		tx.Rollback()
		log.Printf("Ошибка копирования данных в таблицу employee_education: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to copy data into employee_education table"})
		return
	}

	_, err = tx.Exec("DELETE FROM employee_languages_bid WHERE employee_id = $1", bidID)
	if err != nil {
		tx.Rollback()
		log.Printf("Ошибка удаления данных из таблицы employee_languages_bid: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete data from employee_languages_bid table"})
		return
	}

	_, err = tx.Exec("DELETE FROM employee_education_bid WHERE employee_id = $1", bidID)
	if err != nil {
		tx.Rollback()
		log.Printf("Ошибка удаления данных из таблицы employee_education_bid: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete data from employee_education_bid table"})
		return
	}

	_, err = tx.Exec("DELETE FROM employee_bid WHERE id = $1", bidID)
	if err != nil {
		tx.Rollback()
		log.Printf("Ошибка удаления данных из таблицы employee_bid: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete data from employee_bid table"})
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Ошибка фиксации транзакции: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
		return
	}

	log.Printf("Заявка с ID %s успешно принята", bidID)
	c.JSON(http.StatusOK, gin.H{"message": "Application accepted successfully"})
}

func denyRequest(c *gin.Context) {
	bidID := c.Param("id")

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database transaction error"})
		return
	}

	_, err = tx.Exec("DELETE FROM employee_languages_bid WHERE employee_id = $1", bidID)
	if err != nil {
		tx.Rollback()
		log.Printf("Ошибка удаления данных из таблицы employee_languages_bid: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete data from employee_languages_bid table"})
		return
	}

	_, err = tx.Exec("DELETE FROM employee_education_bid WHERE employee_id = $1", bidID)
	if err != nil {
		tx.Rollback()
		log.Printf("Ошибка удаления данных из таблицы employee_education_bid: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete data from employee_education_bid table"})
		return
	}

	_, err = tx.Exec("DELETE FROM employee_bid WHERE id = $1", bidID)
	if err != nil {
		tx.Rollback()
		log.Printf("Ошибка удаления данных из таблицы employee_bid: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete data from employee_bid table"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Application rejected successfully"})
}

func GetEmployees(c *gin.Context) {
	id := c.Query("id")
	fio := c.Query("fio")
	age := c.Query("age")
	jobTitle := c.Query("job_title_id")
	subdivision := c.Query("subdivision_id")
	overall := c.Query("overall_experience")
	s_p := c.Query("s_p_experience")

	query := `
        SELECT 
            id, fio, age, job_title_id, subdivision_id, overall_experience, s_p_experience
        FROM employee
        WHERE 1=1
    `
	args := []interface{}{}

	if id != "" {
		query += " AND id = $" + strconv.Itoa(len(args)+1)
		args = append(args, id)
	}

	if fio != "" {
		query += " AND fio ILIKE $1"
		args = append(args, "%"+fio+"%")
	}
	if age != "" {
		query += " AND age = $" + strconv.Itoa(len(args)+1)
		args = append(args, age)
	}
	if jobTitle != "" {
		query += " AND job_title_id = $" + strconv.Itoa(len(args)+1)
		args = append(args, jobTitle)
	}
	if subdivision != "" {
		query += " AND subdivision_id = $" + strconv.Itoa(len(args)+1)
		args = append(args, subdivision)
	}
	if overall != "" {
		query += " AND overall_experience = $" + strconv.Itoa(len(args)+1)
		args = append(args, overall)
	}
	if s_p != "" {
		query += " AND s_p_experience = $" + strconv.Itoa(len(args)+1)
		args = append(args, s_p)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var employees []map[string]interface{}
	for rows.Next() {
		var employee struct {
			ID                int    `db:"id"`
			FIO               string `db:"fio"`
			Age               int    `db:"age"`
			JobTitleID        int    `db:"job_title_id"`
			SubdivisionID     int    `db:"subdivision_id"`
			OverallExperience int    `db:"overall_experience"`
			SPExperience      int    `db:"s_p_experience"`
		}

		if err := rows.Scan(
			&employee.ID,
			&employee.FIO,
			&employee.Age,
			&employee.JobTitleID,
			&employee.SubdivisionID,
			&employee.OverallExperience,
			&employee.SPExperience,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan employees"})
			return
		}

		employees = append(employees, map[string]interface{}{
			"id":                 employee.ID,
			"fio":                employee.FIO,
			"age":                employee.Age,
			"job_title_id":       employee.JobTitleID,
			"subdivision_id":     employee.SubdivisionID,
			"overall_experience": employee.OverallExperience,
			"s_p_experience":     employee.SPExperience,
		})
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, employees)
}

func GetEmployeeByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	var employee struct {
		ID                int    `db:"id" json:"id"`
		FIO               string `db:"fio" json:"fio"`
		Age               int    `db:"age" json:"age"`
		JobTitleID        int    `db:"job_title_id" json:"job_title_id"`
		SubdivisionID     int    `db:"subdivision_id" json:"subdivision_id"`
		OverallExperience int    `db:"overall_experience" json:"overall_experience"`
		SPExperience      int    `db:"s_p_experience" json:"s_p_experience"`
	}

	err := db.QueryRow(`
        SELECT 
        id, fio, age, job_title_id, subdivision_id, overall_experience, s_p_experience
        FROM employee
        WHERE id = $1
    `, id).Scan(
		&employee.ID,
		&employee.FIO,
		&employee.Age,
		&employee.JobTitleID,
		&employee.SubdivisionID,
		&employee.OverallExperience,
		&employee.SPExperience,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, employee)
}

func editEmployees(c *gin.Context) {
	id := c.Param("id")

	type UpdateEmployee struct {
		FIO               string `json:"fio" binding:"required"`
		Age               int    `json:"age" binding:"min=0"`
		JobTitleID        int    `json:"job_title_id" binding:"required,min=1"`
		SubdivisionID     int    `json:"subdivision_id" binding:"required,min=1"`
		OverallExperience int    `json:"overall_experience" binding:"min=0"`
		SPExperience      int    `json:"s_p_experience" binding:"min=0"`
	}

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	var employee UpdateEmployee
	if err := c.ShouldBindJSON(&employee); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `
        UPDATE employee 
        SET 
            fio = $1,
            age = $2,
            job_title_id = $3,
            subdivision_id = $4,
            overall_experience = $5,
            s_p_experience = $6
        WHERE id = $7
        RETURNING *
    `

	var updatedEmployee struct {
		ID                int    `db:"id"`
		FIO               string `db:"fio"`
		Age               int    `db:"age"`
		JobTitleID        int    `db:"job_title_id"`
		SubdivisionID     int    `db:"subdivision_id"`
		OverallExperience int    `db:"overall_experience"`
		SPExperience      int    `db:"s_p_experience"`
	}

	err := db.QueryRow(query,
		employee.FIO,
		employee.Age,
		employee.JobTitleID,
		employee.SubdivisionID,
		employee.OverallExperience,
		employee.SPExperience,
		id,
	).Scan(
		&updatedEmployee.ID,
		&updatedEmployee.FIO,
		&updatedEmployee.Age,
		&updatedEmployee.JobTitleID,
		&updatedEmployee.SubdivisionID,
		&updatedEmployee.OverallExperience,
		&updatedEmployee.SPExperience,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                 updatedEmployee.ID,
		"fio":                updatedEmployee.FIO,
		"age":                updatedEmployee.Age,
		"job_title_id":       updatedEmployee.JobTitleID,
		"subdivision_id":     updatedEmployee.SubdivisionID,
		"overall_experience": updatedEmployee.OverallExperience,
		"s_p_experience":     updatedEmployee.SPExperience,
	})
}

func deleteEmployees(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database transaction error"})
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM employee_languages WHERE employee_id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete from employee_languages"})
		return
	}

	_, err = tx.Exec("DELETE FROM employee_education WHERE employee_id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete from employee_education"})
		return
	}

	result, err := tx.Exec("DELETE FROM employee WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}

	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
		return
	}

	c.Status(http.StatusNoContent)
}

func GetJobTitles(c *gin.Context) {
	id := c.Query("id")
	name := c.Query("name")
	count := c.Query("count")

	query := `
        SELECT id, name, count
        FROM job_title
        WHERE 1=1
    `
	args := []interface{}{}

	if id != "" {
		query += " AND id = $" + strconv.Itoa(len(args)+1)
		args = append(args, id)
	}

	if name != "" {
		query += " AND name ILIKE $1"
		args = append(args, "%"+name+"%")
	}
	if count != "" {
		query += " AND count = $" + strconv.Itoa(len(args)+1)
		args = append(args, count)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var job_titles []map[string]interface{}
	for rows.Next() {
		var job_title struct {
			ID    int    `db:"id"`
			Name  string `db:"name"`
			Count int    `db:"count"`
		}

		if err := rows.Scan(
			&job_title.ID,
			&job_title.Name,
			&job_title.Count,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan job titles"})
			return
		}

		job_titles = append(job_titles, map[string]interface{}{
			"id":    job_title.ID,
			"name":  job_title.Name,
			"count": job_title.Count,
		})
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, job_titles)
}

func GetSubdivisions(c *gin.Context) {
	id := c.Query("id")
	name := c.Query("name")

	query := `
        SELECT id, name
        FROM subdivision
        WHERE 1=1
    `
	args := []interface{}{}

	if id != "" {
		query += " AND id = $" + strconv.Itoa(len(args)+1)
		args = append(args, id)
	}

	if name != "" {
		query += " AND name ILIKE $1"
		args = append(args, "%"+name+"%")
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var subdivisions []map[string]interface{}
	for rows.Next() {
		var subdivision struct {
			ID   int    `db:"id"`
			Name string `db:"name"`
		}

		if err := rows.Scan(
			&subdivision.ID,
			&subdivision.Name,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan subdivisions"})
			return
		}

		subdivisions = append(subdivisions, map[string]interface{}{
			"id":   subdivision.ID,
			"name": subdivision.Name,
		})
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, subdivisions)
}
