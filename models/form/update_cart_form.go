package models

import (
	"github.com/mholt/binding"
	. "github.com/o0khoiclub0o/piflab-store-api-go/lib"
	. "github.com/o0khoiclub0o/piflab-store-api-go/models"
	. "github.com/o0khoiclub0o/piflab-store-api-go/models/repository"

	"errors"
	"fmt"
	"net/http"
)

type CartForm struct {
	Product_Id  *uint   `json:"product_id"`
	Quantity    *int    `json:"quantity"`
	AccessToken *string `json:"access_token"`
	Fields      string
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
		&form.Fields: binding.Field{
			Form: "fields",
		},
	}
}

func (form *CartForm) Validate(method string, app ...*App) error {
	var cart = new(Cart)
	var err error

	if method == "GET" {
		if form.AccessToken == nil {
			return errors.New("Access Token is required")
		}

		// If use GET method, the user must provide app interface
		// Get cart info based on AccessToken
		if cart, err = (CartRepository{app[0].DB}).GetCart(*form.AccessToken); err != nil {
			if err.Error() == "record not found" {
				return errors.New("Access Token is invalid")
			}

			// unknown err, return anyway
			return err
		}

		// TODO: need to check cart.IsCheckout instead
		if cart.IsCheckout == true {
			return errors.New("Cart is in already checkout, please create another cart")
		}
	}

	if method == "PUT_CART" {
		// PUT_CART can be nil access_token, this may be a create cart request
		if form.AccessToken != nil {
			if _, err = (CartRepository{app[0].DB}).GetCart(*form.AccessToken); err != nil {
				if err.Error() == "record not found" {
					return errors.New("Access Token is invalid")
				}

				// unknown err, return anyway
				return err
			}
		}

		if form.Product_Id == nil {
			return errors.New("No Product selected")
		}
		if _, err := (ProductRepository{}).FindById(*form.Product_Id); err != nil {
			return fmt.Errorf("Product Id %v not found", *form.Product_Id)
		}

		if form.Quantity == nil {
			return errors.New("No Quantity specified")
		}
		if *form.Quantity == 0 {
			return errors.New("Quantity should not be 0")
		}
	}

	if method == "DELETE" {
		if form.AccessToken == nil {
			return errors.New("Access Token is required")
		}
	}

	if method == "PUT_ITEM" {
		if form.AccessToken == nil {
			return errors.New("Access Token is required")
		}
		if _, err = (CartRepository{app[0].DB}).GetCart(*form.AccessToken); err != nil {
			if err.Error() == "record not found" {
				return errors.New("Access Token is invalid")
			}

			// unknown err, return anyway
			return err
		}

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
	}

	return nil
}

func (form *CartForm) GetProductInfo(product_id uint) (product_name string, product_price int, err error) {
	product, _ := (ProductRepository{}).FindById(product_id)
	if product == nil {
		return "", 0, fmt.Errorf("Product Id %v not found", product_id)
	}
	return product.Name, product.Price, nil
}

func (form *CartForm) Cart(app *App, item_id ...uint) (*Cart, error) {
	var cart = new(Cart)
	var err error
	var product_name string
	var product_price int

	if form.AccessToken != nil {
		// Get cart info based on AccessToken
		if cart, err = (CartRepository{app.DB}).GetCart(*form.AccessToken); err != nil {
			if err.Error() == "record not found" {
				return cart, errors.New("Access Token is invalid")
			}

			// unknown err, return anyway
			return cart, err
		}
	}

	// DELETE method should not update
	if form.Product_Id != nil && form.Quantity != nil {
		product_name, product_price, err = form.GetProductInfo(*form.Product_Id)
		if err != nil {
			return nil, err
		}
		err = cart.UpdateItems(form.Product_Id, nil, *form.Quantity, product_name, product_price)
	}

	// PUT CartItem, should retrieve ProductId based on ItemId
	if form.Product_Id == nil && form.Quantity != nil {
		product_id := cart.GetProductId(item_id[0])
		if product_id == 0 {
			return nil, fmt.Errorf("Item id %v not found", item_id[0])
		}
		product_name, product_price, err = form.GetProductInfo(product_id)
		if err != nil {
			return nil, err
		}
		err = cart.UpdateItems(nil, &item_id[0], *form.Quantity, product_name, product_price)
	}

	if cart.IsCheckout == true {
		return cart, errors.New("Cart is already checkout, please create another cart")
	}

	return cart, err
}
