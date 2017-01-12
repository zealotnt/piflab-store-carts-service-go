package lib

import (
	"bytes"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"net/http"
	"regexp"

	"github.com/ttacon/libphonenumber"
)

var validation_result = map[libphonenumber.ValidationResult]string{
	libphonenumber.IS_POSSIBLE:          "IS_POSSIBLE",
	libphonenumber.INVALID_COUNTRY_CODE: "INVALID_COUNTRY_CODE",
	libphonenumber.TOO_SHORT:            "TOO_SHORT",
	libphonenumber.TOO_LONG:             "TOO_LONG",
}

func ValidatePhone(phone string) error {
	num, _ := libphonenumber.Parse(phone, "VN")

	if libphonenumber.IsPossibleNumberWithReason(num) != libphonenumber.IS_POSSIBLE {
		return errors.New(validation_result[libphonenumber.IsPossibleNumberWithReason(num)])
	}

	return nil
}

func ValidateEmail(email string) bool {
	Re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return Re.MatchString(email)
}

func JSON(w http.ResponseWriter, params ...interface{}) {
	setHTTPStatus(w, params)

	obj := params[0]

	switch obj.(type) {
	case error:
		obj = &Error{obj.(error)}
	}

	json.NewEncoder(w).Encode(obj)
}

func Image(w http.ResponseWriter, img image.Image) {
	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, img); err != nil {
		JSON(w, errors.New("unable to encode image."))
		return
	}

	// w.Header().Set("Content-Type", "image/png")
	if _, err := w.Write(buffer.Bytes()); err != nil {
		JSON(w, errors.New("unable to write image."))
	}
}

func setHTTPStatus(w http.ResponseWriter, params []interface{}) {
	w.Header().Add("Access-Control-Allow-Origin", `*`)
	w.Header().Add("Access-Control-Allow-Methods", `GET, POST, PUT, DELETE, OPTIONS`)
	w.Header().Set("Content-Type", `application/json`)

	if len(params) == 2 {
		status := params[1].(int)

		if status == 401 {
			w.Header().Add("WWW-Authenticate", `xBasic realm="fake"`)
		}

		w.WriteHeader(status)
	}
}
