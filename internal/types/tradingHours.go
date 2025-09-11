package types

import (
	"fmt"
	"time"
)

type TradingHours struct {
	OpenHour    int
	OpenMinute  int
	CloseHour   int
	CloseMinute int
	WeekendsOff bool
	ExchangeTZ  string
}

func (th *TradingHours) IsOpenAt(t time.Time) bool {
	if th.WeekendsOff && (t.Weekday() == time.Saturday || t.Weekday() == time.Sunday) {
		return false
	}
	openTime, closeTime, err := th.GetTradingHours()
	if err != nil {
		panic(fmt.Errorf("failed to get trading hours: %w", err))
	}
	return !t.Before(openTime) && !t.After(closeTime)
}

func (th *TradingHours) GetTradingHours() (time.Time, time.Time, error) {
	loc, err := time.LoadLocation(th.ExchangeTZ)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to load timezone: %w", err)
	}
	now := time.Now().UTC()
	closingTime := time.Date(now.Year(), now.Month(), now.Day(), th.CloseHour, th.CloseMinute, 0, 0, loc).UTC()
	openingTime := time.Date(now.Year(), now.Month(), now.Day(), th.OpenHour, th.OpenMinute, 0, 0, loc).UTC()
	return openingTime, closingTime, nil
}

func (th *TradingHours) getNextOpenTime(t time.Time) time.Time {
	if th.WeekendsOff && (t.Weekday() == time.Saturday || t.Weekday() == time.Sunday) {
		t = shiftToWorkingDay(t)
	}
	openTime, _, err := th.GetTradingHours()
	if err != nil {
		panic(fmt.Errorf("failed to get trading hours: %w", err))
	}
	if t.Before(openTime) {
		return openTime
	}
	t = openTime.AddDate(0, 0, 1)
	if th.WeekendsOff {
		return shiftToWorkingDay(t)
	}
	return t
}
