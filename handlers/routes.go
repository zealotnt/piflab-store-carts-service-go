package handlers

import (
	. "github.com/o0khoiclub0o/piflab-store-api-go/lib"
)

func GetRoutes() Routes {
	return Routes{
		Route{"GET", "/", IndexHandler},

		Route{"GET", "/cart", GetCartHandler},
		Route{"PUT", "/cart/items", UpdateCartHandler},
		Route{"PUT", "/cart/items/{id}", UpdateCartItemHandler},
		Route{"DELETE", "/cart/items/{id}", DeleteCartItemHandler},
		Route{"POST", "/cart/checkout", CheckoutCartHandler},
	}
}
