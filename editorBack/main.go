package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/dgraph-io/dgo/v210/protos/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Code struct {
	Code []string `json:"Code"`
	Uid  string   `json:"uid"`
}

func main() {
	port := "3000"
	r := chi.NewRouter()

	r.Use(middleware.Logger,
		middleware.Recoverer,
	)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "PUT", "POST", "DELETE", "HEAD", "OPTION"},
		AllowedHeaders:   []string{"User-Agent", "Content-Type", "Accept", "Accept-Encoding", "Accept-Language", "Cache-Control", "Connection", "DNT", "Host", "Origin", "Pragma", "Referer"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Bienvenido-Go!!"))
	})

	// EndPoints

	//get all programs
	r.Get("/programs", getPrograms)

	//get all programs
	r.Get("/page/{uid}", getProgramsPage)

	//save programs
	r.Post("/programs", addProgram)

	//get program
	r.Get("/program/{uid}", getProgram)

	//run program
	r.Post("/run/program", runProgram)

	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatal(err)
	}
}

// getPrograms get all the stored programs in the database
func getPrograms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dgClient := newClient()
	txn := dgClient.NewTxn()

	resp, err := txn.Query(context.Background(), queryAllPrograms)
	if err != nil {
		log.Fatal(err)
	}

	w.Write(resp.Json)
}

//addProgram save the program in the database
func addProgram(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var rawCode Code

	err := json.NewDecoder(r.Body).Decode(&rawCode)
	if err != nil {
		log.Fatal(err)
	}

	p := Code{Code: rawCode.Code}

	pb, err := json.Marshal(p)
	if err != nil {
		log.Fatal(err)
	}

	dgClient := newClient()
	txn := dgClient.NewTxn()

	mu := &api.Mutation{
		CommitNow: true,
		SetJson:   pb,
	}

	resp, err := txn.Mutate(context.Background(), mu)
	if err != nil {
		log.Fatal(err)
	}

	w.Write(resp.Json)

}

// getProgram get one program by uid
func getProgram(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dateParam := chi.URLParam(r, "uid")

	query := getQuery(dateParam)
	dgClient := newClient()
	txn := dgClient.NewTxn()

	resp, err := txn.Query(context.Background(), query)
	if err != nil {
		//log.Fatal(err)
		w.Write([]byte("Error2!!"))
	}

	w.Write(resp.Json)
}

// runProgram executes the program
func runProgram(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var rawCode Code

	err := json.NewDecoder(r.Body).Decode(&rawCode)
	if err != nil {
		log.Fatal(err)
	}

	raw := []byte(rawCode.Code[0])

	err = os.WriteFile("archivoPrueba.py", raw, 0644)
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("C:\\Usuarios\\Carranza\\AppData\\Local\\Programs\\Python\\Python310\\python.exe", "./archivoPrueba.py")

	out, err := cmd.Output()
	if err != nil {
		err = json.NewEncoder(w).Encode("Syntax error")
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	json.NewEncoder(w).Encode(string(out))
}

// getProgram get one program by uid
func getProgramsPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dateParam := chi.URLParam(r, "uid")

	query := getQueryPagination(dateParam)
	dgClient := newClient()
	txn := dgClient.NewTxn()

	resp, err := txn.Query(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}

	w.Write(resp.Json)
}

//getQuery builds the query that executes by getPrograms and getProgram
func getQuery(uid string) string {
	return fmt.Sprintf(queryProgramByUid, uid)
}

//getQuery pagination get programs
func getQueryPagination(uid string) string {
	return fmt.Sprintf(queryPaginationPrograms, uid)
}
