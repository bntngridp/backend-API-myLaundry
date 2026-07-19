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

	// 1. Seed Admins First
	log.Println("Seeding Admins...")
	admin1 := models.User{
		Username: "Admin Laundry Kesatu",
		Email:    "admin@mylaundry.com",
		Password: string(hashedPassword),
		Role:     "admin",
	}
	admin2 := models.User{
		Username: "Admin Laundry Kedua",
		Email:    "admin2@mylaundry.com",
		Password: string(hashedPassword),
		Role:     "admin",
	}
	DB.Create(&admin1)
	DB.Create(&admin2)

	// 2. Seed Couriers & Customers with CreatedByAdminID
	log.Println("Seeding Couriers and Customers...")
	couriersAndCustomers := []models.User{
		{
			Username:         "Kurir Bagus",
			Email:            "courier@mylaundry.com",
			Password:         string(hashedPassword),
			Role:             "courier",
			CreatedByAdminID: &admin1.ID,
		},
		{
			Username:         "Kurir Amanah",
			Email:            "courier2@mylaundry.com",
			Password:         string(hashedPassword),
			Role:             "courier",
			CreatedByAdminID: &admin2.ID,
		},
		{
			Username:         "Budi Customer",
			Email:            "customer@mylaundry.com",
			Password:         string(hashedPassword),
			Role:             "customer",
			CreatedByAdminID: &admin1.ID,
		},
		{
			Username:         "Ahmad Customer",
			Email:            "customer2@mylaundry.com",
			Password:         string(hashedPassword),
			Role:             "customer",
			CreatedByAdminID: &admin2.ID,
		},
	}
	for _, u := range couriersAndCustomers {
		DB.Create(&u)
	}

	// Re-fetch customer users to set up addresses
	var customer1, customer2 models.User
	DB.Where("email = ?", "customer@mylaundry.com").First(&customer1)
	DB.Where("email = ?", "customer2@mylaundry.com").First(&customer2)

	var courier1, courier2 models.User
	DB.Where("email = ?", "courier@mylaundry.com").First(&courier1)
	DB.Where("email = ?", "courier2@mylaundry.com").First(&courier2)

	// 3. Seed Services (Products) with AdminID
	log.Println("Seeding services...")
	services := []models.Service{
		{
			Title:    "Wash & Fold Regular",
			Time:     48,
			Price:    6000,
			Category: "Kiloan",
			AdminID:  &admin1.ID,
		},
		{
			Title:    "Ironing Only Regular",
			Time:     24,
			Price:    4000,
			Category: "Kiloan",
			AdminID:  &admin1.ID,
		},
		{
			Title:    "Wash & Iron Express",
			Time:     12,
			Price:    10000,
			Category: "Kiloan",
			AdminID:  &admin1.ID,
		},
		{
			Title:    "Suit Jacket Dry Clean",
			Time:     72,
			Price:    15000,
			Category: "Satuan",
			AdminID:  &admin2.ID,
		},
		{
			Title:    "Blanket Single Wash",
			Time:     48,
			Price:    12000,
			Category: "Satuan",
			AdminID:  &admin2.ID,
		},
	}
	for _, s := range services {
		DB.Create(&s)
	}

	// 4. Seed Addresses
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

	// Re-fetch services for orders
	var s1, s2, s3, s4, s5 models.Service
	DB.Where("title = ? AND admin_id = ?", "Wash & Fold Regular", admin1.ID).First(&s1)
	DB.Where("title = ? AND admin_id = ?", "Ironing Only Regular", admin1.ID).First(&s2)
	DB.Where("title = ? AND admin_id = ?", "Wash & Iron Express", admin1.ID).First(&s3)
	DB.Where("title = ? AND admin_id = ?", "Suit Jacket Dry Clean", admin2.ID).First(&s4)
	DB.Where("title = ? AND admin_id = ?", "Blanket Single Wash", admin2.ID).First(&s5)

	// 5. Seed Orders
	log.Println("Seeding orders...")
	orders := []models.Order{
		{
			CustomerID: customer1.ID,
			ServiceID:  s1.ID,
			AddressID:  address1.ID,
			Weight:     5.0,
			TotalPrice: 5.0 * s1.Price,
			Status:     "menunggu pembayaran", // Active (unpaid) - visible to Admin 1
			AdminID:    &admin1.ID,
		},
		{
			CustomerID: customer1.ID,
			ServiceID:  s3.ID,
			AddressID:  address1.ID,
			CourierID:  &courier1.ID,
			Weight:     3.5,
			TotalPrice: 3.5 * s3.Price,
			Status:     "in progress", // Active (processing) - visible to Admin 1
			AdminID:    &admin1.ID,
		},
		{
			CustomerID: customer2.ID,
			ServiceID:  s4.ID,
			AddressID:  address2.ID,
			CourierID:  &courier2.ID,
			AdminID:    &admin2.ID, // Processed by Admin 2 (Admin Laundry Kedua)
			Quantity:   2,
			TotalPrice: 2.0 * s4.Price,
			Status:     "done", // History (completed) - visible ONLY to Admin 2
		},
		{
			CustomerID: customer2.ID,
			ServiceID:  s5.ID,
			AddressID:  address2.ID,
			AdminID:    &admin2.ID,
			Quantity:   1,
			TotalPrice: 1.0 * s5.Price,
			Status:     "cancelled", // History (cancelled) - visible to Admin 2
		},
		{
			CustomerID: customer1.ID,
			ServiceID:  s2.ID,
			AddressID:  address1.ID,
			CourierID:  &courier1.ID,
			Weight:     4.0,
			TotalPrice: 4.0 * s2.Price,
			Status:     "courier en route", // Active (delivering) - visible to Admin 1
			AdminID:    &admin1.ID,
		},
	}

	for _, order := range orders {
		if err := DB.Create(&order).Error; err != nil {
			log.Println("Failed to seed order:", err)
		}
	}

	log.Println("Database fresh re-seeding completed.")
}
