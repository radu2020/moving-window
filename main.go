package main

import (
	"MovingWindowRequest/utils"
	"context"
	"database/sql"
	"encoding/json"
	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"time"
)

const dbPath string = "./simple.db"
const windowDuration int = 60

// Types

type Entry struct {
	Ts time.Time `json:"ts,omitempty"`
	Ip string    `json:"ip,omitempty"`
}

type EntryMap map[string]Entry

type RequestService struct {
	Db    *sql.DB
	Cache *EntryMap
}

type HTTPResponse struct {
	Count   int    `json:"count,omitempty"`
	Message string `json:"message,omitempty"`
}

// DB Operations

func OpenDbConn() (*sql.DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	return conn, err
}

func (s RequestService) CloseDbConn() error {
	return s.Db.Close()
}

func (s RequestService) DropTable() error {
	statement, err := s.Db.Prepare("DROP TABLE IF EXISTS requests;")
	statement.Exec()
	return err
}

func (s RequestService) SetupDB() error {
	statement, err := s.Db.Prepare(
		"CREATE TABLE IF NOT EXISTS requests (id INTEGER PRIMARY KEY AUTOINCREMENT, ts DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL, ip TEXT NOT NULL)")
	statement.Exec()
	return err
}

func (s RequestService) FetchDB() ([]Entry, error) {
	// Fetch DB Data
	dateNow := time.Now().Format("2006-01-02 15:04:05")
	dateMinuteAgo := time.Now().Add(-time.Duration(windowDuration) * time.Second).Format("2006-01-02 15:04:05")
	rows, err := s.Db.Query("SELECT ts, ip FROM requests WHERE ts BETWEEN ? AND ?", dateMinuteAgo, dateNow)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var data []Entry
	for rows.Next() {
		i := Entry{}
		err = rows.Scan(&i.Ts, &i.Ip)
		if err != nil {
			return nil, err
		}
		data = append(data, i)
	}
	return data, nil
}

func (s RequestService) saveCacheToDB() error {
	updateRequests, err := s.Db.Prepare("INSERT INTO requests (ts, ip) VALUES (?, ?)")
	tx, err := s.Db.Begin()
	for _, v := range *s.Cache {
		_, err := tx.Stmt(updateRequests).Exec(v.Ts, v.Ip)
		if err != nil {
			log.Fatal(err)
		}
	}
	err = tx.Commit()
	return err
}

// Caching

func (s RequestService) LoadIntoCache(data []Entry) {
	newMap := *s.Cache
	for _, v := range data {
		newMap[uuid.New().String()] = Entry{
			Ts: v.Ts,
			Ip: v.Ip,
		}
	}
	*s.Cache = newMap
}

func (s RequestService) updateCache(ip string) {
	newMap := *s.Cache
	newMap[uuid.New().String()] = Entry{
		Ts: time.Now(),
		Ip: ip,
	}
	*s.Cache = newMap
}

func (s RequestService) clearCache() {
	minuteAgo := time.Now().Add(-time.Duration(windowDuration) * time.Second)
	for k, v := range *s.Cache {
		if v.Ts.Before(minuteAgo) {
			delete(*s.Cache, k)
		}
	}
}

// HTTP

// RequestCountHandler updates cache and returns total amount of requests in last 60 seconds.
func (s RequestService) RequestCountHandler(w http.ResponseWriter, r *http.Request) {
	s.updateCache(utils.GetUserIP(r))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HTTPResponse{Count: len(*s.Cache)})
	return
}

func main() {
	// Setup resources

	// DB Connection
	conn, err := OpenDbConn()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Service
	service := RequestService{
		Db:    conn,
		Cache: &EntryMap{},
	}

	// Setup DB
	//service.DropTable()
	service.SetupDB()

	// Load recent db data into memory
	data, err := service.FetchDB()
	if err != nil {
		log.Printf("Could not fetch data.")
	}
	service.LoadIntoCache(data)
	log.Printf("%d entries were recovered from DB and loaded into memory.", len(data))

	// Free up memory from old requests
	scheduledJob := func() { service.clearCache() }
	gc := gocron.NewScheduler(time.UTC)
	gc.Every(1).Seconds().Do(scheduledJob)
	gc.StartAsync()

	// Start server
	http.HandleFunc("/count", service.RequestCountHandler)
	srv := &http.Server{
		Addr: ":3000",
	}
	go func() {
		log.Printf("Server available on http://localhost:%d/count\n", 3000)
		srv.ListenAndServe()
	}()

	// Wait for termination signal and register database & http server clean-up operations
	wait := utils.GracefulShutdown(context.Background(), 2*time.Second, map[string]utils.Operation{
		"database": func(ctx context.Context) error {
			// Try to save data, if this fails try to close db
			err := service.saveCacheToDB()
			if err != nil {
				log.Printf("Could not save data. Closing DB Conection gracefully...")
			}
			return service.CloseDbConn()
		},
		"http-server": func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})

	<-wait

}
