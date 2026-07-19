package config

import (
	"log"

	"github.com/raihansyahrin/backend_laundry_app.git/models"
	"golang.org/x/crypto/bcrypt"
)

func SeedDatabase() {
	log.Println("Starting database seeding...")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash seeder passwords:", err)
	}

	// 1. Seed Users (Ensure 2 Admins, 2 Couriers, 2 Customers)
	seededUsers := []models.User{
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

	for _, user := range seededUsers {
		var existingUser models.User
		if err := DB.Where("email = ?", user.Email).First(&existingUser).Error; err != nil {
			log.Println("Creating user:", user.Email)
			if err := DB.Create(&user).Error; err != nil {
				log.Println("Failed to seed user:", user.Email, err)
			}
		}
	}

	// 2. Seed Services
	seededServices := []models.Service{
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

	for _, service := range seededServices {
		var existingService models.Service
		if err := DB.Where("title = ?", service.Title).First(&existingService).Error; err != nil {
			log.Println("Creating service:", service.Title)
			if err := DB.Create(&service).Error; err != nil {
				log.Println("Failed to seed service:", service.Title, err)
			}
		}
	}

	// 3. Seed Addresses for Customers
	var customer1, customer2 models.User
	_ = DB.Where("email = ?", "customer@mylaundry.com").First(&customer1).Error
	_ = DB.Where("email = ?", "customer2@mylaundry.com").First(&customer2).Error

	if customer1.ID != 0 {
		var addrCount int64
		DB.Model(&models.Address{}).Where("customer_id = ?", customer1.ID).Count(&addrCount)
		if addrCount == 0 {
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
		}
	}

	if customer2.ID != 0 {
		var addrCount int64
		DB.Model(&models.Address{}).Where("customer_id = ?", customer2.ID).Count(&addrCount)
		if addrCount == 0 {
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
		}
	}

	// 4. Seed Orders (Recreate to ensure comprehensive coverage across states)
	var orderCount int64
	DB.Model(&models.Order{}).Count(&orderCount)
	if orderCount <= 2 {
		log.Println("Populating orders for all sections...")
		// Clear existing small seeds to prevent collisions or messy duplicates
		DB.Exec("DELETE FROM orders")

		// Re-fetch objects for fresh reference
		var s1, s2, s3, s4, s5 models.Service
		DB.Where("title = ?", "Wash & Fold Regular").First(&s1)
		DB.Where("title = ?", "Ironing Only Regular").First(&s2)
		DB.Where("title = ?", "Wash & Iron Express").First(&s3)
		DB.Where("title = ?", "Suit Jacket Dry Clean").First(&s4)
		DB.Where("title = ?", "Blanket Single Wash").First(&s5)

		var a1, a2 models.Address
		DB.Where("customer_id = ?", customer1.ID).First(&a1)
		DB.Where("customer_id = ?", customer2.ID).First(&a2)

		var courier1, courier2 models.User
		DB.Where("email = ?", "courier@mylaundry.com").First(&courier1)
		DB.Where("email = ?", "courier2@mylaundry.com").First(&courier2)

		orders := []models.Order{
			{
				CustomerID: customer1.ID,
				ServiceID:  s1.ID,
				AddressID:  a1.ID,
				Weight:     5.0,
				TotalPrice: 5.0 * s1.Price,
				Status:     "menunggu pembayaran", // Active order (unpaid)
			},
			{
				CustomerID: customer1.ID,
				ServiceID:  s3.ID,
				AddressID:  a1.ID,
				CourierID:  &courier1.ID,
				Weight:     3.5,
				TotalPrice: 3.5 * s3.Price,
				Status:     "in progress", // Active order (processing)
			},
			{
				CustomerID: customer2.ID,
				ServiceID:  s4.ID,
				AddressID:  a2.ID,
				CourierID:  &courier1.ID,
				Quantity:   2,
				TotalPrice: 2.0 * s4.Price,
				Status:     "completed", // History order (completed)
			},
			{
				CustomerID: customer2.ID,
				ServiceID:  s5.ID,
				AddressID:  a2.ID,
				Quantity:   1,
				TotalPrice: 1.0 * s5.Price,
				Status:     "cancelled", // History order (cancelled)
			},
			{
				CustomerID: customer1.ID,
				ServiceID:  s2.ID,
				AddressID:  a1.ID,
				CourierID:  &courier2.ID,
				Weight:     4.0,
				TotalPrice: 4.0 * s2.Price,
				Status:     "courier en route", // Active order (delivering)
			},
		}

		for _, order := range orders {
			if err := DB.Create(&order).Error; err != nil {
				log.Println("Failed to seed order:", err)
			}
		}
	}

	log.Println("Database seeding completed.")
}
