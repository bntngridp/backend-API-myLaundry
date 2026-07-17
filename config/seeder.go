package config

import (
	"log"

	"github.com/raihansyahrin/backend_laundry_app.git/models"
	"golang.org/x/crypto/bcrypt"
)

func SeedDatabase() {
	log.Println("Starting database seeding...")

	// 1. Seed Users
	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count == 0 {
		log.Println("Seeding users...")
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal("Failed to hash seeder passwords:", err)
		}

		users := []models.User{
			{
				Username: "Admin Laundry",
				Email:    "admin@mylaundry.com",
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
				Username: "Budi Customer",
				Email:    "customer@mylaundry.com",
				Password: string(hashedPassword),
				Role:     "customer",
			},
		}

		for _, user := range users {
			if err := DB.Create(&user).Error; err != nil {
				log.Println("Failed to seed user:", user.Email, err)
			}
		}
	}

	// 2. Seed Services
	DB.Model(&models.Service{}).Count(&count)
	if count == 0 {
		log.Println("Seeding services...")
		services := []models.Service{
			{
				Title:    "Wash & Fold Regular",
				Time:     48, // hours
				Price:    6000,
				Category: "Kiloan",
			},
			{
				Title:    "Ironing Only Regular",
				Time:     24, // hours
				Price:    4000,
				Category: "Kiloan",
			},
			{
				Title:    "Wash & Iron Express",
				Time:     12, // hours
				Price:    10000,
				Category: "Kiloan",
			},
			{
				Title:    "Suit Jacket Dry Clean",
				Time:     72, // hours
				Price:    15000,
				Category: "Satuan",
			},
			{
				Title:    "Blanket Single Wash",
				Time:     48, // hours
				Price:    12000,
				Category: "Satuan",
			},
		}

		for _, service := range services {
			if err := DB.Create(&service).Error; err != nil {
				log.Println("Failed to seed service:", service.Title, err)
			}
		}
	}

	// Get Customer for linking addresses and orders
	var customer models.User
	if err := DB.Where("role = ?", "customer").First(&customer).Error; err == nil {
		// 3. Seed Addresses
		DB.Model(&models.Address{}).Count(&count)
		if count == 0 {
			log.Println("Seeding addresses...")
			address := models.Address{
				CustomerID:    customer.ID,
				ReceiverName:  "Budi Utomo",
				PhoneNumber:   "081234567890",
				HouseNumber:   "No. 45",
				ResidenceName: "Kompleks Telkom University",
				AddressNotes:  "Pagar warna hitam, samping warung",
				StreetName:    "Jl. Telekomunikasi 1",
				District:      "Sukapura",
				SubDistrict:   "Dayeuhkolot",
				City:          "Bandung",
				Area:          "Bojongsoang",
			}
			if err := DB.Create(&address).Error; err != nil {
				log.Println("Failed to seed address:", err)
			}
		}

		// 4. Seed Orders
		DB.Model(&models.Order{}).Count(&count)
		if count == 0 {
			log.Println("Seeding orders...")
			// Get default service & address
			var service models.Service
			var address models.Address
			var courier models.User

			_ = DB.First(&service).Error
			_ = DB.First(&address).Error
			_ = DB.Where("role = ?", "courier").First(&courier).Error

			orders := []models.Order{
				{
					CustomerID: customer.ID,
					ServiceID:  service.ID,
					AddressID:  address.ID,
					Weight:     5.0,
					TotalPrice: 5.0 * service.Price,
					Status:     "menunggu pembayaran",
				},
				{
					CustomerID: customer.ID,
					ServiceID:  service.ID,
					AddressID:  address.ID,
					CourierID:  &courier.ID,
					Weight:     7.2,
					TotalPrice: 7.2 * service.Price,
					Status:     "in progress",
				},
			}

			for _, order := range orders {
				if err := DB.Create(&order).Error; err != nil {
					log.Println("Failed to seed order:", err)
				}
			}
		}
	}

	log.Println("Database seeding completed.")
}
