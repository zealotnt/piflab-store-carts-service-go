package repository

import (
	. "github.com/o0khoiclub0o/piflab-store-api-go/lib"
	. "github.com/o0khoiclub0o/piflab-store-api-go/models"

	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type ProductRepository struct {
}

func (repo ProductRepository) FindByListId(ids []uint64) (*ProductListId, error) {
	product_list := &ProductListId{}
	if len(ids) == 0 {
		return nil, errors.New("ids array null")
	}
	if len(ids) == 1 {
		product, err := repo.FindById(uint(ids[0]))
		if err != nil {
			if err.Error() == "record not found" {
				product_list.ErrorList = append(product_list.ErrorList, ids[0])
				return product_list, nil
			}
			return nil, err
		}
		product_list.ProductSlice = append(product_list.ProductSlice, *product)
	}
	if len(ids) > 1 {
		product_ids := strings.Trim(strings.Replace(fmt.Sprint(ids), " ", ",", -1), "[]")
		response, body := HttpRequest("GET", GetProductService()+"/products/"+product_ids, nil)
		if response.Status != "200 OK" {
			return nil, ParseError(body)
		}

		if err := json.Unmarshal([]byte(body), &product_list); err != nil {
			return nil, err
		}

		// Try to parse missing product ids by comma seperated list
		if product_list.Error != nil {
			str_list := strings.Split(*product_list.Error, ",")
			for _, str := range str_list {
				id, _ := strconv.ParseUint(str, 10, 32)
				product_list.ErrorList = append(product_list.ErrorList, id)
			}
		}
	}

	return product_list, nil
}

func (repo ProductRepository) FindById(id uint) (*Product, error) {
	product := &Product{}
	response, body := HttpRequest("GET", GetProductService()+"/products/"+strconv.Itoa(int(id)), nil)
	if response.Status != "200 OK" {
		return nil, ParseError(body)
	}

	if err := json.Unmarshal([]byte(body), &product); err != nil {
		return nil, err
	}

	return product, nil
}

func (repo ProductRepository) GetPage(offset uint, limit uint, search string) (*ProductPage, error) {
	product_by_page := &ProductPage{}

	response, body := HttpRequest("GET",
		GetProductService()+"/products?offset="+
			strconv.Itoa(int(offset))+
			"&limit="+strconv.Itoa(int(limit))+
			"&q="+search,
		nil)
	if response.Status != "200 OK" {
		return nil, ParseError(body)
	}

	if err := json.Unmarshal([]byte(body), &product_by_page); err != nil {
		return nil, err
	}

	return product_by_page, nil
}

func (repo ProductRepository) createProduct(product *Product) error {
	response, body := HttpRequest("POST",
		GetProductService()+"/products",
		product)
	if response.Status != "201 Created" {
		return ParseError(body)
	}
	if err := json.Unmarshal([]byte(body), product); err != nil {
		return err
	}

	return nil
}

func (repo ProductRepository) updateProduct(product *Product) error {
	response, body := HttpRequest("PUT",
		GetProductService()+"/products/"+strconv.Itoa(int(product.Id)),
		product)
	if response.Status != "200 OK" {
		return ParseError(body)
	}
	if err := json.Unmarshal([]byte(body), product); err != nil {
		return err
	}

	return nil
}

func (repo ProductRepository) SaveProduct(product *Product) error {
	if product.Id == 0 {
		return repo.createProduct(product)
	}
	return repo.updateProduct(product)

	return nil
}

func (repo ProductRepository) DeleteProduct(id uint) (*Product, error) {
	product, err := repo.FindById(id)
	if err != nil {
		return product, err
	}
	response, body := HttpRequest("DELETE", GetProductService()+"/products/"+strconv.Itoa(int(id)), "")
	if response.Status != "200 OK" {
		return nil, ParseError(body)
	}

	return product, nil
}
