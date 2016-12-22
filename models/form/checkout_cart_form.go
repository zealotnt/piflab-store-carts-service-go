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
	var cart = new(Cart)
	var err error

	if cart, err = (CartRepository{app.DB}).GetCart(*form.AccessToken); err != nil {
		if err.Error() == "record not found" {
			return cart, errors.New("Access Token is invalid")
		}

		// unknown err, return anyway
		return cart, err
	}

	// TODO: need to check cart.IsCheckout instead
	if cart.IsCheckout == true {
		return cart, errors.New("Cart is already checked out, please create another cart")
	}

	cart.OrderInfo.CustomerName = *form.CustomerName
	cart.OrderInfo.CustomerAddress = *form.CustomerAddress
	cart.OrderInfo.CustomerPhone = *form.CustomerPhone
	cart.OrderInfo.CustomerEmail = *form.CustomerEmail

	if form.CustomerNote != nil {
		cart.OrderInfo.CustomerNote = *form.CustomerNote
	}

	return cart, err
}
