package utils

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestRead(t *testing.T) {
	s := Settings{
		AccessKeyId: "1234567890",
		SecretKey:   "09876654321",
		SocksV5Port: 1774,
	}
	setAllENVs(&s)
	env := ENVData{}
	settings, err := env.Read()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if (*settings).AccessKeyId != s.AccessKeyId || (*settings).SecretKey != s.SecretKey || (*settings).SocksV5Port != s.SocksV5Port {
		t.Error("Unexpected values from the env")
	}
}

func setAllENVs(settings *Settings) {
	values := reflect.ValueOf(*settings)
	types := reflect.TypeOf(*settings)
	for i := 0; i < values.NumField(); i++ {
		key := fmt.Sprintf("%v", types.Field(i).Name)
		value := fmt.Sprintf("%v", values.Field(i).Interface())
		_ = os.Setenv(key, value)
	}

}
