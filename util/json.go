package util

import . "github.com/alsey/uid-gen/logger"
import "encoding/json"

func Stringify(obj interface{}) (str string, err error) {

	var obj_b []byte

	if obj_b, err = json.Marshal(obj); nil != err {
		Error("failed to marshal %v, %v", obj, err)
		return "", err
	}

	str = byte2str(obj_b)
	return str, nil
}

func Parse(str string, obj interface{}) {
	json.Unmarshal([]byte(str), &obj)
}

func byte2str(c []byte) string {
	n := -1
	for i, b := range c {
		if b == 0 {
			break
		}
		n = i
	}
	return string(c[:n+1])
}
