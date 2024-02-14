# Savannah

This is a minimal order api using OpenID Connect 

## Prerequisite
- go
- docker
- make
- migrate

## To get started:
```
git clone https://github.com/wxmbugu/savannah
``` 
build :
```
go build ./cmd/app/
```
setp postgres local:
```
make postgres
```
creatdb :
```
make creatdb
```
migrations:
```
make migrateup
```
run server:
```
./app
```

env setup:
```
 - Check the .env.example for reference
```

#### Routes
```
1. Authentication
1.1 Login

    URI: /login
    Method: GET, OPTIONS
    Description: Initiates the OpenID Connect login process and sets a callback cookie.

1.2 Google Callback

    URI: /auth/google/callback
    Method: GET, OPTIONS
    Description: Handles the callback from Google authentication.

2. Authenticated Routes

All authenticated routes are under the /v1 prefix and require authentication through OpenID Connect.
2.1 Create Customer

    URI: /v1/customers
    Method: POST, OPTIONS
    Description: Creates a new customer.

2.2 Get Customer

    URI: /v1/customers/{id}
    Method: GET, OPTIONS
    Description: Retrieves customer details based on the provided id.

2.3 Create Order

    URI: /v1/orders
    Method: POST, OPTIONS
    Description: Creates a new order.

2.4 Get Order

    URI: /v1/orders/{id}
    Method: GET, OPTIONS
    Description: Retrieves order details based on the provided id.

2.5 Get Item

    URI: /v1/items/{id}
    Method: GET, OPTIONS
    Description: Retrieves item details based on the provided id.

3. Additional Information
3.1 Routes Information

    URI: /api/v1/routes/{stopId}[/{type}/]
    Method: GET
    Description: Retrieves a list of routes that stop at a specific stop. The optional type parameter controls whether this returns the full route info or just a list of short names.
```



