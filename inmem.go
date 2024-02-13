package savannah

import (
	"errors"
	"sync"
)

// MockInMemDB is a mock implementation of the database interface
type MockInMemDB struct {
	mu       sync.RWMutex
	UserData map[int]User
	ItemData map[int]Item
	Orders   map[int]Orders
}

func NewMockStore() *MockInMemDB {
	usermap := make(map[int]User)
	item_map := make(map[int]Item)
	order_map := make(map[int]Orders)
	return &MockInMemDB{
		UserData: usermap,
		ItemData: item_map,
		Orders:   order_map,
	}
}

func NewMockService() Service {
	return Service{
		service: NewMockStore(),
	}
}

func (m *MockInMemDB) CreateUser(user User) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	user.ID = generateUniqueUserID()
	m.UserData[user.ID] = user
	return &user, nil
}

func (m *MockInMemDB) CreateItem(item Item) (*Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item.ID = generateUniqueItemID()
	m.ItemData[item.ID] = item
	return &item, nil
}

func (m *MockInMemDB) CreateOrders(order Orders) (*Orders, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	order.ID = generateUniqueOrderID()
	m.Orders[order.ID] = order
	return &order, nil
}

func (m *MockInMemDB) FindUser(id int) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if user, ok := m.UserData[id]; ok {
		return &user, nil
	}
	return nil, errors.New("user not found")
}

func (m *MockInMemDB) FindUserbyEmail(email string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, user := range m.UserData {
		if user.Email == email {
			return &user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *MockInMemDB) FindItem(id int) (*Item, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if item, ok := m.ItemData[id]; ok {
		return &item, nil
	}
	return nil, errors.New("item not found")
}

func (m *MockInMemDB) FindOrders(id int) (*Orders, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if order, ok := m.Orders[id]; ok {
		return &order, nil
	}
	return nil, errors.New("order not found")
}

func (m *MockInMemDB) DeleteUser(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.UserData, id)
	return nil
}

func (m *MockInMemDB) DeleteItem(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.ItemData, id)
	return nil
}

func (m *MockInMemDB) DeleteOrders(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Orders, id)
	return nil
}

func (m *MockInMemDB) UpdateUser(user User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.UserData[user.ID] = user
	return nil
}

func (m *MockInMemDB) UpdateItem(item Item) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ItemData[item.ID] = item
	return nil
}

func (m *MockInMemDB) UpdateOrders(order Orders) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Orders[order.ID] = order
	return nil
}

var (
	userIDCounter  int
	itemIDCounter  int
	orderIDCounter int
	idMutex        sync.Mutex
)

func generateUniqueUserID() int {
	idMutex.Lock()
	defer idMutex.Unlock()
	userIDCounter++
	return userIDCounter
}

func generateUniqueItemID() int {
	idMutex.Lock()
	defer idMutex.Unlock()
	itemIDCounter++
	return itemIDCounter
}

func generateUniqueOrderID() int {
	idMutex.Lock()
	defer idMutex.Unlock()
	orderIDCounter++
	return orderIDCounter
}
