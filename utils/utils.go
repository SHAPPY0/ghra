package utils

import (
	"io"
	"time"
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
	dt, err := time.Parse(time.RFC3339, dtStr)
	if err != nil {
		return dtStr
	}
	return dt.Format(format)
}