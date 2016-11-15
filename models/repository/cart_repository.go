package repository

import (
	"github.com/icrowley/fake"
	. "github.com/o0khoiclub0o/piflab-store-api-go/lib"
	. "github.com/o0khoiclub0o/piflab-store-api-go/models"

	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type CartRepository struct {
	*DB
}

func (repo CartRepository) getOrderItemsInfo(order_items []CartItem, get_product_price_name bool) error {
	for idx, item := range order_items {
		product := &Product{}
		var err error

		product.Id = item.ProductId
		product, err = (ProductRepository{}).FindById(product.Id)
		if err != nil {
			return fmt.Errorf("Product %v", err)
		}

		// This option is for cart/checkout
		// + when cart, we will update the product price and name whenever there is a change
		// + when checkout, we will not fetch the product price and name, it is stored in the order's db table
		if get_product_price_name == true {
			order_items[idx].ProductPrice = product.Price
			order_items[idx].ProductName = product.Name
		}

		order_items[idx].ProductImageThumbnailUrl = product.ImageThumbnailUrl
	}

	return nil
}

func (repo CartRepository) generateAccessToken(cart *Cart) error {
	rand.Seed(time.Now().UTC().UnixNano())

try_gen_other_value:
	cart.AccessToken = fake.CharactersN(32)

	temp_cart := &Cart{}
	if err := repo.DB.Where("access_token = ?", cart.AccessToken).Find(temp_cart).Error; err != nil {
		// Check if err is not found -> access_token is unique
		if err.Error() == "record not found" {
			return nil
		}

		// Otherwise, this is database operation error
		return errors.New("Database error")
	}

	// duplicate, try again
	goto try_gen_other_value
}

func (repo CartRepository) clearNullQuantity() {
	repo.DB.Delete(CartItem{}, "quantity=0")
}

func (repo CartRepository) createCart(cart *Cart) error {
	if err := repo.generateAccessToken(cart); err != nil {
		return err
	}

	if err := repo.getOrderItemsInfo(cart.Items, true); err != nil {
		return err
	}

	if err := repo.DB.Create(cart).Error; err != nil {
		return err
	}

	return nil
}

func (repo CartRepository) updateCart(cart *Cart) error {
	tx := repo.DB.Begin()

	if err := repo.getOrderItemsInfo(cart.Items, true); err != nil {
		tx.Rollback()
		return err
	}

	// Update the cart
	if err := tx.Save(cart).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	repo.clearNullQuantity()

	// TODO: bring this method out of repo, should call in handler
	// Don't return access_token when updating
	cart.EraseAccessToken()

	return nil
}

func (repo CartRepository) GetCart(access_token string) (*Cart, error) {
	cart := &Cart{}
	items := &[]CartItem{}

	// find a cart by its access_token
	if err := repo.DB.Where("access_token = ?", access_token).Find(cart).Error; err != nil {
		return nil, err
	}

	// use cart.Id to find its CartItem data (cart.Id is its forein key)
	if err := repo.DB.Where("cart_id = ?", cart.Id).Find(items).Error; err != nil {
		return nil, err
	}

	// use the cart.Items to update products information
	cart.Items = *items

	repo.getOrderItemsInfo(cart.Items, true)

	return cart, nil
}

func (repo CartRepository) SaveCart(cart *Cart) error {
	if cart.AccessToken == "" {
		return repo.createCart(cart)
	}
	return repo.updateCart(cart)
}

func (repo CartRepository) DeleteCartItem(cart *Cart, item_id uint) error {
	item := CartItem{}

	repo.getOrderItemsInfo(cart.Items, true)

	// use cart.Id to find its CartItem data (cart.Id is its forein key)
	if err := repo.DB.Where("id = ? AND cart_id = ?", item_id, cart.Id).Find(&item).Error; err != nil {
		if err.Error() == "record not found" {
			return errors.New("Not found Item Id in a Cart")
		}

		return err
	}

	repo.DB.Delete(&item)

	// use cart.Id to find its CartItem data (cart.Id is its forein key)
	items := &[]CartItem{}
	repo.DB.Where("cart_id = ?", cart.Id).Find(items)
	cart.Items = *items

	return nil
}

func (repo CartRepository) CheckoutCart(cart *Cart) (checkout *CheckoutReturn, err error) {
	type CheckoutCartForm struct {
		AccessToken     string     `json:"access_token"`
		Items           []CartItem `json:"items,omitempty"`
		CustomerName    string     `json:"name"`
		CustomerAddress string     `json:"address"`
		CustomerPhone   string     `json:"phone"`
		CustomerEmail   string     `json:"email"`
		CustomerNote    string     `json:"note"`
	}

	form := CheckoutCartForm{
		AccessToken:     cart.AccessToken,
		Items:           cart.Items,
		CustomerName:    cart.CustomerName,
		CustomerAddress: cart.CustomerAddress,
		CustomerPhone:   cart.CustomerPhone,
		CustomerEmail:   cart.CustomerEmail,
		CustomerNote:    cart.CustomerNote,
	}
	var ret = new(CheckoutReturn)

	// Call ORERS_SERVICE_API
	response, body := HttpRequest("POST",
		GetOrderService()+"/cart/checkout",
		form)
	if response.Status != "200 OK" {
		return nil, ParseError(body)
	}

	if err := json.Unmarshal([]byte(body), ret); err != nil {
		return nil, err
	}

	// If success, save the new status IsCheckout to true and save to db
	cart.IsCheckout = true
	repo.DB.Save(cart)

	return ret, nil
}
