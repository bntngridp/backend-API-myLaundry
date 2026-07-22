# myLaundry Backend API 🧺

Golang-based RESTful API designed as the backend power engine for the **myLaundry** enterprise system. Built with modular layers, GORM ORM, Gin framework, and MySQL database. It implements strict **multi-tenant (Admin) isolation** — ensuring each Admin manages their own independent pool of products, customers, couriers, and orders.

---

## 🛠️ Tech Stack & Key Libraries

*   **Language**: Go (Golang)
*   **Framework**: [Gin Gonic](https://github.com/gin-gonic/gin) (HTTP Web Framework)
*   **Database ORM**: [GORM](https://gorm.io/) (Object Relational Mapping for MySQL)
*   **Authentication**: JWT (JSON Web Tokens) & [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) (Secure password hashing)
*   **API Documentation**: Swagger (injected via docs)

---

## ⚙️ Features & Architecture

1.  **Multi-Tenant Admin Isolation**: All core entities (Services, Customers, Couriers, Orders) are linked to a specific `AdminID` or `CreatedByAdminID`. Data queries automatically filter based on the logged-in Admin's ID.
2.  **Order Lifecycle Manager**: Complete workflow tracking from order placement, courier assignment, courier arrival, processing (in progress), to completion.
3.  **Automatic Database Migrations & Seeding**:
    *   On start, GORM automatically migrates models (`User`, `Address`, `Service`, `Order`).
    *   Wipes and seeds fresh dummy data (10+ items per role/section for Admin 2) for rapid local development.
4.  **CORS Enabled**: Configured to work smoothly with local development servers (e.g. `http://localhost:3000`).

---

## 🚀 Getting Started (Local Development)

### Prerequisites
*   Go (v1.18 or higher)
*   MySQL Server (running on port `3306` or `3308`)

### Setup Configuration
1.  Copy the example environment file:
    ```bash
    cp .env.example .env
    ```
2.  Open `.env` and configure your local MySQL credentials:
    ```ini
    PORT=8083
    GIN_MODE=debug
    DB_DSN=root:@tcp(127.0.0.1:3306)/mylaundry?charset=utf8mb4&parseTime=True&loc=Local
    JWT_SECRET=your_super_secret_jwt_key
    ```
3.  Ensure the target database (e.g. `backend_laundry_app`) exists in your MySQL server.

### Running the Server
Start the API server:
```bash
go run main.go
```
The server will boot up, wipe/re-seed the tables, and start listening on the port configured in `.env` (default is `:8083`).

---

## 🌐 API Endpoint Map

### 🔑 Authentication (`/api/auth`)
*   `POST /api/auth/register` - Create new admin account.
*   `POST /api/auth/login` - Authenticate account and receive JWT token.
*   `GET /api/auth/me` - Retrieve logged-in profile data (Requires Token).

### 🛍️ Orders (`/api/orders`)
*   `GET /api/orders/` - Retrieve active orders for the logged-in admin (Tenant Isolated).
*   `POST /api/orders/` - Create a new order (Customer).
*   `POST /api/orders/order-complete` - Mark order status as completed/done (Admin).
*   `POST /api/orders/accept/:id` - Courier accepts/picks up order.
*   `POST /api/orders/courier-arrived` - Courier arrives at customer residence.
*   `POST /api/orders/accept-cash-payment` - Process cash collection and transition order to `in progress`.
*   `POST /api/orders/order-delivery` - Courier delivers clean clothes.

### 👔 Services / Products (`/api/services`)
*   `GET /api/services/` - List all products for the logged-in admin.
*   `POST /api/services/` - Create a new product (Admin).
*   `GET /api/services/:id` - Retrieve product details (Isolated).
*   `PUT /api/services/:id` - Edit product details.
*   `DELETE /api/services/:id` - Delete product.

### 🚚 Couriers (`/api/couriers`) & 👥 Customers (`/api/customers`)
*   `GET /api/couriers/` - List all couriers registered by the logged-in admin.
*   `GET /api/customers/` - List all customers registered by the logged-in admin.

---

## 📂 Project Structure

```
├── config/             # Database connection, GORM setup, and Seeder script
├── controllers/        # Business logic handlers (Delivery / Controller layer)
├── models/             # GORM models (Domain entities / Database schemas)
├── routes/             # Gin route mappings and CORS middlewares
├── utils/              # Helper utilities (JWT generation, password hashing)
├── main.go             # Entrypoint bootstrap file
```
