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
	BeneficiaryAppointmentID int
	BeneficiaryID            int
	Name                     string
	AppointmentID            int
	Date                     time.Time
	Timeslot                 string
	AppointmentCenter        string
	Dose                     string
}

//
func (beneficiaryAppointment *BeneficiaryAppointment) Get() error {
	switch {
	case beneficiaryAppointment.BeneficiaryAppointmentID != 0:
		query := fmt.Sprintf("SELECT * FROM %v WHERE beneficiary_appointment_id = ?", TableBeneficiaryAppointment)
		row := VaccineDB.QueryRow(query, beneficiaryAppointment.BeneficiaryAppointmentID)
		if row != nil {
			err := row.Scan(&beneficiaryAppointment.BeneficiaryAppointmentID, &beneficiaryAppointment.BeneficiaryID,
				&beneficiaryAppointment.AppointmentID, &beneficiaryAppointment.AppointmentCenter, &beneficiaryAppointment.Dose)
			if err != nil {
				fmt.Println(err)
				return err
			}
		}
	case beneficiaryAppointment.AppointmentID != 0 && beneficiaryAppointment.BeneficiaryID != 0:
		query := fmt.Sprintf("SELECT * FROM %v WHERE beneficiary_id = ? AND appointment_id = ?", TableBeneficiaryAppointment)
		row := VaccineDB.QueryRow(query, beneficiaryAppointment.BeneficiaryID, beneficiaryAppointment.AppointmentID)
		if row != nil {
			err := row.Scan(&beneficiaryAppointment.BeneficiaryAppointmentID, &beneficiaryAppointment.BeneficiaryID,
				&beneficiaryAppointment.AppointmentID, &beneficiaryAppointment.AppointmentCenter, &beneficiaryAppointment.Dose)
			if err != nil {
				fmt.Println(err)
				return err
			}
		}
	default:
		err := fmt.Errorf("invalid query data")
		return err
	}
	return nil
}

//this is a function strictly for the controller function and acts more as an upsert rather than a strict update
//
//only allows updates on values that would make sense updating via a customer perspective (timeslot, day, center) dose and
//beneficiary info should not change for exisiting appointments
//
func (beneficiaryAppointment *BeneficiaryAppointment) Update(toUpdate Appointment, updateCenter string) (bool, error) {

	if (toUpdate.Date == time.Time{} && toUpdate.Timeslot == "") || updateCenter == "" {
		err := fmt.Errorf("incomplete Appointment Info")
		return false, err
	}
	//verify BA exists
	err := beneficiaryAppointment.Get()
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	err = toUpdate.Get()
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			_, err := toUpdate.Add()
			if err != nil {
				return false, err
			}
			err = toUpdate.Get()
			if err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	}

	query := fmt.Sprintf("UPDATE %v SET appointment_id = ?, appointment_center = ? WHERE beneficiary_appointment_id = ?", TableBeneficiaryAppointment)
	_, err = VaccineDB.Query(query, toUpdate.ID, updateCenter, beneficiaryAppointment.BeneficiaryAppointmentID)
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	return true, nil
}

//simple delete function for Beneficiary Appointment on the autoincrement key
func (beneficiaryAppointment *BeneficiaryAppointment) Delete() (bool, error) {
	if beneficiaryAppointment.BeneficiaryAppointmentID == 0 {
		return false, fmt.Errorf("beneficiary_appointment_id is nil")
	}
	query := "DELETE FROM %v WHERE beneficiary_appointment_id = ?"
	_, err := VaccineDB.Query(query, beneficiaryAppointment.BeneficiaryAppointmentID)
	if err != nil {
		return false, err
	}
	return true, nil
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
	query := fmt.Sprintf("SELECT ba.beneficiary_appointment_id, bene.beneficiary_id, bene.beneficiary_name, app.appointment_id, app.appointment_date, app.appointment_slot, ba.appointment_center, ba.appointment_dose FROM %v ba JOIN %v app ON ba.appointment_id = app.appointment_id JOIN %v bene ON ba.beneficiary_id = bene.beneficiary_id WHERE ba.beneficiary_id = ?", TableBeneficiaryAppointment, TableAppointment, TableBeneficiary)
	rows, err := VaccineDB.Query(query, beneficiaryID)
	if err != nil {
		fmt.Println(err)
		return []AppointmentData{}, err
	}
	for rows.Next() {
		var appData AppointmentData
		err = rows.Scan(&appData.BeneficiaryAppointmentID, &appData.BeneficiaryID, &appData.Name, &appData.AppointmentID, &appData.Date, &appData.Timeslot, &appData.AppointmentCenter, &appData.Dose)
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

//function call to add all necessary appointment fields into the database
func AddFullAppointment(beneficiaryID int, appointmentDate time.Time, timeslot string, dose string, appointmentCenter string) error {
	appointment := Appointment{Date: appointmentDate, Timeslot: timeslot}
	err := appointment.Get()
	if err.Error() == "sql: no rows in result set" {
		_, err = appointment.Add()
		if err != nil {
			return err
		}
		err = appointment.Get()
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	beneficiary := Beneficiary{ID: beneficiaryID}
	err = beneficiary.Get()
	if err != nil {
		return err
	}

	//this function will return 0 rows if designated booking info is full. This function validates that the full critera block an
	//individual from booking a slot but does not explain which. Breaking this up into multiple checks may be the better way to go
	_, err = VerifyFreeAppointment(appointmentDate, timeslot, dose, appointmentCenter)
	if err != nil {
		return err
	}

	//This verifys the date constraints and that the user has not already had both doses. add unique constraint for
	//dose + beneficiary ID would probably be a good idea
	appList, err := GetFullAppointmentsByBeneficiary(beneficiaryID)
	if err != nil {
		return err
	}
	if len(appList) == 2 {
		err = fmt.Errorf("already recieved both shots")
		return err
	} else if len(appList) == 1 {
		if (appointmentDate.Sub(appList[0].Date).Hours() / 24) < 15 {
			err = fmt.Errorf("beneficiary is still in 15 day waiting period")
			return err
		}
	}
	beneficiaryAppointment := BeneficiaryAppointment{
		AppointmentID:     appointment.ID,
		BeneficiaryID:     beneficiaryID,
		AppointmentCenter: appointmentCenter,
		Dose:              dose,
	}

	_, err = beneficiaryAppointment.Add()
	if err != nil {
		return err
	}
	return nil
}

//this works but doesnt account for rollbacks on data if there ever is an issue in the transaction pipeline. If I had more time,
//I'd have added rollback transaction logic within this as well
func UpdateFullAppointment(beneficaryID int, appointmentDate time.Time, newAppointmentDate time.Time,
	timeslot string, newTimeslot string, newCenter string) error {
	beneficiary := Beneficiary{ID: beneficaryID}
	//verify beneficiary exists
	err := beneficiary.Get()
	if err != nil {
		return err
	}

	appts, err := GetFullAppointmentsByBeneficiary(beneficiary.ID)
	//verify appointment exists and is valid, if it does upsert
	if appts[len(appts)-1].Date != appointmentDate {
		err := fmt.Errorf("appointment time request does not match valid updateable appointment")
		return err
	}
	if len(appts) == 2 {
		if (newAppointmentDate.Sub(appts[0].Date).Hours() / 24) < 15 {
			err = fmt.Errorf("beneficiary is still in 15 day waiting period")
			return err
		}
	}

	beneficiaryAppointment := BeneficiaryAppointment{BeneficiaryAppointmentID: appts[0].BeneficiaryAppointmentID}

	_, err = beneficiaryAppointment.Update(Appointment{Date: newAppointmentDate, Timeslot: newTimeslot}, newCenter)
	if err != nil {
		return err
	}
	return nil
}
