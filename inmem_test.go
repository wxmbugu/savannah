package savannah

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var db database

func TestMain(m *testing.M) {

	db = NewMockStore()
	os.Exit(m.Run())
}

func TestMockInMemDB_CreateUser(t *testing.T) {

	user := User{
		Email: "john@example.com",
		Code:  String(10),
	}

	createdUser, err := db.CreateUser(user)
	assert.NoError(t, err)
	assert.NotNil(t, createdUser)
	assert.NotZero(t, createdUser.ID)
}

func TestMockInMemDB_FindUser(t *testing.T) {

	user := User{
		Email: "john@example.com",
		Code:  String(10),
	}

	createdUser, err := db.CreateUser(user)
	assert.NoError(t, err)
	assert.NotNil(t, createdUser)

	foundUser, err := db.FindUser(createdUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, createdUser.ID, foundUser.ID)
}

func TestMockInMemDB_UpdateUser(t *testing.T) {

	user := User{
		Email: "john@example.com",
		Code:  String(10),
	}

	createdUser, err := db.CreateUser(user)
	assert.NoError(t, err)
	assert.NotNil(t, createdUser)

	updatedUser := User{
		Email: "updated@example.com",
		Code:  String(10),
	}
	err = db.UpdateUser(updatedUser)
	assert.NoError(t, err)

	foundUser, err := db.FindUser(updatedUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, updatedUser.Code, foundUser.Code)
	assert.Equal(t, updatedUser.Email, foundUser.Email)
}

func TestMockInMemDB_DeleteUser(t *testing.T) {

	user := User{
		Email: "john@example.com",
		Code:  String(10),
	}

	createdUser, err := db.CreateUser(user)
	assert.NoError(t, err)
	assert.NotNil(t, createdUser)

	err = db.DeleteUser(createdUser.ID)
	assert.NoError(t, err)
	_, err = db.FindUser(createdUser.ID)
	assert.Error(t, err)
}
func TestMockInMemDB_CreateItem(t *testing.T) {
	item := Item{
		Price:       29.99,
		Name:        "Sample Item",
		Description: "A sample description",
	}

	createdItem, err := db.CreateItem(item)
	assert.NoError(t, err)
	assert.NotNil(t, createdItem)
	assert.NotZero(t, createdItem.ID)
}

func TestMockInMemDB_FindItem(t *testing.T) {
	item := Item{
		Price:       29.99,
		Name:        "Sample Item",
		Description: "A sample description",
	}

	createdItem, err := db.CreateItem(item)
	assert.NoError(t, err)
	assert.NotNil(t, createdItem)

	foundItem, err := db.FindItem(createdItem.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundItem)
	assert.Equal(t, createdItem.ID, foundItem.ID)
}

func TestMockInMemDB_UpdateItem(t *testing.T) {
	item := Item{
		Price:       29.99,
		Name:        "Sample Item",
		Description: "A sample description",
	}

	createdItem, err := db.CreateItem(item)
	assert.NoError(t, err)
	assert.NotNil(t, createdItem)

	updatedItem := Item{
		ID:          createdItem.ID,
		Price:       39.99,
		Name:        "Updated Item",
		Description: "An updated description",
	}
	err = db.UpdateItem(updatedItem)
	assert.NoError(t, err)

	foundItem, err := db.FindItem(updatedItem.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundItem)
	assert.Equal(t, updatedItem.Name, foundItem.Name)
	assert.Equal(t, updatedItem.Description, foundItem.Description)
}

func TestMockInMemDB_DeleteItem(t *testing.T) {
	item := Item{
		Price:       29.99,
		Name:        "Sample Item",
		Description: "A sample description",
	}

	createdItem, err := db.CreateItem(item)
	assert.NoError(t, err)
	assert.NotNil(t, createdItem)

	err = db.DeleteItem(createdItem.ID)
	assert.NoError(t, err)
	_, err = db.FindItem(createdItem.ID)
	assert.Error(t, err)
}

func TestMockInMemDB_CreateOrders(t *testing.T) {
	order := Orders{
		UserId: 1,
		ItemID: 2,
		Qty:    3,
		Time:   time.Now(),
	}
	createdOrder, err := db.CreateOrders(order)
	assert.NoError(t, err)
	assert.NotNil(t, createdOrder)
	assert.NotZero(t, createdOrder.ID)
}

func TestMockInMemDB_FindOrders(t *testing.T) {
	order := Orders{
		UserId: 1,
		ItemID: 2,
		Qty:    3,
		Time:   time.Now(),
	}

	createdOrder, err := db.CreateOrders(order)
	assert.NoError(t, err)
	assert.NotNil(t, createdOrder)

	foundOrder, err := db.FindOrders(createdOrder.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundOrder)
	assert.Equal(t, createdOrder.ID, foundOrder.ID)
}

func TestMockInMemDB_UpdateOrders(t *testing.T) {
	order := Orders{
		UserId: 1,
		ItemID: 2,
		Qty:    3,
		Time:   time.Now(),
	}

	createdOrder, err := db.CreateOrders(order)
	assert.NoError(t, err)
	assert.NotNil(t, createdOrder)

	updatedOrder := Orders{
		ID:     createdOrder.ID,
		UserId: 1,
		ItemID: 2,
		Qty:    5,
		Time:   time.Now(),
	}
	err = db.UpdateOrders(updatedOrder)
	assert.NoError(t, err)

	foundOrder, err := db.FindOrders(updatedOrder.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundOrder)
	assert.Equal(t, updatedOrder.Qty, foundOrder.Qty)
}

func TestMockInMemDB_DeleteOrders(t *testing.T) {
	order := Orders{
		UserId: 1,
		ItemID: 2,
		Qty:    3,
		Time:   time.Now(),
	}

	createdOrder, err := db.CreateOrders(order)
	assert.NoError(t, err)
	assert.NotNil(t, createdOrder)

	err = db.DeleteOrders(createdOrder.ID)
	assert.NoError(t, err)
	_, err = db.FindOrders(createdOrder.ID)
	assert.Error(t, err)
}
