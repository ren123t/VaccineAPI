package models

import (
	"fmt"
	"time"
)

//likely there wont really be crud on this model available to endpoints
//this is mostly here for helper calls and queries
type BeneficiaryAppointment struct {
	BeneficiaryAppointmentID int    `json:"beneficiary_appointment_id"`
	BeneficiaryID            int    `json:"beneficiary_id"`
	AppointmentID            int    `json:"appointment_id"`
	AppointmentCenter        string `json:"appointment_center"`
	Dose                     string `json:"appointment_dose"`
}

//struct for formatted return of relevant appointment data
type AppointmentData struct {
	BeneficiaryID     int
	Name              string
	AppointmentID     int
	Date              time.Time
	Timeslot          string
	AppointmentCenter string
	Dose              string
}

//
func (beneficiaryAppointment *BeneficiaryAppointment) Get() (BeneficiaryAppointment, error) {

	return *beneficiaryAppointment, nil
}

func (beneficiaryAppointment *BeneficiaryAppointment) Update() (bool, error) {

	return false, nil
}

func (beneficiaryAppointment *BeneficiaryAppointment) Delete() (bool, error) {

	return false, nil
}

func (beneficiaryAppointment *BeneficiaryAppointment) Add() (bool, error) {
	query := fmt.Sprintf("INSERT INTO %v (beneficiary_id, appointment_id, appointment_center, appointment_dose) VALUES (?, ?, ?, ?)", TableBeneficiary)
	_, err := VaccineDB.Query(query,
		beneficiaryAppointment.BeneficiaryID,
		beneficiaryAppointment.AppointmentID,
		beneficiaryAppointment.AppointmentCenter,
		beneficiaryAppointment.Dose)
	if err != nil {
		return false, err
	}

	return true, nil
}

//making this an eventual generic function would be much better
func GetFullAppointmentsByBeneficiary(beneficiaryID int) ([]AppointmentData, error) {
	listAppData := []AppointmentData{}
	query := fmt.Sprintf("SELECT bene.beneficiary_id, bene.beneficiary_name, app.appointment_id, app.appointment_date, app.appointment_slot, ba.appointment_center, ba.appointment_dose FROM %v ba JOIN %v app ON ba.appointment_id = app.appointment_id JOIN %v bene ON ba.beneficiary_id = bene.beneficiary_id WHERE ba.beneficiary_id = ?", TableBeneficiaryAppointment, TableAppointment, TableBeneficiary)
	rows, err := VaccineDB.Query(query, beneficiaryID)
	if err != nil {
		fmt.Println(err)
		return []AppointmentData{}, err
	}
	for rows.Next() {
		var appData AppointmentData
		err = rows.Scan(&appData.BeneficiaryID, &appData.Name, &appData.AppointmentID, &appData.Date, &appData.Timeslot, &appData.AppointmentCenter, &appData.Dose)
		if err != nil {
			return []AppointmentData{}, err
		}
		listAppData = append(listAppData, appData)
	}

	return listAppData, nil
}

//if this query returns a row, that means the slot is not filled up. this checks for constraints on both 15 per dose per center
//it also checks for maximum of 10 per slot per day
//this query assumes that appointment FK exists and is not used otherwise
func VerifyFreeAppointment(appointmentDate time.Time, appointmentSlot string, dose string, appointmentCenter string) ([]BeneficiaryAppointment, error) {
	query := fmt.Sprintf("SELECT * FROM %v ba JOIN %v app ON ba.appointment_id = app.appointment_id WHERE (SELECT COUNT(*) FROM %v WHERE app.appointment_date = ? AND app.appointment_slot = ?) < 11 AND (SELECT COUNT(*) FROM %v WHERE app.appointment_date = ? AND ba.appointment_center = ? AND ba.appointment_dose = ?) < 16",
		TableBeneficiaryAppointment, TableAppointment, TableBeneficiaryAppointment, TableBeneficiaryAppointment)

	rows, err := VaccineDB.Query(query, appointmentDate, appointmentSlot, appointmentDate, appointmentCenter, dose)
	if err != nil {
		fmt.Println(err)
	}
	listBAs := []BeneficiaryAppointment{}
	for rows.Next() {
		var ba BeneficiaryAppointment
		err = rows.Scan(&ba.AppointmentID, &ba.BeneficiaryID, &ba.AppointmentCenter, &ba.Dose)
		if err != nil {

		}
		listBAs = append(listBAs, ba)
	}
	return listBAs, nil
}

//this is something that will very likely be needed. will return complete appointment data for GUI. can be done with joins
func GetFullAppointment(beneficiaryAppointmentID int) AppointmentData {
	return AppointmentData{}
}
