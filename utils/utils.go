package utils

import (
	"io"
	"fmt"
	"time"
	"regexp"
	"strings"
	"strconv"
	"io/ioutil"
	"encoding/json"
)

func GetBody(data io.ReadCloser, v any) error {
	body, err := ioutil.ReadAll(data)
	if err != nil {
		return err
	}
	defer data.Close()
	if err = json.Unmarshal(body, &v); err != nil {
		return err
	}
	return nil
}

func FormateDate(format string, dateTime interface{})  string {
	dtStr := dateTime.(string)
	dtStr = strings.Replace(dtStr, " ", "T", 1)
	dt, err := time.Parse(time.RFC3339, dtStr)
	if err != nil {
		return dtStr
	}
	l := dt.Format(format)
	return l
}

func TimeDuration(t string) string {
	dtStr := strings.Replace(t, " ", "T", 1)
	dt, err := time.Parse(time.RFC3339, dtStr)
	if err != nil {
		return dtStr
	}
	duration := time.Since(dt)
	switch {
	case duration < time.Minute:
		return fmt.Sprintf("%d seconds ago", int(duration.Seconds()))
	case duration < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	case duration < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	default:
		return fmt.Sprintf("%d days ago", int(duration.Hours()/24))
	}
}

func CheckRoute(path string, pattern string) bool {
	if path == "" || pattern == "" {
		return false
	}
	re := regexp.MustCompile(pattern)
	return re.MatchString(path)
}

func StrToInt(strVal string) int {
	intVal := 0
	if strVal != ""{
		i, err := strconv.Atoi(strVal)
		if err != nil {
			return intVal
		}
		intVal = i
	}
	return intVal
}