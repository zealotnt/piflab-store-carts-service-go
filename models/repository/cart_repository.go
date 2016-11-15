package repository

import (
	"github.com/icrowley/fake"
	. "github.com/o0khoiclub0o/piflab-store-api-go/lib"
	. "github.com/o0khoiclub0o/piflab-store-api-go/models"

	"errors"
	"math/rand"
	"time"
)

type CartRepository struct {
	*DB
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

	// cart.Items[0].Id = 1
	// cart.Items[0].OrderId = 1
	// cart.Items[0].ProductName = "abc"
	// cart.Items[0].ProductPrice = 123
	PR_DUMP(cart)

	if err := repo.DB.Create(cart).Error; err != nil {
		return err
	}

	return nil
}

func (repo CartRepository) updateCart(cart *Cart) error {
	tx := repo.DB.Begin()

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

func (repo CartRepository) FindByOrderId(cart_code string) (*Cart, error) {
	cart := &Cart{}
	items := &[]CartItem{}

	// find a cart by its access_token
	if err := repo.DB.Where("code = ?", cart_code).Find(cart).Error; err != nil {
		return nil, err
	}

	// use cart.Id to find its CartItem data (cart.Id is its forein key)
	if err := repo.DB.Where("cart_id = ?", cart.Id).Find(items).Error; err != nil {
		return nil, err
	}

	// use the cart.Items to update products information
	cart.Items = *items

	return cart, nil
}

func (repo CartRepository) GetOrderByOrdercode(cart_code string) (*Cart, error) {
	// TODO: Call Order_Service_API
	// cart := &Cart{}
	// items := &[]CartItem{}

	// // find a cart by its cart_code
	// if err := repo.DB.Where("code = ?", cart_code).Find(cart).Error; err != nil {
	// 	return nil, err
	// }

	// // use cart.Id to find its CartItem data (cart.Id is its forein key)
	// if err := repo.DB.Where("cart_id = ?", cart.Id).Find(items).Error; err != nil {
	// 	return nil, err
	// }

	// // use the cart.Items to update products information
	// cart.Items = *items

	// return cart, nil
	return nil, nil
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

func (repo CartRepository) CheckoutCart(cart *Cart) error {
	// TODO: Call ORERS_SERVICE_API

	// TODO: if success, save the new status IsCheckout to true and save to db

	return nil
}
