package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"wedding-go/models"

	dbconnection "wedding-go/db-connection"
	dbModels "wedding-go/models"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("hello handler")
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprint(w, "Hello, World!")
}

func greetingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Name string `json:"name"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		log.Printf("json decode failed: %v", err)
		return
	}

	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintf(w, "Hello, %s!", body.Name)
}

func populateGuests(db *sql.DB) http.HandlerFunc {
	guestModel := models.NewGuestModel(db)
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		var payload dbModels.NewFamilyPayload

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		type Family struct {
			FamilyName string `json:"family_name"`
			FamilyID   int    `json:"family_id"`
		}
		type responseBody struct {
			Message  string   `json:"message"`
			Families []Family `json:"families"`
		}
		var resp responseBody

		var familiesInsertErrors int

		for _, fam := range payload.Families {

			familyName, familyID, err := guestModel.InsertNewFamiliesAndGuests(fam)
			if err != nil {
				log.Printf("Error adding family %v: %v", fam.FamilyName, err)
				familiesInsertErrors++
				continue
			}

			newFamily := Family{
				FamilyName: familyName,
				FamilyID:   familyID,
			}
			resp.Families = append(resp.Families, newFamily)
		}

		if familiesInsertErrors > 0 {
			if familiesInsertErrors == len(payload.Families) {
				resp.Message = "All families failed to insert"
			} else {
				resp.Message = "Some families failed to be inserted"
			}
		} else {
			resp.Message = "Successfully added families"
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			return
		}
	}
}

func searchInviteHandler(db *sql.DB) http.HandlerFunc {
	guestModel := models.NewGuestModel(db)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		var rsvpCode = strings.ReplaceAll(r.URL.Query().Get("rsvpCode"), " ", "")
		if rsvpCode == "" {
			http.Error(w, "Missing rsvpCode in API call", http.StatusBadRequest)
			return
		}

		var family dbModels.Family
		type Response struct {
			Message string          `json:"message"`
			Code    int             `json:"code"`
			Data    dbModels.Family `json:"data"`
		}
		var err error

		family, err = guestModel.GetFamilyByInvitationCode(rsvpCode)
		if err != nil {
			log.Printf("query failed: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if family.FamilyMembers == nil {
			resp := Response{
				Message: "Invitation code not found",
				Code:    http.StatusNotFound,
				Data: dbModels.Family{
					FamilyMembers: []dbModels.Guest{},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				log.Printf("json encode failed: %v", err)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		resp := Response{
			Message: "Success",
			Code:    http.StatusOK,
			Data:    family,
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("json encode failed: %v", err)
		}
	}
}

func handleRsvpResponse(db *sql.DB) http.HandlerFunc {
	guestModel := models.NewGuestModel(db)
	log.Printf("In handleRsvpResponse")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		var payload dbModels.RsvpResponse

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("Responding RSVPs for family %d with values %v", payload.FamilyID, payload.Responses)

		err := guestModel.RespondRsvp(payload)
		if err != nil {
			log.Printf("Failed to respond to rsvp: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode("It's all good!")
		if err != nil {
			return
		}
	}
}

// functions to get environment variables
func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
func getEnvInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		log.Printf("invalid %s value %q, using fallback %d", key, raw, fallback)
		return fallback
	}
	return value
}

// CORS stuff
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:6767")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// handle preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	log.Println("Initializing database")

	// TODO: offload this to environment variables once app is running in Kubernetes
	config := dbconnection.DBConnection{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvInt("DB_PORT", 5432),
		Username: getEnv("DB_USER", "wedding_user"),
		Password: getEnv("DB_PASSWORD", "wedding_pass"),
		Database: getEnv("DB_NAME", "weddingdb"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Connect to Database
	db, err := dbconnection.Connect(config)
	if err != nil {
		log.Printf("connect failed: %v", err)
		return
	}

	// Close Database connection when exiting program
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("Closing DB Connection failed with: %v", err)
		}
	}(db)

	mux := http.NewServeMux()

	mux.HandleFunc("/hello", helloHandler)
	mux.HandleFunc("/greeting", greetingHandler)
	mux.HandleFunc("/populateGuests", populateGuests(db))
	mux.HandleFunc("/searchInvite", searchInviteHandler(db))
	mux.HandleFunc("/respond-rsvp", handleRsvpResponse(db))

	log.Println("Listening on :8080")
	if err := http.ListenAndServe(":8080", cors(mux)); err != nil {
		log.Fatal(err)
	}
}
