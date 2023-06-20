package controllers

import (
	"net/http"
	"newproject/newproject/models"
	"newproject/newproject/util"
	"time"

	"github.com/fatih/structs"
	jsoniter "github.com/json-iterator/go"
)

type errorStruct struct {
	Error string `json:"error"`
}

func AddBeneficiary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	type Request struct {
		BeneficiaryName  string `json:"beneficiary_name"`
		BeneficiaryDOB   string `json:"beneficiary_dob"`
		BeneficiarySSN   string `json:"beneficiary_ssn"`
		BeneficiaryPhone string `json:"beneficiary_phone"`
	}

	var req Request

	err := jsoniter.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		util.Logger.Log(structs.Map(errorStruct{Error: err.Error()}))
		w.WriteHeader(http.StatusBadRequest)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err.Error()})
		return
	}

	dob, err := time.Parse("02-01-2006", req.BeneficiaryDOB)
	if err != nil {
		util.Logger.Log(structs.Map(errorStruct{Error: err.Error()}))
		w.WriteHeader(http.StatusBadRequest)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err.Error()})
		return
	}
	beneficiary := models.Beneficiary{Name: req.BeneficiaryName, DateOfBirth: dob, SocialSecurityNumber: req.BeneficiarySSN, Phone: req.BeneficiaryPhone}
	_, err = beneficiary.Add()
	if err != nil {
		util.Logger.Log(structs.Map(errorStruct{Error: err.Error()}))
		w.WriteHeader(http.StatusInternalServerError)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
}

func GetBeneficiary(w http.ResponseWriter, r *http.Request) {
	var beneficiary models.Beneficiary

	err := jsoniter.NewDecoder(r.Body).Decode(&beneficiary)
	if err != nil {
		util.Logger.Log(structs.Map(errorStruct{Error: err.Error()}))
		w.WriteHeader(http.StatusBadRequest)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err.Error()})
		return
	}
	err = beneficiary.Get()
	if err != nil {
		util.Logger.Log(structs.Map(errorStruct{Error: err.Error()}))
		w.WriteHeader(http.StatusInternalServerError)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	jsoniter.NewEncoder(w).Encode(beneficiary)
}

//Not needed for V.1
func GetBeneficiaries(w http.ResponseWriter, r *http.Request) {
	// w.Header().Set("Access-Control-Allow-Origin", "*")

	// type Request struct {
	// }

	// var req Request

	// err := jsoniter.NewDecoder(r.Body).Decode(&req)
	// if err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	jsoniter.NewEncoder(w).Encode(errorStruct{Error: err.Error()})
	// 	return
	// }
	// if true {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	jsoniter.NewEncoder(w).Encode("")
	// 	return
	// }

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
		util.Logger.Log(structs.Map(errorStruct{Error: err.Error()}))
		w.WriteHeader(http.StatusBadRequest)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err.Error()})
		return
	}

	appointDate, err := time.Parse("02-01-2006", req.AppointmentDate)
	if err != nil {
		util.Logger.Log(structs.Map(errorStruct{Error: err.Error()}))
		w.WriteHeader(http.StatusBadRequest)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err.Error()})
		return
	}
	err = models.AddFullAppointment(req.BeneficiaryID, appointDate, req.AppointmentTimeslot, req.Dose, req.AppointmentCenter)
	if err != nil {
		util.Logger.Log(structs.Map(errorStruct{Error: err.Error()}))
		w.WriteHeader(http.StatusInternalServerError)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
}

//This could have been implemented better, although this is named Appointment, we are really updating both
//appointment and beneficiary_appointment tables respectively.
//
//There is no mention of issues with missed appointments. This is an issue that is not addressed in the problem statement
//and technically breaks the framework of the problem. This will allow for reschedules on appointments that are pre-existing
//that you can avoid the 2 appointment framework but you'd have to assume the function calling already knows a slot that is
//passed current date is probably a missed appointment that needs to be reschedule or "updated"
func UpdateAppointment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	type Request struct {
		BeneficiaryID          int    `json:"beneficiary_id"`
		AppointmentDate        string `json:"appointment_date"`
		AppointmentTimeslot    string `json:"appointment_slot"`
		NewAppointmentDate     string `json:"new_appointment_date"`
		NewAppointmentTimeslot string `json:"new_appointment_slot"`
		NewAppointmentCenter   string `json:"new_appointment_center"`
	}

	var req Request

	err := jsoniter.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		util.Logger.Log(structs.Map(errorStruct{Error: err.Error()}))
		w.WriteHeader(http.StatusBadRequest)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err.Error()})
		return
	}

	//convert string date to time.Time
	appointDate, err := time.Parse("02-01-2006", req.AppointmentDate)
	newAppointDate, err := time.Parse("02-01-2006", req.NewAppointmentDate)

	err = models.UpdateFullAppointment(req.BeneficiaryID, appointDate, newAppointDate, req.AppointmentTimeslot, req.NewAppointmentTimeslot, req.NewAppointmentCenter)
	if err != nil {
		util.Logger.Log(structs.Map(errorStruct{Error: err.Error()}))
		w.WriteHeader(http.StatusBadRequest)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
}

//Delete via beneficiaryAppointment table
func DeleteAppointment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var beneficiaryAppointment models.BeneficiaryAppointment

	err := jsoniter.NewDecoder(r.Body).Decode(&beneficiaryAppointment)
	if err != nil {
		util.Logger.Log(structs.Map(errorStruct{Error: err.Error()}))
		w.WriteHeader(http.StatusBadRequest)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err.Error()})
		return
	}

	_, err = beneficiaryAppointment.Delete()
	if err != nil {
		util.Logger.Log(structs.Map(errorStruct{Error: err.Error()}))
		w.WriteHeader(http.StatusBadRequest)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err.Error()})
		return
	}

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
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err.Error()})
		return
	}

	appointDate, err := time.Parse("02-01-2006", req.AppointmentDate)
	if err != nil {
		util.Logger.Log(structs.Map(errorStruct{Error: err.Error()}))
		w.WriteHeader(http.StatusBadRequest)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err.Error()})
		return
	}
	appointment := models.Appointment{Date: appointDate, Timeslot: req.AppointmentTimeslot}
	err = appointment.Get()
	if err != nil {
		util.Logger.Log(structs.Map(errorStruct{Error: err.Error()}))
		w.WriteHeader(http.StatusInternalServerError)
		jsoniter.NewEncoder(w).Encode(errorStruct{Error: err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
	jsoniter.NewEncoder(w).Encode(appointment)
}

//no need to implement in V.1
func GetAppointments(w http.ResponseWriter, r *http.Request) {
	// w.Header().Set("Access-Control-Allow-Origin", "*")

	// type Request struct {
	// }

	// var req Request

	// jsoniter.NewDecoder(r.Body).Decode(&req)

	// if true {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	jsoniter.NewEncoder(w).Encode("")
	// 	return
	// }

	//Get Appointments
}
