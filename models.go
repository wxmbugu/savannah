package savannah

import (
	"time"
)

type (
	User struct {
		ID    int    `json:"id"`
		Code  string `json:"code"`
		Email string `json:"email" validate:"required"`
	}
	Item struct {
		ID          int     `json:"id"`
		Price       float32 `json:"price" validate:"required"`
		Name        string  `json:"name" validate:"required"`
		Description string  `json:"description" validate:"required"`
	}
	Orders struct {
		ID     int       `json:"id"`
		UserId int       `json:"user_id"`
		ItemID int       `json:"item_id"  validate:"required"`
		Qty    int       `json:"qty" validate:"required"`
		Time   time.Time `json:"time" `
	}
	database interface {
		CreateUser(user User) (*User, error)
		CreateItem(item Item) (*Item, error)
		CreateOrders(order Orders) (*Orders, error)

		FindUser(id int) (*User, error)
		FindUserbyEmail(string) (*User, error)
		FindItem(id int) (*Item, error)
		FindOrders(id int) (*Orders, error)

		DeleteUser(id int) error
		DeleteItem(id int) error
		DeleteOrders(id int) error

		UpdateUser(user User) error
		UpdateItem(item Item) error
		UpdateOrders(order Orders) error
	}
)
