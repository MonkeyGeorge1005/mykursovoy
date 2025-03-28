package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chromedp/chromedp"

	"github.com/gin-gonic/gin"
)

func TestPostRequest(t *testing.T) {
	connStr := "user=postgres dbname=Cursovoy sslmode=disable password=Djcmvfv583746 host=localhost port=5432"
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

	// Установка режима тестирования для Gin
	gin.SetMode(gin.TestMode)

	// Подготовка тестовых данных
	reqBody := map[string]interface{}{
		"fio":                "Тестовый Тест Тестович",
		"overall_experience": 7,
		"s_p_experience":     5,
		"job_title_id":       1,
		"subdivision_id":     2,
		"languages": []map[string]interface{}{
			{
				"language_id": 1,
				"proficiency": "B2",
			},
		},
		"educations": []map[string]interface{}{
			{
				"education_id": 1,
				"place":        "Тестовый университет",
			},
		},
	}

	body, _ := json.Marshal(reqBody)

	// Создание HTTP POST запроса
	req, err := http.NewRequest("POST", "/api/submit-application", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Создание ResponseRecorder для записи ответа
	w := httptest.NewRecorder()

	// Создание нового маршрутизатора Gin
	router := gin.Default()
	router.POST("/api/submit-application", postRequest)

	// Выполнение запроса
	router.ServeHTTP(w, req)

	// Проверка статус кода
	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Проверка тела ответа
	expected := `{"message":"Application submitted successfully"}`
	if w.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			w.Body.String(), expected)
	}

	// Дополнительно проверяем, что данные действительно были добавлены в базу данных
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM employee_bid WHERE fio = $1", "Тестовый Тест Тестович").Scan(&count)
	if err != nil {
		t.Fatalf("Database query error: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 record in database, got %d", count)
	}
}

func TestE2EFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("window-size", "1920,1080"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx,
		chromedp.WithLogf(t.Logf),
		chromedp.WithDebugf(t.Logf),
	)
	defer cancel()

	const (
		loginURL = "http://localhost:8081/login"
		email    = "test@example.com"
		password = "58374658"
	)

	err := chromedp.Run(ctx,
		chromedp.Navigate(loginURL),
		chromedp.WaitReady(`#email`, chromedp.ByQuery),
		chromedp.SendKeys(`#email`, email, chromedp.ByQuery),
		chromedp.SendKeys(`#password`, password, chromedp.ByQuery),
		chromedp.Click(`#login`, chromedp.ByQuery),
		chromedp.WaitReady(`#main-container`, chromedp.ByQuery),

		// Ожидаем загрузки выпадающих списков
		chromedp.WaitReady(`#job_title option:nth-child(2)`, chromedp.ByQuery),
		chromedp.WaitReady(`#subdivision option:nth-child(2)`, chromedp.ByQuery),

		// Заполнение основных полей
		chromedp.SendKeys(`#fio`, "Иванов Иван Иванович", chromedp.ByQuery),
		chromedp.SendKeys(`#age`, "30", chromedp.ByQuery),
		chromedp.SetValue(`#job_title`, "2", chromedp.ByQuery), // Используем ID значения
		chromedp.SetValue(`#subdivision`, "3", chromedp.ByQuery),
		chromedp.SendKeys(`#overall_experience`, "4", chromedp.ByQuery),
		chromedp.SendKeys(`#s_p_experience`, "3", chromedp.ByQuery),

		// Добавление языка
		chromedp.Click(`#addLanguage`, chromedp.ByQuery),
		chromedp.WaitReady(`#languages-container .dynamic-input-group:last-child select`, chromedp.ByQuery),
		chromedp.SetValue(
			`#languages-container .dynamic-input-group:last-child select[name^="language_"]`,
			"1", // ID языка
			chromedp.ByQuery,
		),
		chromedp.SetValue(
			`#languages-container .dynamic-input-group:last-child select[name^="proficiency_"]`,
			"B2",
			chromedp.ByQuery,
		),

		// Добавление образования
		chromedp.Click(`#addEducation`, chromedp.ByQuery),
		chromedp.WaitReady(`#educations-container .dynamic-input-group:last-child select`, chromedp.ByQuery),
		chromedp.SetValue(
			`#educations-container .dynamic-input-group:last-child select[name^="education_"]`,
			"1", // ID образования
			chromedp.ByQuery,
		),
		chromedp.SendKeys(
			`#educations-container .dynamic-input-group:last-child input[name^="education_place_"]`,
			"Москва",
			chromedp.ByQuery,
		),

		// Отправка формы
		chromedp.Click(`#submitApplication`, chromedp.ByQuery),
		chromedp.WaitVisible(`.success-message`, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Тест провален: %v", err)
	}

	log.Print("Тест прошёл успешно!")
}
