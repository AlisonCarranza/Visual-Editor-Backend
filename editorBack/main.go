package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/dgraph-io/dgo/v210/protos/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

//Code is the program code structure
type Code struct {
	Code []string `json:"code"`
	Uid  string   `json:"uid"`
}

//Err is the error's response structure
type Err struct {
	Message   string `json:"message"`
	CodeError int    `json:"codeError"`
}

func main() {
	loadEnv()
	port := os.Getenv("PORT")
	host := os.Getenv("HOST")

	r := chi.NewRouter()

	r.Use(middleware.Logger,
		middleware.Recoverer,
	)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{host},
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
	r.Get("/programs-page/{uid}", getProgramsPage)

	//save programs
	r.Post("/programs", addProgram)

	//get program
	r.Get("/programs/{uid}", getProgram)

	//run program
	r.Post("/program/run", runProgram)

	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		fmt.Printf("Error listen port: %s", err)
	}
}

//loadEnv load environment variables
func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error load file .env: %s", err)
	}
}

// getPrograms get all the stored programs in the database
func getPrograms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	dgClient := newClient()
	txn := dgClient.NewTxn()

	resp, err := txn.Query(context.Background(), queryAllPrograms)
	if err != nil {
		HTTPError(w, r, http.StatusNotFound, err.Error(), 1)
		return
	}

	JSON(w, r, http.StatusOK, nil)
	w.Write(resp.Json)

}

//addProgram add the program in the database
func addProgram(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var rawCode Code

	err := json.NewDecoder(r.Body).Decode(&rawCode)
	if err != nil {
		HTTPError(w, r, http.StatusBadRequest, err.Error(), 2)
		return
	}

	p := Code{Code: rawCode.Code}

	pb, err := json.Marshal(p)
	if err != nil {
		HTTPError(w, r, http.StatusBadRequest, err.Error(), 3)
		return
	}

	dgClient := newClient()
	txn := dgClient.NewTxn()

	mu := &api.Mutation{
		CommitNow: true,
		SetJson:   pb,
	}

	resp, err := txn.Mutate(context.Background(), mu)
	if err != nil {
		HTTPError(w, r, http.StatusBadRequest, err.Error(), 4)
		return
	}

	JSON(w, r, http.StatusCreated, nil)
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
		HTTPError(w, r, http.StatusNotFound, err.Error(), 5)
		return
	}

	JSON(w, r, http.StatusOK, nil)
	w.Write(resp.Json)
}

// getProgram get a number of programs from a certain uid
func getProgramsPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dateParam := chi.URLParam(r, "uid")

	query := getQueryPagination(dateParam)
	dgClient := newClient()
	txn := dgClient.NewTxn()

	resp, err := txn.Query(context.Background(), query)
	if err != nil {
		HTTPError(w, r, http.StatusNotFound, err.Error(), 6)
		return
	}

	JSON(w, r, http.StatusOK, nil)
	w.Write(resp.Json)
}

// runProgram executes the program
func runProgram(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var rawCode Code

	err := json.NewDecoder(r.Body).Decode(&rawCode)
	if err != nil {
		HTTPError(w, r, http.StatusBadRequest, err.Error(), 7)
		return
	}

	raw := []byte(rawCode.Code[0])

	loadEnv()
	file := os.Getenv("FILE_TMP")
	cmdUrl := os.Getenv("CMD_RUN")
	dir := os.Getenv("DIR_TMP")

	f, err := os.CreateTemp(dir, file+".*.py")
	if err != nil {
		HTTPError(w, r, http.StatusNotFound, err.Error(), 1)
		return
	}

	err = os.WriteFile(f.Name(), raw, 0644)
	if err != nil {
		HTTPError(w, r, http.StatusBadRequest, err.Error(), 8)
		return
	}

	cmd := exec.Command(cmdUrl, f.Name())

	out, err := cmd.Output()

	if err != nil {
		err = json.NewEncoder(w).Encode("Syntax error")
		if err != nil {
			HTTPError(w, r, http.StatusBadRequest, err.Error(), 9)
			return
		}
		f.Close()
		defer os.Remove(f.Name())
		JSON(w, r, http.StatusOK, nil)
		return
	}

	f.Close()
	defer os.Remove(f.Name())
	JSON(w, r, http.StatusOK, nil)
	json.NewEncoder(w).Encode(string(out))
}

//getQuery builds the query that executes by getPrograms and getProgram
func getQuery(uid string) string {
	return fmt.Sprintf(queryProgramByUid, uid)
}

//getQuery builds the query that executes by getProgramsPage
func getQueryPagination(uid string) string {
	return fmt.Sprintf(queryPaginationPrograms, uid)
}

//HTTPError create a new error type Err
func HTTPError(w http.ResponseWriter, r *http.Request, statusCode int, message string, codeError int) error {
	error := Err{
		Message:   message,
		CodeError: codeError,
	}

	return JSON(w, r, statusCode, error)
}

//JSON Serialize response in json format
func JSON(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) error {
	if data == nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(statusCode)
		return nil
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write(bytes)
	return nil
}
