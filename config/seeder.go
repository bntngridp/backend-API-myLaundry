package config

import (
	"log"
	"time"

	"github.com/raihansyahrin/backend_laundry_app.git/models"
	"golang.org/x/crypto/bcrypt"
)

func SeedDatabase() {
	log.Println("Wiping all database tables for a fresh clean re-seed...")

	// Disable foreign keys check to allow full truncation
	DB.Exec("SET FOREIGN_KEY_CHECKS = 0")
	DB.Exec("TRUNCATE TABLE orders")
	DB.Exec("TRUNCATE TABLE addresses")
	DB.Exec("TRUNCATE TABLE services")
	DB.Exec("TRUNCATE TABLE users")
	DB.Exec("TRUNCATE TABLE login_histories")
	DB.Exec("TRUNCATE TABLE promos")
	DB.Exec("SET FOREIGN_KEY_CHECKS = 1")

	log.Println("All tables successfully wiped.")

	// Hash password for seed users
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash seeder passwords:", err)
	}

	// 1. Seed 3 Admins
	log.Println("Seeding Admins (adminsatu, admindua, admintiga)...")
	adminSatu := models.User{
		Username:    "adminsatu",
		Email:       "adminsatu@mylaundry.com",
		PhoneNumber: "081234567890",
		Password:    string(hashedPassword),
		Role:        "admin",
	}
	adminDua := models.User{
		Username:    "admindua",
		Email:       "admindua@mylaundry.com",
		PhoneNumber: "081234567891",
		Password:    string(hashedPassword),
		Role:        "admin",
	}
	adminTiga := models.User{
		Username:    "admintiga",
		Email:       "admintiga@mylaundry.com",
		PhoneNumber: "081234567892",
		Password:    string(hashedPassword),
		Role:        "admin",
	}
	DB.Create(&adminSatu)
	DB.Create(&adminDua)
	DB.Create(&adminTiga)

	// 2. Seed 3 Couriers & 3 Customers
	log.Println("Seeding Couriers (kurirsatu, kurirdua, kurirtiga) & Customers (customersatu, customerdua, customertiga)...")

	// Couriers
	kurirSatu := models.User{
		Username:         "kurirsatu",
		Email:            "kurirsatu@mylaundry.com",
		PhoneNumber:      "081234567896",
		Password:         string(hashedPassword),
		Role:             "courier",
		CreatedByAdminID: &adminSatu.ID,
	}
	kurirDua := models.User{
		Username:         "kurirdua",
		Email:            "kurirdua@mylaundry.com",
		PhoneNumber:      "081234567897",
		Password:         string(hashedPassword),
		Role:             "courier",
		CreatedByAdminID: &adminDua.ID,
	}
	kurirTiga := models.User{
		Username:         "kurirtiga",
		Email:            "kurirtiga@mylaundry.com",
		PhoneNumber:      "081234567898",
		Password:         string(hashedPassword),
		Role:             "courier",
		CreatedByAdminID: &adminTiga.ID,
	}
	DB.Create(&kurirSatu)
	DB.Create(&kurirDua)
	DB.Create(&kurirTiga)

	// Customers
	customerSatu := models.User{
		Username:         "customersatu",
		Email:            "customersatu@mylaundry.com",
		PhoneNumber:      "081234567893",
		Password:         string(hashedPassword),
		Role:             "customer",
		CreatedByAdminID: &adminSatu.ID,
	}
	customerDua := models.User{
		Username:         "customerdua",
		Email:            "customerdua@mylaundry.com",
		PhoneNumber:      "081234567894",
		Password:         string(hashedPassword),
		Role:             "customer",
		CreatedByAdminID: &adminDua.ID,
	}
	customerTiga := models.User{
		Username:         "customertiga",
		Email:            "customertiga@mylaundry.com",
		PhoneNumber:      "081234567895",
		Password:         string(hashedPassword),
		Role:             "customer",
		CreatedByAdminID: &adminTiga.ID,
	}
	DB.Create(&customerSatu)
	DB.Create(&customerDua)
	DB.Create(&customerTiga)

	// 3. Seed Services (Products)
	log.Println("Seeding services...")
	services := []models.Service{
		{
			Title:    "Wash & Fold Regular",
			Time:     48,
			Price:    6000,
			Category: "Kiloan",
			AdminID:  &adminSatu.ID,
		},
		{
			Title:    "Ironing Only Regular",
			Time:     24,
			Price:    4000,
			Category: "Kiloan",
			AdminID:  &adminSatu.ID,
		},
		{
			Title:    "Wash & Iron Express",
			Time:     12,
			Price:    10000,
			Category: "Kiloan",
			AdminID:  &adminSatu.ID,
		},
		{
			Title:    "Suit Jacket Dry Clean",
			Time:     72,
			Price:    15000,
			Category: "Satuan",
			AdminID:  &adminDua.ID,
		},
		{
			Title:    "Blanket Single Wash",
			Time:     48,
			Price:    12000,
			Category: "Satuan",
			AdminID:  &adminDua.ID,
		},
		{
			Title:    "Carpet Wash Premium",
			Time:     120,
			Price:    25000,
			Category: "Satuan",
			AdminID:  &adminTiga.ID,
		},
	}
	for _, s := range services {
		DB.Create(&s)
	}

	// 4. Seed Addresses
	log.Println("Seeding addresses...")
	addrSatu := models.Address{
		CustomerID:    customerSatu.ID,
		ReceiverName:  "Customer Satu (Rumah)",
		PhoneNumber:   "081234567893",
		HouseNumber:   "No. 1A",
		ResidenceName: "Pondok Sukses",
		AddressNotes:  "Pagar hitam samping warung",
		StreetName:    "Jl. Telekomunikasi No. 1",
		District:      "Bojongsoang",
		SubDistrict:   "Sukapura",
		City:          "Bandung",
		Area:          "Bojongsoang",
	}
	addrDua := models.Address{
		CustomerID:    customerDua.ID,
		ReceiverName:  "Customer Dua (Kos)",
		PhoneNumber:   "081234567894",
		HouseNumber:   "No. 2B",
		ResidenceName: "Kos Sukabirus",
		AddressNotes:  "Lantai 2 Kamar 204",
		StreetName:    "Jl. Sukabirus No. 42",
		District:      "Bojongsoang",
		SubDistrict:   "Sukapura",
		City:          "Bandung",
		Area:          "Bojongsoang",
	}
	addrTiga := models.Address{
		CustomerID:    customerTiga.ID,
		ReceiverName:  "Customer Tiga (Kantor)",
		PhoneNumber:   "081234567895",
		HouseNumber:   "No. 3C",
		ResidenceName: "Gedung Podomoro",
		AddressNotes:  "Lobby Depan Pos Satpam",
		StreetName:    "Jl. Bojongsoang Raya No. 88",
		District:      "Bojongsoang",
		SubDistrict:   "Bojongsoang",
		City:          "Bandung",
		Area:          "Bojongsoang",
	}
	DB.Create(&addrSatu)
	DB.Create(&addrDua)
	DB.Create(&addrTiga)

	// 5. Seed Orders
	log.Println("Seeding orders...")
	var service1 models.Service
	DB.First(&service1)

	orderSatu := models.Order{
		CustomerID: customerSatu.ID,
		CourierID:  &kurirSatu.ID,
		ServiceID:  service1.ID,
		AddressID:  addrSatu.ID,
		Weight:     3.5,
		TotalPrice: 3.5 * service1.Price,
		Status:     "diproses",
		AdminID:    &adminSatu.ID,
	}
	orderDua := models.Order{
		CustomerID: customerDua.ID,
		CourierID:  &kurirDua.ID,
		ServiceID:  service1.ID,
		AddressID:  addrDua.ID,
		Weight:     5.0,
		TotalPrice: 5.0 * service1.Price,
		Status:     "penjemputan",
		AdminID:    &adminDua.ID,
	}
	orderTiga := models.Order{
		CustomerID: customerTiga.ID,
		CourierID:  &kurirTiga.ID,
		ServiceID:  service1.ID,
		AddressID:  addrTiga.ID,
		Weight:     2.0,
		TotalPrice: 2.0 * service1.Price,
		Status:     "selesai",
		AdminID:    &adminTiga.ID,
	}
	DB.Create(&orderSatu)
	DB.Create(&orderDua)
	DB.Create(&orderTiga)

	// 6. Seed Promos (promosatu, promodua, promotiga)
	log.Println("Seeding Promos (promosatu, promodua, promotiga)...")
	futureExpiry := time.Now().AddDate(0, 1, 0) // 1 month from now
	promos := []models.Promo{
		{
			Code:               "promosatu",
			Title:              "Diskon 30% Hemat Laundry",
			Subtitle:           "Diskon hingga Rp 5.000 untuk semua layanan",
			DiscountPercentage: 30,
			MaxDiscountAmount:  5000,
			MinOrderAmount:     15000,
			IsActive:           true,
			ExpiredAt:          &futureExpiry,
		},
		{
			Code:               "promodua",
			Title:              "Diskon 50% Super Hemat",
			Subtitle:           "Diskon hingga Rp 10.000 untuk cucian kiloan",
			DiscountPercentage: 50,
			MaxDiscountAmount:  10000,
			MinOrderAmount:     25000,
			IsActive:           true,
			ExpiredAt:          &futureExpiry,
		},
		{
			Code:               "promotiga",
			Title:              "Free Delivery Promo",
			Subtitle:           "Khusus pengguna baru myLaundry",
			DiscountPercentage: 100,
			MaxDiscountAmount:  10000,
			MinOrderAmount:     10000,
			IsActive:           true,
			ExpiredAt:          &futureExpiry,
		},
	}
	for _, p := range promos {
		DB.Create(&p)
	}

	log.Println("Database fresh re-seeding completed.")
}
