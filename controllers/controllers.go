package controllers

import (
	"fmt"
	"net/http"
	"newproject/newproject/models"
	"time"

	jsoniter "github.com/json-iterator/go"
)

type errorStruct struct {
	Error error `json:"error"`
}

func AddBeneficiary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	type Request struct {
		BeneficiaryID    int    `json:"beneficiary_id"`
		BeneficiaryName  string `json:"beneficiary_name"`
		BeneficiaryDOB   string `json:"beneficiary_dob"`
		BeneficiarySSN   string `json:"beneficiary_ssn"`
		BeneficiaryPhone string `json:"beneficiary_phone"`
	}

	var req Request

	err := jsoniter.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err})
		return
	}

	dob, err := time.Parse("02-01-2006", req.BeneficiaryDOB)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err})
		return
	}
	beneficiary := models.Beneficiary{Name: req.BeneficiaryName, DateOfBirth: dob, SocialSecurityNumber: req.BeneficiarySSN, Phone: req.BeneficiaryPhone}
	_, err = beneficiary.Add()
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err})
		return
	}
	w.WriteHeader(http.StatusOK)
}

func GetBeneficiary(w http.ResponseWriter, r *http.Request) {
	var beneficiary models.Beneficiary

	jsoniter.NewDecoder(r.Body).Decode(&beneficiary)

	if true {
		w.WriteHeader(http.StatusBadRequest)
		jsoniter.NewEncoder(w).Encode("")
	}

	beneficiary.Get()
}

func GetBeneficiaries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	type Request struct {
	}

	var req Request

	jsoniter.NewDecoder(r.Body).Decode(&req)

	if true {
		w.WriteHeader(http.StatusBadRequest)
		jsoniter.NewEncoder(w).Encode("")
	}

	//GetBeneficiaries()
}

// this add is actually adding both appointment in the appointment table if it does not have it
//as well as appointment information in the joining table
func AddAppointment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	type Request struct {
		BeneficiaryID       int    `json:"beneficiary_id"`
		AppointmentDate     string `json:"appointment_date"`
		AppointmentTimeslot string `json:"appointment_slot"`
		Dose                string `json:"appointment_dose"`
		AppointmentCenter   string `json:"appointment_center"`
	}

	var req Request

	err := jsoniter.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err})
		return
	}

	appointDate, err := time.Parse("02-01-2006", req.AppointmentDate)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err})
		return
	}
	appointment := models.Appointment{Date: appointDate, Timeslot: req.AppointmentTimeslot}
	err = appointment.Get()
	if err.Error() == "sql: no rows in result set" {
		_, err = appointment.Add()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsoniter.NewEncoder(w).Encode(errorStruct{Error: err})
			return
		}
		err = appointment.Get()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsoniter.NewEncoder(w).Encode(errorStruct{Error: err})
			return
		}
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err})
		return
	}

	beneficiary := models.Beneficiary{ID: req.BeneficiaryID}
	err = beneficiary.Get()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err})
		return
	}

	//this function will return 0 rows if designated booking info is full. This function validates that the full critera block an
	//individual from booking a slot but does not explain which. Breaking this up into multiple checks may be the better way to go
	_, err = models.VerifyFreeAppointment(appointDate, req.AppointmentTimeslot, req.Dose, req.AppointmentCenter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err})
		return
	}

	//This verifys the date constraints and that the user has not already had both doses. add unique constraint for
	//dose + beneficiary ID would probably be a good idea
	appList, err := models.GetFullAppointmentsByBeneficiary(req.BeneficiaryID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err})
		return
	}
	if len(appList) == 2 {
		err = fmt.Errorf("already recieved both shots")
		w.WriteHeader(http.StatusInternalServerError)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err})
		return
	} else if len(appList) == 1 {
		if (appointDate.Sub(appList[0].Date).Hours() / 24) < 15 {
			err = fmt.Errorf("beneficiary is still in 15 day waiting period")
			w.WriteHeader(http.StatusInternalServerError)
			jsoniter.NewEncoder(w).Encode(errorStruct{Error: err})
			return
		}
	}
	beneficiaryAppointment := models.BeneficiaryAppointment{
		AppointmentID:     appointment.ID,
		BeneficiaryID:     req.BeneficiaryID,
		AppointmentCenter: req.AppointmentCenter,
		Dose:              req.Dose,
	}

	_, err = beneficiaryAppointment.Add()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err})
		return
	}
	w.WriteHeader(http.StatusOK)
}

//This could have been implemented better, although this is named Appointment, we are really updating both
//appointment and beneficiary_appointment tables respectively.
func UpdateAppointment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	type Request struct {
		BeneficiaryID          int    `json:"beneficiary_id"`
		AppointmentDate        string `json:"appointment_date"`
		AppointmentTimeslot    string `json:"appointment_slot"`
		NewAppointmentDate     string `json:"new_appointment_date"`
		NewAppointmentTimeslot string `json:"new_appointment_timeslot"`
	}

	var req Request

	err := jsoniter.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err})
	}

	beneficiary := models.Beneficiary{ID: req.BeneficiaryID}
	//verify beneficiary exists
	err = beneficiary.Get()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err})
	}
	newAppointDate, err := time.Parse("02-01-2006", req.NewAppointmentDate)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err})
		return
	}

	appts, err := models.GetFullAppointmentsByBeneficiary(beneficiary.ID)
	if len(appts) == 2 {
		err = fmt.Errorf("already recieved both shots")
		w.WriteHeader(http.StatusInternalServerError)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err})
		return
	} else if len(appts) == 1 {
		if (newAppointDate.Sub(appts[0].Date).Hours() / 24) < 15 {
			err = fmt.Errorf("beneficiary is still in 15 day waiting period")
			w.WriteHeader(http.StatusInternalServerError)
			jsoniter.NewEncoder(w).Encode(errorStruct{Error: err})
			return
		}
	}

	//BeneficiaryAppointment.Update()

	w.WriteHeader(http.StatusOK)
}

//Delete via beneficiaryAppointment table
func DeleteAppointment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var appointment models.Appointment

	err := jsoniter.NewDecoder(r.Body).Decode(&appointment)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err})
	}

	appointment.Delete()
	w.WriteHeader(http.StatusOK)
}

//This is only for id, but can be expanded to date + timeslot
func GetAppointment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	type Request struct {
		BeneficiaryID       int    `json:"beneficiary_id"`
		AppointmentDate     string `json:"appointment_date"`
		AppointmentTimeslot string `json:"appointment_slot"`
	}

	var req Request

	err := jsoniter.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err})
		return
	}

	appointDate, err := time.Parse("02-01-2006", req.AppointmentDate)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err})
		return
	}
	appointment := models.Appointment{Date: appointDate, Timeslot: req.AppointmentTimeslot}
	err = appointment.Get()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err})
		return
	}
	w.WriteHeader(http.StatusOK)
	jsoniter.NewEncoder(w).Encode(appointment)
}

func GetAppointments(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	type Request struct {
	}

	var req Request

	jsoniter.NewDecoder(r.Body).Decode(&req)

	if true {
		w.WriteHeader(http.StatusBadRequest)
		jsoniter.NewEncoder(w).Encode("")
	}

	//Get Appointments
}
