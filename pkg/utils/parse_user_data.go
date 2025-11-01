package utils

import (
	"fmt"
	"log"
	"strconv"
)

// convertibleNumeric safely converts an interface{} value to float64, defaulting to 0.0.
func convertibleNumeric(v interface{}) float64 {
	if v == nil {
		return 0.0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case string:
		// Attempt to parse string to float
		f, err := strconv.ParseFloat(val, 64)
		if err == nil {
			return f
		}
	}
	return 0.0
}

// ConstructPercentData converts raw data map[string]interface{} based on dataKey,
// applying percentage conversion and renaming keys.
func ConstructPercentData(data []map[string]interface{}, dataKey, keyItem string) []map[string]interface{} {
	if len(data) == 0 {
		return nil
	}

	log.Printf("Constructing percent data for key %s", dataKey)
	result := make([]map[string]interface{}, 0, len(data))

	for _, item := range data {
		// Use helpers to safely retrieve and convert values
		rawValue := item["ratio"]
		key := fmt.Sprintf("%v", item[keyItem]) // Convert key to string, defaulting to empty string if not present

		// 1. Convert value to numeric
		value := convertibleNumeric(rawValue)

		// Create a mutable copy of the item
		resultItem := make(map[string]interface{})
		for k, v := range item {
			resultItem[k] = v
		}

		if dataKey != "content_interest" {
			resultItem["name"] = key // Set 'name' for non-content_interest types
		}

		// Assign the processed value back (unless deleted above)
		if _, exists := resultItem["value"]; exists || dataKey != "content_interest" {
			resultItem["value"] = value
		}

		// Delete 'key' regardless of dataKey
		delete(resultItem, "key")

		result = append(result, resultItem)
	}
	return result
}
