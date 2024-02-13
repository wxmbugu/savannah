package savannah

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type DB struct {
	db *sql.DB
}

func Newdb(conn *sql.DB) *DB {
	return &DB{
		db: conn,
	}
}

func (v *DB) CreateUser(user User) (*User, error) {
	sqlStatement := `
		INSERT INTO users (code, email)
		VALUES ($1, $2)
		RETURNING *;
	`
	err := v.db.QueryRow(sqlStatement, user.Code, user.Email).Scan(
		&user.ID,
		&user.Code,
		&user.Email,
	)
	return &user, err
}

func (v *DB) CreateItem(item Item) (*Item, error) {
	sqlStatement := `
		INSERT INTO items (price, name, description)
		VALUES ($1, $2, $3)
		RETURNING *;
	`
	err := v.db.QueryRow(sqlStatement, item.Price, item.Name, item.Description).Scan(
		&item.ID,
		&item.Price,
		&item.Name,
		&item.Description,
	)
	return &item, err
}

func (v *DB) CreateOrders(order Orders) (*Orders, error) {
	sqlStatement := `
		INSERT INTO orders (item_id, qty, time,user_id)
		VALUES ($1, $2, $3,$4)
		RETURNING *;
	`
	err := v.db.QueryRow(sqlStatement, order.ItemID, order.Qty, order.Time, order.UserId).Scan(
		&order.ID,
		&order.UserId,
		&order.ItemID,
		&order.Qty,
		&order.Time,
	)
	return &order, err
}

func (v *DB) FindItem(id int) (*Item, error) {
	sqlStatement := `
		SELECT * FROM items
		WHERE items.id = $1
	`
	var item Item
	err := v.db.QueryRow(sqlStatement, id).Scan(
		&item.ID,
		&item.Price,
		&item.Name,
		&item.Description,
	)
	return &item, err
}

func (v *DB) FindOrders(id int) (*Orders, error) {
	sqlStatement := `
		SELECT * FROM orders
		WHERE orders.id = $1
	`
	var order Orders
	err := v.db.QueryRow(sqlStatement, id).Scan(
		&order.ID,
		&order.UserId,
		&order.ItemID,
		&order.Qty,
		&order.Time)
	return &order, err
}

func (v *DB) FindUser(id int) (*User, error) {
	sqlStatement := `
		SELECT * FROM users
		WHERE users.id = $1
	`
	var user User
	err := v.db.QueryRow(sqlStatement, id).Scan(
		&user.ID,
		&user.Code,
		&user.Email,
	)
	return &user, err
}
func (v *DB) FindUserbyEmail(email string) (*User, error) {
	sqlStatement := `
		SELECT * FROM users
		WHERE users.email = $1
	`
	var user User
	err := v.db.QueryRow(sqlStatement, email).Scan(
		&user.ID,
		&user.Code,
		&user.Email,
	)
	return &user, err
}
func (v *DB) DeleteItem(id int) error {
	sqlStatement := `
		DELETE FROM items
		WHERE items.id = $1
	`
	_, err := v.db.Exec(sqlStatement, id)
	return err
}

func (v *DB) DeleteOrders(id int) error {
	sqlStatement := `
		DELETE FROM orders
		WHERE orders.id = $1
	`
	_, err := v.db.Exec(sqlStatement, id)
	return err
}

func (v *DB) DeleteUser(id int) error {
	sqlStatement := `
		DELETE FROM users
		WHERE users.id = $1
	`
	_, err := v.db.Exec(sqlStatement, id)
	return err
}

func (v *DB) UpdateItem(item Item) error {
	sqlStatement := `
		UPDATE items
		SET price = $2, name = $3, description = $4
		WHERE id = $1
	`
	_, err := v.db.Exec(sqlStatement, item.ID, item.Price, item.Name, item.Description)
	return err
}

func (v *DB) UpdateOrders(order Orders) error {
	sqlStatement := `
		UPDATE orders
		SET item_id = $2, qty = $3, time = $4, user_id = $5
		WHERE id = $1
	`
	_, err := v.db.Exec(sqlStatement, order.ID, order.ItemID, order.Qty, order.Time, order.UserId)
	return err
}

func (v *DB) UpdateUser(user User) error {
	sqlStatement := `
		UPDATE users
		SET  code = $2, email = $3
		WHERE id = $1
	`
	_, err := v.db.Exec(sqlStatement, user.ID, user.Code, user.Email)
	return err
}
