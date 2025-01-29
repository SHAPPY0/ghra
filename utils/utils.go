package utils

import (
	"io"
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

// func StrToInt(val string) int {
// 	v, err := strconv.Atoi(val)
// 	if err != nil {
// 		return 0
// 	}
// 	return v
// }