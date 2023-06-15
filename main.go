package main

import (
	"fmt"
	"net/http"
	controllers "newproject/newproject/controllers"
	models "newproject/newproject/models"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	mux "github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {
	var log = logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.Out = os.Stdout
	models.ConnectDB()
	defer models.VaccineDB.Close()
	r := mux.NewRouter()
	r.HandleFunc("/beneficiary/addBeneficiary", controllers.AddBeneficiary)
	r.HandleFunc("/beneficiary/getBeneficaries", controllers.GetBeneficiaries)
	r.HandleFunc("/beneficiary/getBeneficary", controllers.GetBeneficiary)
	r.HandleFunc("/appointment/deleteAppointment", controllers.DeleteAppointment)
	r.HandleFunc("/appointment/addAppointment", controllers.AddAppointment)
	r.HandleFunc("/appointment/getAppointment", controllers.GetAppointment)
	r.HandleFunc("/appointment/updateAppointment", controllers.UpdateAppointment)
	r.HandleFunc("/appointment/getAppointments", controllers.GetAppointments)
	http.Handle("/", r)

	file, err := os.OpenFile("./logrus.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		fmt.Println("ded")
	} else {
		log.Out = file
	}
	log.WithFields(logrus.Fields{
		"fish": "wish",
		"true": "moo",
	}).Info("These Rhyme")
	log.Infof("info", "")
	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
