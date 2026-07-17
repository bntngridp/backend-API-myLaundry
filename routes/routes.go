package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/raihansyahrin/backend_laundry_app.git/controllers"
	admin_controllers "github.com/raihansyahrin/backend_laundry_app.git/controllers/admin"
	courier_controllers "github.com/raihansyahrin/backend_laundry_app.git/controllers/courier"
	customer_controller "github.com/raihansyahrin/backend_laundry_app.git/controllers/customer"
	"github.com/raihansyahrin/backend_laundry_app.git/middlewares"
)

func SetupRoutes(router *gin.Engine) {
	authRoutes := router.Group("api/auth")
	{
		authRoutes.POST("/register", controllers.Register)
		authRoutes.POST("/login", controllers.Login)
	}

	userRoutes := router.Group("api/users")
	userRoutes.Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin"))
	{
		userRoutes.GET("/", controllers.GetUsers)
		userRoutes.GET("/:id", controllers.GetUser)
		userRoutes.POST("/", controllers.CreateUser)
		userRoutes.PUT("/:id", controllers.UpdateUser)
		userRoutes.DELETE("/:id", controllers.DeleteUser)
	}

	customerGroup := router.Group("api/customers")
	{
		customerGroup.GET("/", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("customer"), customer_controller.GetCustomers)
		customerGroup.GET("/:id", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("customer"), customer_controller.GetCustomer)
		customerGroup.PUT("/:id", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("customer"), customer_controller.UpdateCustomer)
		customerGroup.DELETE("/:id", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("customer"), customer_controller.DeleteCustomer)
	}

	courierGroup := router.Group("api/couriers")
	{
		courierGroup.GET("/", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("couriers"), courier_controllers.GetCouriers)
		courierGroup.GET("/:id", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("couriers"), courier_controllers.GetCourier)
		courierGroup.PUT("/:id", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("couriers"), courier_controllers.UpdateCourier)
		courierGroup.DELETE("/:id", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("couriers"), courier_controllers.DeleteCourier)
	}

	adminGroup := router.Group("api/admins")
	{
		adminGroup.GET("/", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin"), admin_controllers.GetAdmins)
		adminGroup.GET("/:id", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin"), admin_controllers.GetAdmin)
		adminGroup.PUT("/:id", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin"), admin_controllers.UpdateAdmin)
		adminGroup.DELETE("/:id", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin"), admin_controllers.DeleteAdmin)
	}

	orderRoutes := router.Group("api/orders")
	{
		orderRoutes.GET("/", middlewares.AuthMiddleware(), controllers.GetOrders)
		orderRoutes.PUT("/status", middlewares.AuthMiddleware(), controllers.UpdateOrderStatus)
		orderRoutes.DELETE("/", middlewares.AuthMiddleware(), controllers.DeleteOrder)
		orderRoutes.POST("/payment", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("customer", "courier"), customer_controller.ProcessPayment)

		//Customer
		orderRoutes.POST("/", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("customer"), customer_controller.CreateOrder)
		orderRoutes.GET("/:customer_id", middlewares.AuthMiddleware(), customer_controller.GetOrderDetailForCustomer)

		//Courier
		orderRoutes.POST("/accept/:id", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("courier"), courier_controllers.AcceptOrder)
		orderRoutes.POST("/courier-arrived", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("courier"), courier_controllers.CourierArrived)
		orderRoutes.POST("/accept-cash-payment", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("courier"), courier_controllers.AcceptCashPayment)
		orderRoutes.POST("/order-delivery", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("courier"), courier_controllers.OrderDelivery)

		//Admin
		orderRoutes.POST("/order-complete", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin"), admin_controllers.OrderComplete)
	}

	serviceRoutes := router.Group("api/services")
	{
		serviceController := &admin_controllers.ServiceController{}
		serviceRoutes.GET("/", serviceController.GetServices)
		serviceRoutes.GET("/:id", serviceController.GetServiceByID)
		serviceRoutes.POST("/", serviceController.CreateService)
		serviceRoutes.PUT("/:id", serviceController.UpdateService)
		serviceRoutes.DELETE("/:id", serviceController.DeleteService)
		serviceRoutes.GET("/category/:category", serviceController.GetServiceByCategory)
	}

	addressRoutes := router.Group("api/addresses")
	{
		addressController := &customer_controller.AddressController{}
		addressRoutes.POST("/", middlewares.AuthMiddleware(), addressController.CreateAddress)
		addressRoutes.GET("/user/:user_id", middlewares.AuthMiddleware(), addressController.GetAddressesByUserID)
		addressRoutes.PUT("/:id", middlewares.AuthMiddleware(), addressController.UpdateAddress)
		addressRoutes.DELETE("/:id", middlewares.AuthMiddleware(), addressController.DeleteAddress)
	}
}
