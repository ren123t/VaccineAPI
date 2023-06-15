package models

import (
	"fmt"
	"time"
)

type Appointment struct {
	ID       int       `json:"appointment_id" column:"appointment_id"`
	Date     time.Time `json:"appointment_date" column:"appointment_date"`
	Timeslot string    `json:"appointment_slot" column:"appointment_slot"`
}

func (appointment *Appointment) Add() (bool, error) {
	if appointment.Date.Sub(time.Now()).Hours()/24 > 90 {
		return false, fmt.Errorf("cannot schedule past 90 days")
	}
	query := fmt.Sprintf("INSERT INTO %v (appointment_date, appointment_slot) VALUES (?, ?)", TableAppointment)
	_, err := VaccineDB.Query(query, appointment.Date, appointment.Timeslot)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	return true, nil
}

func (appointment *Appointment) Delete() (bool, error) {
	rows, err := VaccineDB.Query(fmt.Sprintf("DELETE FROM %v WHERE appointment_id = ?", TableAppointment), appointment.ID)
	fmt.Println(rows, " ", err)
	return false, nil
}

//this is only for id, can be made more robust with column tags and printFieldColumn function
func (appointment *Appointment) Get() error {
	switch {
	case appointment.ID != 0:
		query := fmt.Sprintf("SELECT * FROM %v WHERE appointment_id = ?", TableAppointment)
		row := VaccineDB.QueryRow(query, appointment.ID)
		if row != nil {
			err := row.Scan(&appointment.ID, &appointment.Date, &appointment.Timeslot)
			if err != nil {
				fmt.Println(err)
				return err
			}
		}
	case appointment.Date != time.Time{} && appointment.Timeslot != "":
		query := fmt.Sprintf("SELECT * FROM %v WHERE appointment_date = ? AND appointment_slot = ?", TableAppointment)
		row := VaccineDB.QueryRow(query, appointment.Date, appointment.Timeslot)
		if row != nil {
			err := row.Scan(&appointment.ID, &appointment.Date, &appointment.Timeslot)
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

//future release, will need to allow appending clauses with a query builder while only allowing read only clauses
func GetAppointments() ([]Appointment, error) {
	return nil, nil
}
