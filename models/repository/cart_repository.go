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

func UpdateCartItemsNamePrice(cart *Cart) {
	for idx, order := range cart.Items {
		for _, product := range cart.ProductList.ProductSlice {
			if order.ProductId == product.Id {
				cart.Items[idx].ProductName = product.Name
				cart.Items[idx].ProductPrice = product.Price
			}
		}
	}
}

func SetCartItemsThumnailUrl(cart_items []CartItem, product_list ProductListId) {
	for idx, order := range cart_items {
		for _, product := range product_list.ProductSlice {
			if order.ProductId == product.Id {
				cart_items[idx].ProductImageThumbnailUrl = product.ImageThumbnailUrl
			}
		}
	}
}

func GetCartAlerts(cart *Cart, product_list ProductListId) {
	// clear old array
	cart.Alerts = nil

	// check for product deleted
	for _, id := range product_list.ErrorList {
		cart.IsError = true

		cart.Alerts = append(cart.Alerts,
			Alert{Type: "error",
				Message: fmt.Sprintf("Product id %d '%s' is deleted, should remove item_id %d to be able to checkout",
					id,
					cart.GetProductName(uint(id)),
					cart.GetItemId(uint(id))),
			})
	}

	// check for product name, price changes
	for idx, order := range cart.Items {
		for _, product := range product_list.ProductSlice {
			if order.ProductId == product.Id {
				if cart.Items[idx].ProductPrice != product.Price {
					cart.IsWarning = true

					cart.Alerts = append(cart.Alerts,
						Alert{Type: "warning",
							Message: fmt.Sprintf("Product price of p_id %d changed, from %d to %d",
								order.ProductId,
								cart.Items[idx].ProductPrice,
								product.Price),
						})
				}
				if cart.Items[idx].ProductName != product.Name {
					cart.IsWarning = true

					cart.Alerts = append(cart.Alerts,
						Alert{Type: "warning",
							Message: fmt.Sprintf("Product name of p_id %d changed, from '%s' to '%s'",
								order.ProductId,
								cart.Items[idx].ProductName,
								product.Name),
						})
				}
			}
		}
	}
}

func GetProductList(cart *Cart) error {
	var product_id_list []uint64

	// Get the product list
	for _, item := range cart.Items {
		product_id_list = append(product_id_list, uint64(item.ProductId))
	}

	// request to product service
	product_list, err := (ProductRepository{}).FindByListId(product_id_list)
	if err != nil {
		return err
	}

	// update to cart internal structure, for further investigate
	cart.ProductList = *product_list

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

	err := GetProductList(cart)
	if err != nil {
		return nil, err
	}

	GetCartAlerts(cart, cart.ProductList)

	SetCartItemsThumnailUrl(cart.Items, cart.ProductList)

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

	// because there is a last delete cart item action, update the alert message
	err := GetProductList(cart)
	if err != nil {
		return err
	}
	GetCartAlerts(cart, cart.ProductList)

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
