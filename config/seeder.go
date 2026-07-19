package config

import (
	"log"

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
	DB.Exec("SET FOREIGN_KEY_CHECKS = 1")

	log.Println("All tables successfully wiped.")

	// Hash password for seed users
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash seeder passwords:", err)
	}

	// 1. Seed Users
	log.Println("Seeding users...")
	users := []models.User{
		{
			Username: "Admin Laundry Kesatu",
			Email:    "admin@mylaundry.com",
			Password: string(hashedPassword),
			Role:     "admin",
		},
		{
			Username: "Admin Laundry Kedua",
			Email:    "admin2@mylaundry.com",
			Password: string(hashedPassword),
			Role:     "admin",
		},
		{
			Username: "Kurir Bagus",
			Email:    "courier@mylaundry.com",
			Password: string(hashedPassword),
			Role:     "courier",
		},
		{
			Username: "Kurir Amanah",
			Email:    "courier2@mylaundry.com",
			Password: string(hashedPassword),
			Role:     "courier",
		},
		{
			Username: "Budi Customer",
			Email:    "customer@mylaundry.com",
			Password: string(hashedPassword),
			Role:     "customer",
		},
		{
			Username: "Ahmad Customer",
			Email:    "customer2@mylaundry.com",
			Password: string(hashedPassword),
			Role:     "customer",
		},
	}

	for _, user := range users {
		if err := DB.Create(&user).Error; err != nil {
			log.Println("Failed to seed user:", user.Email, err)
		}
	}

	// 2. Seed Services (Products)
	log.Println("Seeding services...")
	services := []models.Service{
		{
			Title:    "Wash & Fold Regular",
			Time:     48,
			Price:    6000,
			Category: "Kiloan",
		},
		{
			Title:    "Ironing Only Regular",
			Time:     24,
			Price:    4000,
			Category: "Kiloan",
		},
		{
			Title:    "Wash & Iron Express",
			Time:     12,
			Price:    10000,
			Category: "Kiloan",
		},
		{
			Title:    "Suit Jacket Dry Clean",
			Time:     72,
			Price:    15000,
			Category: "Satuan",
		},
		{
			Title:    "Blanket Single Wash",
			Time:     48,
			Price:    12000,
			Category: "Satuan",
		},
	}

	for _, service := range services {
		if err := DB.Create(&service).Error; err != nil {
			log.Println("Failed to seed service:", service.Title, err)
		}
	}

	// Re-fetch users for addresses and orders
	var customer1, customer2 models.User
	DB.Where("email = ?", "customer@mylaundry.com").First(&customer1)
	DB.Where("email = ?", "customer2@mylaundry.com").First(&customer2)

	var courier1, courier2 models.User
	DB.Where("email = ?", "courier@mylaundry.com").First(&courier1)
	DB.Where("email = ?", "courier2@mylaundry.com").First(&courier2)

	var admin1, admin2 models.User
	DB.Where("email = ?", "admin@mylaundry.com").First(&admin1)
	DB.Where("email = ?", "admin2@mylaundry.com").First(&admin2)

	// 3. Seed Addresses
	log.Println("Seeding addresses...")
	address1 := models.Address{
		CustomerID:    customer1.ID,
		ReceiverName:  "Budi Utomo",
		PhoneNumber:   "081234567890",
		HouseNumber:   "No. 45",
		ResidenceName: "Kompleks Telkom University",
		AddressNotes:  "Pagar warna hitam, samping warung makan",
		StreetName:    "Jl. Telekomunikasi 1",
		District:      "Sukapura",
		SubDistrict:   "Dayeuhkolot",
		City:          "Bandung",
		Area:          "Bojongsoang",
	}
	DB.Create(&address1)

	address2 := models.Address{
		CustomerID:    customer2.ID,
		ReceiverName:  "Ahmad Fauzi",
		PhoneNumber:   "085721113444",
		HouseNumber:   "Blok C-12",
		ResidenceName: "Perumahan Podomoro Land",
		AddressNotes:  "Belok kanan setelah pos satpam utama",
		StreetName:    "Jl. Bojongsoang Raya",
		District:      "Bojongsoang",
		SubDistrict:   "Bojongsoang",
		City:          "Bandung",
		Area:          "Bojongsoang",
	}
	DB.Create(&address2)

	// Re-fetch services
	var s1, s2, s3, s4, s5 models.Service
	DB.Where("title = ?", "Wash & Fold Regular").First(&s1)
	DB.Where("title = ?", "Ironing Only Regular").First(&s2)
	DB.Where("title = ?", "Wash & Iron Express").First(&s3)
	DB.Where("title = ?", "Suit Jacket Dry Clean").First(&s4)
	DB.Where("title = ?", "Blanket Single Wash").First(&s5)

	// 4. Seed Orders
	log.Println("Seeding orders...")
	orders := []models.Order{
		{
			CustomerID: customer1.ID,
			ServiceID:  s1.ID,
			AddressID:  address1.ID,
			Weight:     5.0,
			TotalPrice: 5.0 * s1.Price,
			Status:     "menunggu pembayaran", // Active (unpaid) - visible to all admins
		},
		{
			CustomerID: customer1.ID,
			ServiceID:  s3.ID,
			AddressID:  address1.ID,
			CourierID:  &courier1.ID,
			Weight:     3.5,
			TotalPrice: 3.5 * s3.Price,
			Status:     "in progress", // Active (processing) - visible to all admins
		},
		{
			CustomerID: customer2.ID,
			ServiceID:  s4.ID,
			AddressID:  address2.ID,
			CourierID:  &courier1.ID,
			AdminID:    &admin1.ID, // Processed by Admin 1 (Admin Laundry Kesatu)
			Quantity:   2,
			TotalPrice: 2.0 * s4.Price,
			Status:     "done", // History (completed) - visible ONLY to Admin 1
		},
		{
			CustomerID: customer2.ID,
			ServiceID:  s5.ID,
			AddressID:  address2.ID,
			Quantity:   1,
			TotalPrice: 1.0 * s5.Price,
			Status:     "cancelled", // History (cancelled) - visible to all admins (no admin assigned)
		},
		{
			CustomerID: customer1.ID,
			ServiceID:  s2.ID,
			AddressID:  address1.ID,
			CourierID:  &courier2.ID,
			Weight:     4.0,
			TotalPrice: 4.0 * s2.Price,
			Status:     "courier en route", // Active (delivering) - visible to all admins
		},
		{
			CustomerID: customer2.ID,
			ServiceID:  s3.ID,
			AddressID:  address2.ID,
			CourierID:  &courier2.ID,
			AdminID:    &admin2.ID, // Processed by Admin 2 (Admin Laundry Kedua)
			Weight:     6.0,
			TotalPrice: 6.0 * s3.Price,
			Status:     "done", // History (completed) - visible ONLY to Admin 2
		},
	}

	for _, order := range orders {
		if err := DB.Create(&order).Error; err != nil {
			log.Println("Failed to seed order:", err)
		}
	}

	log.Println("Database fresh re-seeding completed.")
}
