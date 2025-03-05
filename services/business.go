package services

import (
	"errors"
	"fmt"
	"github.com/kmakasheva/todo-list-project/domain"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	layout := "20060102"
	now = time.Date(2024, time.January, 26, 0, 0, 0, 0, time.UTC)
	dateTime, err := time.Parse(layout, date)
	var nextDate time.Time
	if err != nil {
		return "", fmt.Errorf("error while parsing date %w", err)
	}

	if repeat == "" || now.Format(layout) == "" || date == "" {
		return "", errors.New("make sure you have a valid data")
	}

	if repeat[0] != 'd' && repeat[0] != 'y' && repeat[0] != 'w' && repeat[0] != 'm' {
		return "", errors.New("the repeat enter is incorrect")
	}

	if repeat == "y" {
		nextDate = dateTime.AddDate(1, 0, 0)
		for nextDate.Format(layout) < now.Format(layout) {
			nextDate = nextDate.AddDate(1, 0, 0)
		}
		return nextDate.Format(layout), nil
	}

	value := strings.Split(repeat, " ")

	if value[0] == "d" {
		if len(value) == 2 {
			dNumber, err := strconv.Atoi(value[1])
			if err != nil {
				return "", fmt.Errorf("error while converting str to int %w", err)
			}
			if dNumber <= 0 || dNumber > 366 {
				return "", errors.New("превышен максимально допустимый интервал")
			}
			nextDate = dateTime.AddDate(0, 0, dNumber)
			nextDateStr := nextDate.Format(layout)
			for nextDateStr <= now.Format(layout) {
				dateTime = dateTime.AddDate(0, 0, dNumber)
				nextDateStr = dateTime.Format(layout)
			}
			return nextDateStr, nil
		} else {
			return "", errors.New("enter should be in format 'd' days_number")
		}
	}

	if value[0] == "w" {
		if len(value) == 2 {
			var res []time.Time
			var min time.Time
			nextDate = dateTime
			daysOfWeek := strings.Split(value[1], ",")
			for i := 0; i < len(daysOfWeek); i++ {
				dayOfWeekInt, err := strconv.Atoi(daysOfWeek[i])
				if err != nil {
					return "", fmt.Errorf("error while converting str to int in days of week %w", err)
				}
				if dayOfWeekInt >= 1 && dayOfWeekInt <= 7 {
					if now.Format(layout) > dateTime.Format(layout) {
						for now.Format(layout) > nextDate.Format(layout) {
							nextDate = nextDate.AddDate(0, 0, 1)
						}
						for int(nextDate.Weekday())+1 != dayOfWeekInt {
							nextDate = nextDate.AddDate(0, 0, 1)
						}
						if int(nextDate.Weekday())+1 == dayOfWeekInt {
							res = append(res, nextDate)
						}
					} else {
						for int(nextDate.Weekday())+1 != dayOfWeekInt {
							nextDate = nextDate.AddDate(0, 0, 1)
						}
						if int(nextDate.Weekday())+1 == dayOfWeekInt {
							res = append(res, nextDate)
							nextDate = dateTime
						}
					}
					min = res[0]
					for _, v := range res {
						if v.Format(layout) < min.Format(layout) {
							min = v
						}
					}
				} else {
					return "", errors.New("make sure you have a valid date for week")
				}
			}
			nextDate = min.AddDate(0, 0, 1)
			return nextDate.Format(layout), nil

		} else {
			return "", errors.New("enter should be in format 'w' weeks number")
		}
	}

	if value[0] == "m" {
		if dateTime.Format(layout) < now.Format(layout) {
			dateTime = now
		}
		var res []time.Time
		var min time.Time
		if len(value) == 2 {
			nextDate = dateTime.AddDate(0, 0, 1)
			daysOfMonth := strings.Split(value[1], ",")
			for _, dayOfMonth := range daysOfMonth {
				dayOfMonthInt, err := strconv.Atoi(dayOfMonth)
				if err != nil || dayOfMonthInt < -2 || dayOfMonthInt > 31 {
					return "", fmt.Errorf("enter should be in format 'm' days_number: %w", err)
				}
				if dayOfMonthInt == -1 {
					nextDate = nextDate.AddDate(0, 1, 0)
					nextDate = nextDate.AddDate(0, 0, -nextDate.Day())
					dayOfMonthInt = nextDate.Day()
				} else if dayOfMonthInt == -2 {
					nextDate = nextDate.AddDate(0, 1, 0)
					nextDate = nextDate.AddDate(0, 0, -nextDate.Day()-1)
					dayOfMonthInt = nextDate.Day()
				}
				for nextDate.Day() != dayOfMonthInt {
					nextDate = nextDate.AddDate(0, 0, 1)
				}
				if nextDate.Day() == dayOfMonthInt {
					res = append(res, nextDate)
				}
				nextDate = dateTime.AddDate(0, 0, 1)
			}
			min = res[0]
			for _, v := range res {
				if min.Format(layout) > v.Format(layout) {
					min = v
				}
			}
			return min.Format(layout), nil
		}
		if len(value) == 3 {
			nextDate = dateTime.AddDate(0, 0, 1)
			daysOfMonth := strings.Split(value[1], ",")
			months := strings.Split(value[2], ",")
			for _, month := range months {
				for _, dayOfMonth := range daysOfMonth {
					monthInt, err := strconv.Atoi(month)
					if err != nil || monthInt <= 0 || monthInt > 12 {
						return "", fmt.Errorf("make sure you have entered correct months: %w", err)
					}
					dayOfMonthInt, err := strconv.Atoi(dayOfMonth)
					if err != nil {
						return "", fmt.Errorf("make sure you have entered correct days of months: %w", err)
					}
					for int(nextDate.Month()) != monthInt {
						nextDate = nextDate.AddDate(0, 1, 0)
					}
					if int(nextDate.Month()) == monthInt {
						for nextDate.Day() != dayOfMonthInt {
							nextDate = nextDate.AddDate(0, 0, 1)
						}
						if nextDate.Day() == dayOfMonthInt && int(nextDate.Month()) == monthInt && nextDate.Format(layout) > now.Format(layout) {
							res = append(res, nextDate)
							nextDate = dateTime.AddDate(0, 0, 1)
						} else if nextDate.Day() == dayOfMonthInt && int(nextDate.Month())-1 == monthInt &&
							nextDate.AddDate(0, -1, 0).Format(layout) > now.Format(layout) {
							res = append(res, nextDate.AddDate(0, -1, 0))
							nextDate = dateTime.AddDate(0, 0, 1)
						}
					}
				}
			}
			min = res[0]
			for _, v := range res {
				if v.Format(layout) < min.Format(layout) {
					min = v
				}
			}
			return min.Format(layout), nil
		} else if len(value) != 2 && len(value) != 3 {
			return "", errors.New("enter should be in format 'm' days_number month_number")
		}
	}

	return "", nil
}

func ValidateTask(t domain.Task) error {
	id, err := strconv.Atoi(t.ID)
	if err != nil {
		return errors.New("id seems problematic")
	}
	if id < 0 {
		return errors.New("ID is not valid")
	}
	if _, err := time.Parse("20060102", t.Date); err != nil {
		return errors.New("date is not valid")
	}
	if t.Title == "" {
		return errors.New("title cannot be empty")
	}
	if t.Repeat == "" {
		return errors.New("repeat is empty")
	}
	if t.Repeat[0] != 'y' && t.Repeat[0] != 'm' && t.Repeat[0] != 'w' && t.Repeat[0] != 'd' {
		return errors.New("invalid repeat format")
	}
	row := strings.Split(t.Repeat, " ")
	if row[0] == "y" && len(row) != 1 {
		return errors.New("only one year repeat is available")
	}
	if row[0] == "m" && len(row) > 3 {
		return errors.New("incorrect repeat in months")
	}
	if row[0] == "w" && len(row) != 2 {
		return errors.New("repeat in weeks is not valid")
	}
	if row[0] == "d" && len(row) != 2 {
		return errors.New("repeat in days is not valid")
	}
	return nil
}
