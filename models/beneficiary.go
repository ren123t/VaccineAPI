package models

import (
	"fmt"
	"time"
)

type Beneficiary struct {
	ID                   int       `json:"beneficiary_id"`
	Name                 string    `json:"beneficiary_name"`
	DateOfBirth          time.Time `json:"beneficiary_dob"`
	SocialSecurityNumber string    `json:"beneficiary_ssn"`
	Phone                string    `json:"beneficiary_phone"`
}

func (beneficiary *Beneficiary) Add() (bool, error) {
	if (time.Since(beneficiary.DateOfBirth).Hours() / 8760) < 45 {
		err := fmt.Errorf("user under 45 years of age")
		return false, err
	}
	query := fmt.Sprintf("INSERT INTO %v (beneficiary_name, beneficiary_dob, beneficiary_ssn, beneficiary_phone) VALUES (?, ?, ?, ?, ?)", TableBeneficiary)
	_, err := VaccineDB.Query(query, beneficiary.Name, beneficiary.DateOfBirth, beneficiary.SocialSecurityNumber, beneficiary.Phone)
	if err != nil {
		return false, err
	}
	return true, nil
}

//only allowing get functionality on the 2 unique keys on single row return
func (beneficiary *Beneficiary) Get() error {
	switch {
	case beneficiary.ID != 0:
		query := fmt.Sprintf("SELECT * FROM %v WHERE beneficiary_id = ?", TableBeneficiary)
		row := VaccineDB.QueryRow(query, beneficiary.ID)
		if row != nil {
			err := row.Scan(&beneficiary.ID, &beneficiary.Name, &beneficiary.DateOfBirth, &beneficiary.SocialSecurityNumber, &beneficiary.Phone)
			if err != nil {
				fmt.Println(err)
				return err
			}
		}
	case beneficiary.SocialSecurityNumber != "":
		query := fmt.Sprintf("SELECT * FROM %v WHERE beneficiary_ssn = ?", TableBeneficiary)
		row := VaccineDB.QueryRow(query, beneficiary.SocialSecurityNumber)
		if row != nil {
			err := row.Scan(&beneficiary.ID, &beneficiary.Name, &beneficiary.DateOfBirth, &beneficiary.SocialSecurityNumber, &beneficiary.Phone)
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
func GetBeneficiaries() ([]Beneficiary, error) {

	return nil, nil
}
