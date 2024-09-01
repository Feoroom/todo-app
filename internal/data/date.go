package data

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

var ErrInvalidDateFormat = errors.New("invalid date format")

//type Date time.Time

const (
	layout = "2006-01-02"
)

type Date struct {
	time.Time
}

func (t Date) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("%s", t.Format(layout))
	qStamp := strconv.Quote(stamp)
	return []byte(qStamp), nil
}

func (t *Date) UnmarshalJSON(jsonVal []byte) error {

	unquoteVal, err := strconv.Unquote(string(jsonVal))
	if err != nil {
		return ErrInvalidDateFormat
	}

	date, err := time.Parse(layout, unquoteVal)
	if err != nil {
		return ErrInvalidDateFormat
	}

	t.Time = date

	return nil
}

func (t Date) CheckYear() bool {
	return t.Year() >= time.Now().Year()
}

func (t Date) CheckMonth() bool {

	// 2024 == 2024
	// 7 < 10 Ok
	// 2024 < 2025
	// month - любой
	if t.Year() == time.Now().Year() {
		return t.Month() >= time.Now().Month()
	}

	return true
}

func (t Date) CheckDay() bool {

	if t.Year() == time.Now().Year() && t.Month() == time.Now().Month() {
		return t.Day() >= time.Now().Day()
	}

	return true
}
