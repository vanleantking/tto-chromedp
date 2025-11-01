package utils

import (
	"fmt"
	"time"
)

func FormatDatetime(input string) (int64, error) {
	dateString := "20250801"
	// Go's reference time: Mon Jan 2 15:04:05 MST 2006
	// For YYYYMMDD (20250801), the layout is 20060102
	layout := "20060102"

	// Load the Ho Chi Minh timezone location.
	hcmLocation, err := time.LoadLocation(VN_TIMEZONE)
	if err != nil {
		return 0, fmt.Errorf("Error loading location 'Asia/Ho_Chi_Minh': %v", err)
	}

	t, err := time.Parse(layout, dateString)
	if err != nil {
		return 0, fmt.Errorf("Error parsing date: %v", err)
	}

	// If you specifically need the Unix timestamp (seconds since epoch):
	return t.In(hcmLocation).Unix(), nil
}

func TruncateToDate(inputTimestamp int64) (int64, error) {
	// Load the Ho Chi Minh timezone location.
	hcmLocation, err := time.LoadLocation(VN_TIMEZONE)
	if err != nil {
		return 0, fmt.Errorf("Error loading location 'Asia/Ho_Chi_Minh': %v", err)
	}
	// 2. Convert the Unix timestamp (seconds) to a time.Time object
	// Note: We use Local location for accurate date truncation based on local timezone rules.
	originalTime := time.Unix(inputTimestamp, 0).In(hcmLocation)
	fmt.Println("\nOriginal Time:", originalTime.Format(time.RFC1123Z))

	// 3. Truncate the time to the start of the day (midnight)
	midnightTime := originalTime.Truncate(24 * time.Hour)

	// 4. Get the new Unix timestamp for midnight
	return midnightTime.Unix(), nil
}
