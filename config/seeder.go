package config

import (
	"fmt"
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

	// 1. Seed Admins First
	log.Println("Seeding Admins...")
	admin1 := models.User{
		Username:    "Admin Laundry Kesatu",
		Email:       "admin@mylaundry.com",
		PhoneNumber: "081234567890",
		Password:    string(hashedPassword),
		Role:        "admin",
	}
	admin2 := models.User{
		Username:    "Admin Laundry Kedua",
		Email:       "admin2@mylaundry.com",
		PhoneNumber: "081234567891",
		Password:    string(hashedPassword),
		Role:        "admin",
	}
	DB.Create(&admin1)
	DB.Create(&admin2)

	// 2. Seed Couriers & Customers
	log.Println("Seeding Couriers and Customers...")

	// Seed Couriers & Customers for Admin 1 (keep existing ones)
	courier1 := models.User{
		Username:         "Kurir Bagus",
		Email:            "courier@mylaundry.com",
		PhoneNumber:      "081234567892",
		Password:         string(hashedPassword),
		Role:             "courier",
		CreatedByAdminID: &admin1.ID,
	}
	customer1 := models.User{
		Username:         "Budi Customer",
		Email:            "customer@mylaundry.com",
		PhoneNumber:      "081234567893",
		Password:         string(hashedPassword),
		Role:             "customer",
		CreatedByAdminID: &admin1.ID,
	}
	DB.Create(&courier1)
	DB.Create(&customer1)

	// Seed at least 10 Couriers for Admin 2
	couriersAdmin2 := []models.User{}
	courierNames := []string{
		"Kurir Amanah", "Kurir Cepat", "Kurir Tangkas", "Kurir Jujur", "Kurir Handal",
		"Kurir Ramah", "Kurir Santun", "Kurir Sigap", "Kurir Satset", "Kurir Gesit",
	}
	for i, name := range courierNames {
		email := fmt.Sprintf("courier2_%d@mylaundry.com", i+1)
		phone := fmt.Sprintf("0812888800%02d", i+1)
		if i == 0 {
			email = "courier2@mylaundry.com" // keep the existing primary courier
			phone = "081234567894"
		}
		couriersAdmin2 = append(couriersAdmin2, models.User{
			Username:         name,
			Email:            email,
			PhoneNumber:      phone,
			Password:         string(hashedPassword),
			Role:             "courier",
			CreatedByAdminID: &admin2.ID,
		})
	}
	for _, u := range couriersAdmin2 {
		DB.Create(&u)
	}

	// Seed at least 10 Customers for Admin 2
	customersAdmin2 := []models.User{}
	customerNames := []string{
		"Ahmad Customer", "Bambang Customer", "Chandra Customer", "Dedi Customer", "Eko Customer",
		"Fajar Customer", "Gunawan Customer", "Hendra Customer", "Indra Customer", "Joko Customer",
	}
	for i, name := range customerNames {
		email := fmt.Sprintf("customer2_%d@mylaundry.com", i+1)
		phone := fmt.Sprintf("0812999900%02d", i+1)
		if i == 0 {
			email = "customer2@mylaundry.com" // keep the existing primary customer
			phone = "081234567895"
		}
		customersAdmin2 = append(customersAdmin2, models.User{
			Username:         name,
			Email:            email,
			PhoneNumber:      phone,
			Password:         string(hashedPassword),
			Role:             "customer",
			CreatedByAdminID: &admin2.ID,
		})
	}
	for _, u := range customersAdmin2 {
		DB.Create(&u)
	}

	// Re-fetch customer users to set up addresses
	var seededCustomers2 []models.User
	DB.Where("created_by_admin_id = ? AND role = ?", admin2.ID, "customer").Order("id asc").Find(&seededCustomers2)

	var seededCouriers2 []models.User
	DB.Where("created_by_admin_id = ? AND role = ?", admin2.ID, "courier").Order("id asc").Find(&seededCouriers2)

	// 3. Seed Services (Products) with AdminID
	log.Println("Seeding services...")
	
	// Services for Admin 1 (keep existing ones)
	servicesAdmin1 := []models.Service{
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
	}
	for _, s := range servicesAdmin1 {
		DB.Create(&s)
	}

	// Seed at least 10 Services (Products) for Admin 2
	servicesAdmin2 := []models.Service{
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
		{
			Title:    "Carpet Wash Premium",
			Time:     120,
			Price:    25000,
			Category: "Satuan",
			AdminID:  &admin2.ID,
		},
		{
			Title:    "Bed Cover Double Wash",
			Time:     48,
			Price:    20000,
			Category: "Satuan",
			AdminID:  &admin2.ID,
		},
		{
			Title:    "Leather Jacket Care",
			Time:     168,
			Price:    50000,
			Category: "Satuan",
			AdminID:  &admin2.ID,
		},
		{
			Title:    "Curtain Cleaning Regular",
			Time:     96,
			Price:    8000,
			Category: "Satuan",
			AdminID:  &admin2.ID,
		},
		{
			Title:    "Sneakers Deep Clean",
			Time:     72,
			Price:    30000,
			Category: "Satuan",
			AdminID:  &admin2.ID,
		},
		{
			Title:    "Backpack Wash Regular",
			Time:     48,
			Price:    15000,
			Category: "Satuan",
			AdminID:  &admin2.ID,
		},
		{
			Title:    "Sleeping Bag Wash",
			Time:     72,
			Price:    18000,
			Category: "Satuan",
			AdminID:  &admin2.ID,
		},
		{
			Title:    "Baby Stroller Cleaning",
			Time:     120,
			Price:    60000,
			Category: "Satuan",
			AdminID:  &admin2.ID,
		},
	}
	for _, s := range servicesAdmin2 {
		DB.Create(&s)
	}

	var seededServices2 []models.Service
	DB.Where("admin_id = ?", admin2.ID).Order("id asc").Find(&seededServices2)

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

	// Seed Addresses for Admin 2 customers (10 total)
	seededAddresses2 := []models.Address{}
	for i, cust := range seededCustomers2 {
		addr := models.Address{
			CustomerID:    cust.ID,
			ReceiverName:  fmt.Sprintf("Receiver for %s", cust.Username),
			PhoneNumber:   fmt.Sprintf("085721113%03d", i+1),
			HouseNumber:   fmt.Sprintf("Blok C-%d", i+12),
			ResidenceName: "Perumahan Podomoro Land",
			AddressNotes:  "Belok kanan setelah pos satpam utama",
			StreetName:    "Jl. Bojongsoang Raya",
			District:      "Bojongsoang",
			SubDistrict:   "Bojongsoang",
			City:          "Bandung",
			Area:          "Bojongsoang",
		}
		DB.Create(&addr)
		seededAddresses2 = append(seededAddresses2, addr)
	}

	// Re-fetch services for orders
	var s1, s2, s3 models.Service
	DB.Where("title = ? AND admin_id = ?", "Wash & Fold Regular", admin1.ID).First(&s1)
	DB.Where("title = ? AND admin_id = ?", "Ironing Only Regular", admin1.ID).First(&s2)
	DB.Where("title = ? AND admin_id = ?", "Wash & Iron Express", admin1.ID).First(&s3)

	// 5. Seed Orders
	log.Println("Seeding orders...")

	// Orders for Admin 1 (keep existing ones)
	ordersAdmin1 := []models.Order{
		{
			CustomerID: customer1.ID,
			ServiceID:  s1.ID,
			AddressID:  address1.ID,
			Weight:     5.0,
			TotalPrice: 5.0 * s1.Price,
			Status:     "menunggu pembayaran",
			AdminID:    &admin1.ID,
		},
		{
			CustomerID: customer1.ID,
			ServiceID:  s3.ID,
			AddressID:  address1.ID,
			CourierID:  &courier1.ID,
			Weight:     3.5,
			TotalPrice: 3.5 * s3.Price,
			Status:     "in progress",
			AdminID:    &admin1.ID,
		},
		{
			CustomerID: customer1.ID,
			ServiceID:  s2.ID,
			AddressID:  address1.ID,
			CourierID:  &courier1.ID,
			Weight:     4.0,
			TotalPrice: 4.0 * s2.Price,
			Status:     "courier en route",
			AdminID:    &admin1.ID,
		},
	}
	for _, order := range ordersAdmin1 {
		DB.Create(&order)
	}

	// Seed at least 10 Orders for Admin 2
	statuses := []string{
		"done", "cancelled", "done", "cancelled", "done",
		"awaiting payment", "in progress", "courier en route", "pending", "arrived",
	}

	for i := 0; i < 10; i++ {
		cust := seededCustomers2[i%len(seededCustomers2)]
		serv := seededServices2[i%len(seededServices2)]
		addr := seededAddresses2[i%len(seededAddresses2)]
		cour := seededCouriers2[i%len(seededCouriers2)]
		status := statuses[i]

		var courierID *uint
		if status != "pending" && status != "awaiting payment" {
			courierID = &cour.ID
		}

		qty := float64((i % 3) + 1)
		order := models.Order{
			CustomerID: cust.ID,
			ServiceID:  serv.ID,
			AddressID:  addr.ID,
			CourierID:  courierID,
			AdminID:    &admin2.ID,
			Quantity:   int(qty),
			TotalPrice: qty * serv.Price,
			Status:     status,
		}
		if err := DB.Create(&order).Error; err != nil {
			log.Println("Failed to seed Admin 2 order:", err)
		}
	}

	// 5. Seed Promos
	log.Println("Seeding Promos...")
	futureExpiry := time.Now().AddDate(0, 1, 0) // 1 month from now
	promos := []models.Promo{
		{
			Code:               "BersihTanpaPusing",
			Title:              "Diskon 30% Hemat Laundry",
			Subtitle:           "Diskon hingga Rp 5.000 untuk semua layanan",
			DiscountPercentage: 30,
			MaxDiscountAmount:  5000,
			MinOrderAmount:     15000,
			IsActive:           true,
			ExpiredAt:          &futureExpiry,
		},
		{
			Code:               "CucianWangi",
			Title:              "Diskon 50% Super Hemat",
			Subtitle:           "Diskon hingga Rp 10.000 untuk cucian kiloan",
			DiscountPercentage: 50,
			MaxDiscountAmount:  10000,
			MinOrderAmount:     25000,
			IsActive:           true,
			ExpiredAt:          &futureExpiry,
		},
		{
			Code:               "MulaiLaundry",
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
