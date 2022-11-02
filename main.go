package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"

	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	username = "root"
	password = "root"
	hostname = "db:3306"
	dbname   = "goapi"
)

type Transaction struct {
	TransactionId      int
	RequestId          int
	TerminalId         int
	PartnerObjectId    int
	AmountTotal        float64
	AmountOriginal     float64
	CommissionPS       float64
	CommissionClient   float64
	CommissionProvider float64
	DateInput          string
	DatePost           string
	Status             string
	PaymentType        string
	PaymentNumber      string
	ServiceId          int
	Service            string
	PayeeId            int
	PayeeName          string
	PayeeBankMfo       int
	PayeeBankAccount   string
	PaymentNarrative   string
}

func dsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s", username, password, hostname, dbname)
}

func dbConnection() (*sql.DB, error) {

	db, err := sql.Open("mysql", dsn())
	if err != nil {
		log.Printf("Error %s when opening DB", err)
		return nil, err
	}
	return db, nil
}

func createTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE transactions(
		TransactionId INT PRIMARY KEY,
		RequestId INT,
		TerminalId INT,
		PartnerObjectId INT,
		AmountTotal DOUBLE,
		AmountOriginal DOUBLE,
		CommissionPS DOUBLE,
		CommissionClient DOUBLE,
		CommissionProvider DOUBLE,
		DateInput DATETIME,
		DatePost DATETIME,
		Status VARCHAR(50),
		PaymentType VARCHAR(50),
		PaymentNumber VARCHAR(50),
		ServiceId INT,
		Service VARCHAR(50),
		PayeeId INT,
		PayeeName VARCHAR(50),
		PayeeBankMfo INT,
		PayeeBankAccount VARCHAR(50),
		PaymentNarrative VARCHAR(200))`)
	if err != nil {
		return err
	}
	return nil
}

func insert(db *sql.DB, allRecords [][]string) error {
	query := `INSERT INTO transactions(
		TransactionId,
		RequestId,
		TerminalId,
		PartnerObjectId,
		AmountTotal,
		AmountOriginal,
		CommissionPS,
		CommissionClient,
		CommissionProvider,
		DateInput,
		DatePost,
		Status,
		PaymentType,
		PaymentNumber,
		ServiceId,
		Service,
		PayeeId,
		PayeeName,
		PayeeBankMfo,
		PayeeBankAccount,
		PaymentNarrative) VALUES `

	var trans Transaction
	var inserts []string
	var params []interface{}

	for i := 1; i < len(allRecords); i++ {
		trans.TransactionId, _ = strconv.Atoi(allRecords[i][0])
		trans.RequestId, _ = strconv.Atoi(allRecords[i][1])
		trans.TerminalId, _ = strconv.Atoi(allRecords[i][2])
		trans.PartnerObjectId, _ = strconv.Atoi(allRecords[i][3])
		trans.AmountTotal, _ = strconv.ParseFloat(allRecords[i][4], 64)
		trans.AmountOriginal, _ = strconv.ParseFloat(allRecords[i][5], 64)
		trans.CommissionPS, _ = strconv.ParseFloat(allRecords[i][6], 64)
		trans.CommissionClient, _ = strconv.ParseFloat(allRecords[i][7], 64)
		trans.CommissionProvider, _ = strconv.ParseFloat(allRecords[i][8], 64)
		trans.DateInput = allRecords[i][9]
		trans.DatePost = allRecords[i][10]
		trans.Status = allRecords[i][11]
		trans.PaymentType = allRecords[i][12]
		trans.PaymentNumber = allRecords[i][13]
		trans.ServiceId, _ = strconv.Atoi(allRecords[i][14])
		trans.Service = allRecords[i][15]
		trans.PayeeId, _ = strconv.Atoi(allRecords[i][16])
		trans.PayeeName = allRecords[i][17]
		trans.PayeeBankMfo, _ = strconv.Atoi(allRecords[i][18])
		trans.PayeeBankAccount = allRecords[i][19]
		trans.PaymentNarrative = allRecords[i][20]

		params = append(params,
			trans.TransactionId,
			trans.RequestId,
			trans.TerminalId,
			trans.PartnerObjectId,
			trans.AmountTotal,
			trans.AmountOriginal,
			trans.CommissionPS,
			trans.CommissionClient,
			trans.CommissionProvider,
			trans.DateInput,
			trans.DatePost,
			trans.Status,
			trans.PaymentType,
			trans.PaymentNumber,
			trans.ServiceId,
			trans.Service,
			trans.PayeeId,
			trans.PayeeName,
			trans.PayeeBankMfo,
			trans.PayeeBankAccount,
			trans.PaymentNarrative)

		inserts = append(inserts, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	}

	queryVals := strings.Join(inserts, ",")
	query = query + queryVals

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, params...)
	if err != nil {
		log.Printf("Error %s when inserting row into transactions table", err)
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when finding rows affected", err)
		return err
	}
	log.Printf("%d transactions created simultaneously", rows)
	return nil
}

// обработка закачки
func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseMultipartForm(100 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			log.Println("An error encountered ::", err)
			return
		}
		defer file.Close()
		fmt.Fprintf(w, "%v", handler.Header)
		f, err := os.OpenFile(handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Println("An error encountered ::", err)
			return
		}
		defer f.Close()
		io.Copy(f, file)
		recordFile, err := os.Open(f.Name())
		if err != nil {
			log.Println("An error encountered ::", err)
			return
		}
		defer recordFile.Close()

		reader := csv.NewReader(recordFile)
		allRecords, err := reader.ReadAll()
		if err != nil {
			log.Println("An error encountered ::", err)
			return
		}
		log.Println("Successfully loaded file")
		defer f.Close()
		db, err := dbConnection()
		if err != nil {
			log.Printf("Error %s when getting db connection", err)
			return
		}
		defer db.Close()

		log.Printf("Successfully connected to database")

		err = createTable(db)
		if err != nil {
			log.Printf("Error %s during creating transactions table", err)
			return
		}

		err = insert(db, allRecords)
		if err != nil {
			log.Printf("Insert transaction failed with error %s", err)
			return
		}

	}

}

// Обработка GET запросов
func getData(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	sqlquery := ""

	transaction_id := query.Get("transaction_id")
	if transaction_id != "" {
		sqlquery = fmt.Sprintf("SELECT * FROM transactions WHERE TransactionId = %s", transaction_id)
	}

	status := query.Get("status")
	if status != "" {
		sqlquery = fmt.Sprintf("SELECT * FROM transactions WHERE Status = '%s'", status)
	}

	payment_type := query.Get("payment_type")
	if payment_type != "" {
		sqlquery = fmt.Sprintf("SELECT * FROM transactions WHERE PaymentType = '%s'", payment_type)
	}

	payment_narrative := query.Get("payment_narrative")
	if payment_narrative != "" {
		sqlquery = fmt.Sprintf("SELECT * FROM transactions WHERE PaymentNarrative LIKE '%%%s%%'", payment_narrative)
	}

	date_post := query.Get("date_post")
	if date_post != "" {
		dates := strings.Split(date_post, ",")
		sqlquery = fmt.Sprintf("SELECT * FROM transactions WHERE DatePost >= '%s 00:00:00' AND DatePost <= '%s 23:59:59'", dates[0], dates[1])
	}

	terminal_id := query.Get("terminal_id")
	if terminal_id != "" {
		sqlquery = fmt.Sprintf("SELECT * FROM transactions WHERE TerminalId IN (%s)", terminal_id)
	}

	db, err := dbConnection()
	if err != nil {
		log.Printf("Error %s when getting db connection", err)
		return
	}

	defer db.Close()

	log.Printf("Successfully connected to database")
	res, err := db.Query(sqlquery)

	if err != nil {
		log.Println("An error encountered ::", err)
		return
	}

	jsonresponse := []map[string]interface{}{}

	for res.Next() {
		var trans Transaction
		err := res.Scan(
			&trans.TransactionId,
			&trans.RequestId,
			&trans.TerminalId,
			&trans.PartnerObjectId,
			&trans.AmountTotal,
			&trans.AmountOriginal,
			&trans.CommissionPS,
			&trans.CommissionClient,
			&trans.CommissionProvider,
			&trans.DateInput,
			&trans.DatePost,
			&trans.Status,
			&trans.PaymentType,
			&trans.PaymentNumber,
			&trans.ServiceId,
			&trans.Service,
			&trans.PayeeId,
			&trans.PayeeName,
			&trans.PayeeBankMfo,
			&trans.PayeeBankAccount,
			&trans.PaymentNarrative)
		if err != nil {
			log.Println("An error encountered ::", err)
			return
		}

		v := reflect.ValueOf(trans)
		rowvalues := make(map[string]interface{}, v.NumField())
		for i := 0; i < v.NumField(); i++ {
			rowvalues[v.Type().Field(i).Name] = v.Field(i).Interface()
		}

		jsonresponse = append(jsonresponse, rowvalues)

	}
	defer res.Close()
	if err := json.NewEncoder(w).Encode(jsonresponse); err != nil {
		log.Println("An error encountered ::", err)
		return
	}

}

func main() {
	http.HandleFunc("/api/upload", upload)
	http.HandleFunc("/api/getdata", getData)
	log.Printf("Server started at port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
