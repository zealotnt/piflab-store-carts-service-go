package models

import (
	"github.com/mholt/binding"
	. "github.com/o0khoiclub0o/piflab-store-api-go/lib"
	. "github.com/o0khoiclub0o/piflab-store-api-go/models"
	. "github.com/o0khoiclub0o/piflab-store-api-go/models/repository"

	"errors"
	"net/http"
)

type CartForm struct {
	Product_Id  *uint   `json:"product_id"`
	Quantity    *int    `json:"quantity"`
	AccessToken *string `json:"access_token"`

	Name  *string `json:"name"`
	Price *uint   `json:"price"`

	Fields string
}

func (form *CartForm) FieldMap(req *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&form.Product_Id: binding.Field{
			Form: "product_id",
		},
		&form.Quantity: binding.Field{
			Form: "quantity",
		},
		&form.AccessToken: binding.Field{
			Form: "access_token",
		},
		&form.Name: binding.Field{
			Form: "name",
		},
		&form.Price: binding.Field{
			Form: "price",
		},
		&form.Fields: binding.Field{
			Form: "fields",
		},
	}
}

func (form *CartForm) Validate(method string, app ...*App) error {
	if method == "GET" {
		if form.AccessToken == nil {
			return errors.New("Access Token is required")
		}

		// If use GET method, the user must provide app interface
		var order = new(Order)
		var err error
		// Get order info based on AccessToken
		if order, err = (OrderRepository{app[0].DB}).GetOrder(*form.AccessToken); err != nil {
			if err.Error() == "record not found" {
				return errors.New("Access Token is invalid")
			}

			// unknown err, return anyway
			return err
		}

		// TODO: need to check order.IsCheckout instead
		if order.IsCheckout == true {
			return errors.New("Order is in already checkout, please create another order")
		}
	}

	if method == "PUT_CART" {
		if form.Product_Id == nil {
			return errors.New("No Product selected")
		}

		if form.Quantity == nil {
			return errors.New("No Quantity specified")
		}
		if *form.Quantity == 0 {
			return errors.New("Quantity should not be 0")
		}

		if form.Name == nil {
			return errors.New("Product name required when save cart")
		}
		if form.Price == nil {
			return errors.New("Price name required when save cart")
		}
	}

	if method == "DELETE" {
		if form.AccessToken == nil {
			return errors.New("Access Token is required")
		}
	}

	if method == "PUT_ITEM" {
		// don't use product_id when update Cart Item
		if form.Product_Id != nil {
			form.Product_Id = nil
		}

		if form.Quantity == nil {
			return errors.New("No Quantity specified")
		}
		if *form.Quantity < 0 {
			return errors.New("Quantity should bigger or equal to 0")
		}

		if form.Name == nil {
			return errors.New("Product name required when save cart")
		}
		if form.Price == nil {
			return errors.New("Price name required when save cart")
		}
	}

	return nil
}

func (form *CartForm) Order(app *App, item_id ...uint) (*Order, error) {
	var order = new(Order)
	var err error

	if form.AccessToken != nil {
		// Get order info based on AccessToken
		if order, err = (OrderRepository{app.DB}).GetOrder(*form.AccessToken); err != nil {
			if err.Error() == "record not found" {
				return order, errors.New("Access Token is invalid")
			}

			// unknown err, return anyway
			return order, err
		}
	}

	// DELETE method should not update
	if form.Product_Id != nil && form.Quantity != nil {
		err = order.UpdateItems(form.Product_Id, nil, *form.Quantity, *form.Name, int(*form.Price))
	}

	// PUT CartItem, should retrieve ProductId based on ItemId
	if form.Product_Id == nil && form.Quantity != nil {
		err = order.UpdateItems(nil, &item_id[0], *form.Quantity, *form.Name, int(*form.Price))
	}

	// If this is the first time create order,
	// this will avoid error when create order
	// (pq: invalid input value for enum order_status: "")
	// Note: Need to implement
	// if order.Status == "" {
	// 	order.Status = "cart"
	// }

	// if order.Status != "cart" {
	// 	return order, errors.New("Order is in " + order.Status + " state, please use another cart")
	// }

	return order, err
}
