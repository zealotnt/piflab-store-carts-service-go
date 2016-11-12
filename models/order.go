package models

import (
	"errors"
	"fmt"
	"time"
)

type Amount struct {
	Subtotal uint `json:"subtotal"`
	Shipping uint `json:"shipping"`
	Total    uint `json:"total"`
}

type OrderInfo struct {
	OrderCode       string `json:"-" sql:"column:code"`
	CustomerName    string `json:"name" sql:"column:customer_name"`
	CustomerAddress string `json:"address" sql:"column:customer_address"`
	CustomerPhone   string `json:"phone" sql:"column:customer_phone"`
	CustomerEmail   string `json:"email" sql:"column:customer_email"`
	CustomerNote    string `json:"note" sql:"column:note"`
}

type CheckoutReturn struct {
	Id        string      `json:"id,omitempty"`
	Items     []OrderItem `json:"items,omitempty"`
	Amounts   Amount      `json:"amounts" sql:"-"`
	OrderInfo *OrderInfo  `json:"customer,omitempty" sql:"-"`
	Status    string      `json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Order struct {
	Id          uint   `json:"-"`
	AccessToken string `json:"access_token,omitempty"`
	IsCheckout  bool   `json:"is_checkout"`

	Items []OrderItem `json:"items" sql:"cart_items"`

	OrderInfo `json:"-" sql:"-"`

	Amounts Amount `json:"amounts" sql:"-"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrderStatusLog struct {
	Id        uint
	Code      string
	Status    string
	CreatedAt time.Time
}

type OrderItem struct {
	Id                       uint    `json:"id" sql:"id"`
	CartId                   uint    `json:"-" sql:"REFERENCES carts(id)"`
	ProductId                uint    `json:"product_id" sql:"column:product_id"`
	ProductName              string  `json:"name" sql:"column:name"`
	ProductImageThumbnailUrl *string `json:"image_thumbnail_url" sql:"-"`
	ProductPrice             int     `json:"price" sql:"column:price"`
	Quantity                 int     `json:"quantity"`
}

func (Order) TableName() string {
	return "carts"
}

func (OrderItem) TableName() string {
	return "cart_items"
}

func (order *Order) UpdateItems(product_id *uint, item_id *uint, quantity int, product_name string, product_price int) error {
	for idx, item := range order.Items {
		if product_id != nil {
			if item.ProductId == *product_id {
				// The item already in the list, add or subtract quantity
				order.Items[idx].Quantity += quantity
				if order.Items[idx].Quantity < 0 {
					order.Items[idx].Quantity = 0
				}
				order.Items[idx].ProductName = product_name
				order.Items[idx].ProductPrice = product_price
				return nil
			}
		}

		// update quantity base on {product_id, quantity}
		if item_id != nil {
			if item.Id == *item_id {
				order.Items[idx].Quantity = quantity
				order.Items[idx].ProductName = product_name
				order.Items[idx].ProductPrice = product_price
				return nil
			}
		}
	}

	if item_id != nil {
		return fmt.Errorf("Item ID %v not found", item_id)
	}

	if quantity < 0 {
		return errors.New("Quantity for item should bigger than 0")
	}

	// The item is new, add it to the []item list {product_id, quantity}
	if product_id != nil {
		order.Items = append(order.Items,
			OrderItem{
				ProductId:    *product_id,
				Quantity:     quantity,
				ProductName:  product_name,
				ProductPrice: product_price,
			})
	}

	return nil
}

func (order *Order) CalculateAmount() {
	for _, item := range order.Items {
		order.Amounts.Subtotal += uint(item.ProductPrice) * uint(item.Quantity)
	}
	order.Amounts.Shipping = 0
	order.Amounts.Total = order.Amounts.Shipping + order.Amounts.Subtotal
}

func (order *Order) EraseAccessToken() {
	order.AccessToken = ""
}

func (order *Order) RemoveZeroQuantityItems() {
	for idx, _ := range order.Items {
		if order.Items[idx].Quantity <= 0 {
			order.Items = append(order.Items[:idx], order.Items[idx+1:]...)
			return
		}
	}
}
