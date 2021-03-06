package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var db *sql.DB

var router *mux.Router

var addr = flag.String("addr", "localhost:3001", "http service address")

const (
	dbhost = "localhost"
	dbport = 5432
	dbuser = "caninodev"
	dbpass = "testing"
	dbname = "lecheleccion"
)

type Article struct {
	ID                  int
	URL                 string
	URLToImage          string
	Source              string
	PublicationDate     time.Time
	Title               string
	Body                string
	ExternalReferenceID int
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type BiasMeta struct {
	ID           int
	Librt        float32
	Grn          float32
	Lbrl         float32
	Cnsvrt      float32
	biasableType string
	biasableID   int
	createdAt    time.Time
	updatedAt    time.Time
}
type ArticlesMessage struct {
	Articles []Article
	Librt    float32
	Grn      float32
	Lbrl     float32
	Cnsvrt   float32
}

func main() {
	initConnDB()
	defer db.Close()
	createRouter()
}

func createRouter() {
	router = mux.NewRouter()
	router.HandleFunc("/api/articles", GetArticles).Methods("GET")
	router.HandleFunc("/api/{requestType}/{id}", HandleType).Methods("GET")
	router.HandleFunc("/api/user/{id}", GetUser).Methods("GET")
	log.Fatal(http.ListenAndServe(":3001", router))
}

func HandleType(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var jsonData []byte
	var err error
	articleID := params["id"]
	switch params["requestType"] {
	case "article":
		data, err := queryArticle(articleID)
		if err == nil {
			w.WriteHeader(http.StatusOK)
		}
		jsonData, err = json.Marshal(data)
	case "bias":
		data, err := queryBias(articleID)
		if err == nil {
			w.WriteHeader(http.StatusOK)
		}
		jsonData, err = json.Marshal(data)
	default:
		w.WriteHeader(http.StatusBadRequest)
		err = json.NewEncoder(w).Encode([]byte("Malformed params. Please try again."))
	}
	_, err = w.Write(jsonData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
	}
}

func (a *ArticlesMessage) queryAverages() {

	sqlStatement := `SELECT AVG(t.Libertarian) AS "Libertarian", AVG(t.Green) AS "Green", AVG(t.Liberal) AS "Liberal", AVG(t.Conservative) AS "Conservative" FROM collections.biases t;`

	row := db.QueryRow(sqlStatement)
	err := row.Scan(
		&a.Librt,
		&a.Grn,
		&a.Lbrl,
		&a.Cnsvrt,
	)
	if err != nil {
		log.Print(err)
	}

}

func GetArticles(w http.ResponseWriter, r *http.Request) {
	var msg ArticlesMessage
	msg.Articles = queryArticles()
	msg.queryAverages()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(msg)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	// ...
}

func queryArticles() []Article {
	var articles []Article

	sqlStatement := `SELECT t.* FROM collections.articles t LIMIT 50`

	rows, err := db.Query(sqlStatement)
	if err != nil {
		log.Print("Error: ", err)
	}

	defer rows.Close()

	for rows.Next() {
		article := Article{}
		err = rows.Scan(
			&article.ID,
			&article.URL,
			&article.URLToImage,
			&article.Source,
			&article.PublicationDate,
			&article.Title,
			&article.Body,
			&article.ExternalReferenceID,
			&article.CreatedAt,
			&article.UpdatedAt)

		articles = append(articles, article)
	}

	return articles

}

func queryArticle(articleID string) (Article, error) {
	article := Article{}

	log.Print(articleID)
	sqlStatement := `SELECT t.* FROM collections.articles t WHERE id = $1`

	row := db.QueryRow(sqlStatement, articleID)

	err := row.Scan(
		&article.ID,
		&article.URL,
		&article.URLToImage,
		&article.Source,
		&article.PublicationDate,
		&article.Title,
		&article.Body,
		&article.ExternalReferenceID,
		&article.CreatedAt,
		&article.UpdatedAt)

	return article, err
}

func queryBias(articleID string) (BiasMeta, error) {
	var bias BiasMeta

	sqlStatement := `SELECT t.* FROM collections.biases t WHERE biasable_id = $1`

	row := db.QueryRow(sqlStatement, articleID)

	err := row.Scan(
		&bias.ID,
		&bias.Librt,
		&bias.Grn,
		&bias.Lbrl,
		&bias.Cnsvrt,
		&bias.biasableType,
		&bias.biasableID,
		&bias.createdAt,
		&bias.updatedAt)

	return bias, err
}

func initConnDB() {
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbhost, dbport,
		dbuser, dbpass, dbname)

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected!")
}
