package models

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var VaccineDB *sql.DB

const TableAppointment = "appointment"
const TableBeneficiary = "beneficiary"
const TableBeneficiaryAppointment = "beneficiary_appointment"

func ConnectDB() {
	VaccineDB, _ = sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/vaccinedb?parseTime=true")

	// See "Important settings" section.
	VaccineDB.SetConnMaxLifetime(time.Minute * 3)
	VaccineDB.SetMaxOpenConns(10)
	VaccineDB.SetMaxIdleConns(10)

	err := VaccineDB.Ping()

	b := Beneficiary{ID: 10}
	if err != nil {
		log.Fatal(err)
	}
	err = b.Get()
	fmt.Println(b, "  ", err)
}

//returns column name from struct tag
func printFieldColumn(obj interface{}, field string) string {
	retField, found := reflect.TypeOf(obj).FieldByName(field)
	if !found {
		fmt.Println("not found")
		return ""
	}

	column := retField.Tag.Get("column")
	return column
}
