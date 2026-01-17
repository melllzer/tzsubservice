package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type Subscription struct {
	ID          string  `json:"id"`
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"` // "MM-YYYY"
	EndDate     *string `json:"end_date,omitempty"`
}

var db *sql.DB

func main() {
	// Пароль изменён на 3228
	dbConn := "host=localhost port=5432 user=postgres password=3228 dbname=subservice_db sslmode=disable"
	var err error
	db, err = sql.Open("postgres", dbConn)
	if err != nil {
		log.Fatal("❌ DB open error:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("❌ DB connection failed. Check:\n- PostgreSQL is running\n- DB 'subservice_db' exists\n- Password is correct (3228)")
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS subscriptions (
			id TEXT PRIMARY KEY,
			service_name TEXT NOT NULL,
			price INTEGER NOT NULL CHECK (price >= 0),
			user_id TEXT NOT NULL,
			start_date TEXT NOT NULL,
			end_date TEXT
		);
	`)
	if err != nil {
		log.Fatal("❌ Table creation failed:", err)
	}

	// Роуты
	http.HandleFunc("/subscriptions", handleSubscriptions)
	http.HandleFunc("/subscriptions/", handleSubscriptionByID)
	http.HandleFunc("/subscriptions/total-cost", handleTotalCost)

	log.Println("✅ Сервер запущен на :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// POST /subscriptions — создать
// GET /subscriptions?user_id=...&service_name=... — список
func handleSubscriptions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createSubscription(w, r)
	case http.MethodGet:
		listSubscriptions(w, r)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

// GET, PUT, DELETE /subscriptions/{id}
func handleSubscriptionByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}
	id := parts[2]

	switch r.Method {
	case http.MethodGet:
		getSubscription(w, r, id)
	case http.MethodPut:
		updateSubscription(w, r, id)
	case http.MethodDelete:
		deleteSubscription(w, r, id)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

// POST /subscriptions/total-cost
func handleTotalCost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Только POST", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID      string `json:"user_id"`
		ServiceName string `json:"service_name,omitempty"`
		PeriodStart string `json:"period_start"` // "MM-YYYY"
		PeriodEnd   string `json:"period_end"`   // "MM-YYYY"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}

	total, err := calculateTotalCost(req.UserID, req.ServiceName, req.PeriodStart, req.PeriodEnd)
	if err != nil {
		http.Error(w, "Ошибка расчёта", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"total_cost": total})
}

// === Реализация ===

func createSubscription(w http.ResponseWriter, r *http.Request) {
	var sub Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}
	sub.ID = "sub_" + time.Now().Format("20060102150405")

	_, err := db.Exec(
		"INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date) VALUES ($1, $2, $3, $4, $5, $6)",
		sub.ID, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate,
	)
	if err != nil {
		log.Printf("Insert error: %v", err)
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sub)
}

func getSubscription(w http.ResponseWriter, r *http.Request, id string) {
	var sub Subscription
	err := db.QueryRow(
		"SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions WHERE id = $1",
		id,
	).Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate)
	if err == sql.ErrNoRows {
		http.Error(w, "Не найдено", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sub)
}

func listSubscriptions(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")

	query := "SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions WHERE 1=1"
	args := []interface{}{}
	argPos := 1

	if userID != "" {
		query += " AND user_id = $" + strconv.Itoa(argPos)
		args = append(args, userID)
		argPos++
	}
	if serviceName != "" {
		query += " AND service_name = $" + strconv.Itoa(argPos)
		args = append(args, serviceName)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, "Ошибка запроса", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var subs []Subscription
	for rows.Next() {
		var s Subscription
		err := rows.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &s.EndDate)
		if err != nil {
			http.Error(w, "Ошибка чтения", http.StatusInternalServerError)
			return
		}
		subs = append(subs, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subs)
}

func updateSubscription(w http.ResponseWriter, r *http.Request, id string) {
	var sub Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}

	result, err := db.Exec(
		"UPDATE subscriptions SET service_name=$1, price=$2, user_id=$3, start_date=$4, end_date=$5 WHERE id=$6",
		sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate, id,
	)
	if err != nil {
		http.Error(w, "Ошибка обновления", http.StatusInternalServerError)
		return
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		http.Error(w, "Не найдено", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteSubscription(w http.ResponseWriter, r *http.Request, id string) {
	result, err := db.Exec("DELETE FROM subscriptions WHERE id = $1", id)
	if err != nil {
		http.Error(w, "Ошибка удаления", http.StatusInternalServerError)
		return
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		http.Error(w, "Не найдено", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Помощник: сравнивает "MM-YYYY" периоды
func isInRange(start, periodStart, periodEnd string) bool {
	// Простая лексикографическая проверка: "01-2025" < "07-2025" < "12-2025"
	return start >= periodStart && start <= periodEnd
}

func calculateTotalCost(userID, serviceName, periodStart, periodEnd string) (int, error) {
	query := `
		SELECT price, start_date FROM subscriptions 
		WHERE user_id = $1 AND ($2 = '' OR service_name = $2)
	`
	args := []interface{}{userID, serviceName}

	rows, err := db.Query(query, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	total := 0
	for rows.Next() {
		var price int
		var startDate string
		if err := rows.Scan(&price, &startDate); err != nil {
			return 0, err
		}
		if isInRange(startDate, periodStart, periodEnd) {
			total += price
		}
	}
	return total, nil
}