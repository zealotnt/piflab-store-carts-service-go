package models

import (
	"github.com/mholt/binding"
	. "github.com/o0khoiclub0o/piflab-store-api-go/lib"
	. "github.com/o0khoiclub0o/piflab-store-api-go/models"
	. "github.com/o0khoiclub0o/piflab-store-api-go/models/repository"

	"errors"
	"net/http"
)

type CheckoutCartForm struct {
	AccessToken     *string `json:"access_token"`
	CustomerName    *string `json:"name"`
	CustomerAddress *string `json:"address"`
	CustomerPhone   *string `json:"phone"`
	CustomerEmail   *string `json:"email"`
	CustomerNote    *string `json:"note"`
	Fields          string
}

func (form *CheckoutCartForm) FieldMap(req *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&form.AccessToken: binding.Field{
			Form: "access_token",
		},
		&form.CustomerName: binding.Field{
			Form: "name",
		},
		&form.CustomerAddress: binding.Field{
			Form: "address",
		},
		&form.CustomerPhone: binding.Field{
			Form: "phone",
		},
		&form.CustomerEmail: binding.Field{
			Form: "email",
		},
		&form.CustomerNote: binding.Field{
			Form: "note",
		},
		&form.Fields: binding.Field{
			Form: "fields",
		},
	}
}

func (form *CheckoutCartForm) Validate() error {
	if form.AccessToken == nil {
		return errors.New("Access Token is required")
	}

	if form.CustomerName == nil {
		return errors.New("Customer's Name is required")
	}

	if form.CustomerAddress == nil {
		return errors.New("Customer's Address is required")
	}

	if form.CustomerPhone == nil {
		return errors.New("Customer's Phone number is required")
	}

	if form.CustomerEmail == nil {
		return errors.New("Customer's Email is required")
	}
	if !ValidateEmail(*form.CustomerEmail) {
		return errors.New("Customer's Email address is invalid")
	}

	return nil
}

func (form *CheckoutCartForm) Cart(app *App) (*Cart, error) {
	var order = new(Cart)
	var err error

	if order, err = (CartRepository{app.DB}).GetCart(*form.AccessToken); err != nil {
		if err.Error() == "record not found" {
			return order, errors.New("Access Token is invalid")
		}

		// unknown err, return anyway
		return order, err
	}

	// TODO: need to check order.IsCheckout instead
	if order.IsCheckout == true {
		return order, errors.New("Cart is already checked out, please create another cart")
	}

	order.OrderInfo.CustomerName = *form.CustomerName
	order.OrderInfo.CustomerAddress = *form.CustomerAddress
	order.OrderInfo.CustomerPhone = *form.CustomerPhone
	order.OrderInfo.CustomerEmail = *form.CustomerEmail

	if form.CustomerNote != nil {
		order.OrderInfo.CustomerNote = *form.CustomerNote
	}

	return order, err
}
