package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type USD struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type ApiServer struct {
	Db *sql.DB
}

func main() {
	//Fazemos a conexao com o sqlite
	db, err := sql.Open("sqlite3", "./project.db")
	if err != nil {
		log.Fatal(err)
	}
	//Tratamos para o caso do db n ter startado
	if db == nil {
		log.Fatal("db is nil")
	}

	defer db.Close()

	sqlStmt := `
		CREATE TABLE IF NOT EXISTS moneyInfo (
			bid TEXT
		);
	`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal("%q: %s\n", err, sqlStmt)
	}

	//Criamos uma nova instancia da ApiServer e inicializamos o campo Db com a conexao feita acima
	//& significa que estamos apontando para o endereco de memoria
	server := &ApiServer{Db: db}
	if server.Db == nil {
		log.Fatal("server.Db is nil")
	}
	server.Start()
}

// Criamos a funcao de start para inicializar o servidor e chamar o handler
func (s *ApiServer) Start() {
	http.HandleFunc("/cotacao", s.apiCall)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *ApiServer) apiCall(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)

	defer cancel()

	price, err := searchValue(ctx)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Println("Operation took longer than 200ms")
		} else {
			log.Println("Operation failed:", err)
		}
	}

	addBidToDatabase(ctx, s.Db, price.USDBRL.Bid)

	if errors.Is(err, context.DeadlineExceeded) {
		log.Println("Operation took longer than expected")
	} else {
		log.Println("Operation failed:", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(price)
}

func searchValue(ctx context.Context) (*USD, error) {
	req, err := http.NewRequest("GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	//TODO - entender direito o DO
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var price USD

	err = json.Unmarshal(body, &price)
	if err != nil {
		return nil, err
	}

	return &price, nil
}

func addBidToDatabase(ctx context.Context, db *sql.DB, bid string) {
	sqlStmt := "INSERT INTO moneyInfo (bid) VALUES (?);"

	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	_, err := db.ExecContext(ctx, sqlStmt, bid)

	if errors.Is(err, context.DeadlineExceeded) {
		log.Println("Operation took longer than 10ms")
	} else {
		log.Println("Operation failed:", err)
	}

}
