package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/MakMoinee/go-mith/pkg/email"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var port int = 8002

// Type Structures

type User struct {
	UserID               int
	Name                 string
	EmailAddr            string
	ContactNo            string
	MembershipTier       string
	PasswordHash         string
	IsActivated          int
	VerificationCodeHash string
}

type Car struct {
	CarID      int
	Model      string
	PlateNo    string
	RentalRate int
	Location   string
}

type Booking struct {
	BookingID int
	StartTime string
	EndTime   string
	Date      string
	CarID     int
	Model     string
	UserID    int
	PaymentID int
}

type Payment struct {
	PaymentID   int
	Amount      int
	DateCreated string
	Status      string
	UserID      int
	CarID       int
}

var db *sql.DB

// Function to connect Database -- MUST BE USED AT ALL CRUD FUNCTIONS
func connectToDB() (*sql.DB, error) {
	if db != nil {
		// Check if the database connection is already established
		err := db.Ping()
		if err == nil {
			return db, nil
		}
	}

	// If not connected or there's an error, establish a new connection
	db, err := sql.Open("mysql", "root:Shskjssk10!@tcp(127.0.0.1:3306)/CNADAssg1DB")
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
		return nil, err
	}

	// Ping the database to ensure the connection is successful
	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging the database:", err)
		return nil, err
	}

	return db, nil
}

func main() {
	router := mux.NewRouter()

	// Test Initial Database Connection
	router.HandleFunc("/api/v1/test", testingDB).Methods("GET")

	// Routes
	router.HandleFunc("/api/v1/payment", postPayment).Methods("POST")
	router.HandleFunc("/api/v1/paymentConfirmation", sendReceipt).Methods("POST")

	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://127.0.0.1:8002"}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS", "PUT"}),
	)(router)

	// Print port
	fmt.Printf("Listening at port %d\n", port)
	url := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(url, corsHandler))
}

// Create Payment
func postPayment(w http.ResponseWriter, r *http.Request) {
	// Connect to Database
	db, err := connectToDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println("Error connecting to the database")
		return
	}

	// Read Data from Body
	var newPayment Payment
	err = json.NewDecoder(r.Body).Decode(&newPayment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Posting Payment into Database
	result, err := db.Exec(`
		INSERT INTO Payment (Amount, Status, UserID, CarID)
		VALUES 
		(?, 'Successful', ?, ?)`, newPayment.Amount, newPayment.UserID, newPayment.CarID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println("Something went wrong with creation")
		return
	}

	// Get the last inserted ID
	lastInsertId, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println("Error getting last inserted ID")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Create a response struct
	type PaymentResponse struct {
		PaymentID int `json:"PaymentID"`
	}

	// Create a new PaymentResponse instance
	response := PaymentResponse{
		PaymentID: int(lastInsertId),
	}

	json.NewEncoder(w).Encode(response)
	w.WriteHeader(http.StatusOK)
}

func sendReceipt(w http.ResponseWriter, r *http.Request) {
	// Get Payment Details from Body
	type Email struct {
		Name      string
		EmailAddr string
		Model     string
		Date      string
		StartTime string
		EndTime   string
		Amount    int
	}

	var emailDetails Email

	err := json.NewDecoder(r.Body).Decode(&emailDetails)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Getting Secret Code
	godotenv.Load("../../.env")
	var emailKey = os.Getenv("EMAIL_KEY")

	// Send Email Verification Code
	emailService := email.NewEmailService(587, "smtp.gmail.com", "pookiebears2006@gmail.com", emailKey)

	var messageBody string

	messageBody = fmt.Sprintf(`
Dear %s,
	
This email confirms your booking for %s on %s %s to %s and that payment of $%d has been made.

Thank you for trusting us! We hope you have a wonderful time!
	`, emailDetails.Name, emailDetails.Model, emailDetails.Date, emailDetails.StartTime, emailDetails.EndTime, emailDetails.Amount)

	isEmailSent, err := emailService.SendEmail(emailDetails.EmailAddr, "Payment Confirmed", messageBody)
	if err != nil {
		log.Fatalf("Error sending email: %s", err)
	}

	if isEmailSent {
		log.Println("Email Send Successfully")
	} else {
		log.Println("Failed to send email")
	}
}

// Test Database Connection
func testingDB(w http.ResponseWriter, r *http.Request) {
	db, err := connectToDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println("Error connecting to the database")
		return
	}
	fmt.Println("Database has been successfully connected!")
	defer db.Close()
}
