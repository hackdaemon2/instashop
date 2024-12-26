package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hackdaemon2/instashop/config"
	_ "github.com/hackdaemon2/instashop/docs"
	"github.com/hackdaemon2/instashop/handler"
	"github.com/hackdaemon2/instashop/middleware"
	"github.com/jinzhu/gorm"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// noRouteOrMethod is a function to handle HTTP method not supported
// for an endpoint or a non-existent URL is hit
func noRouteOrMethod(status int, message string) gin.HandlerFunc {
	return func(context *gin.Context) {
		context.JSON(status, gin.H{"error": true, "message": message})
	}
}

// SetupRouter configures the routes that this application uses
func setupRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()

	router.NoRoute(noRouteOrMethod(http.StatusNotFound, "route not found"))
	router.NoMethod(noRouteOrMethod(http.StatusMethodNotAllowed, "method not allowed"))

	// Serve the Swagger UI and Doc JSON
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiV1 := router.Group("/api/v1")
	apiV1.POST("/user/signup", handler.Signup(db))
	apiV1.POST("/user/login", handler.Login(db))

	authenticated := apiV1.Group("/")
	authenticated.Use(middleware.Authenticate())
	authenticated.POST("/user/order", handler.PlaceOrder(db))
	authenticated.GET("/user/order", handler.GetUserOrders(db))
	authenticated.PUT("/user/order/:order_reference/cancel", handler.CancelUserOrder(db))
	authenticated.GET("/user/product/:product_code", handler.GetProduct(db))

	admin := apiV1.Group("/admin")
	admin.Use(middleware.Authenticate())
	admin.Use(middleware.IsAdmin())
	admin.PUT("/order/:order_reference/status", handler.UpdateOrderStatus(db))
	admin.POST("/product", handler.CreateProduct(db))
	admin.PUT("/product/:product_code", handler.UpdateProduct(db))
	admin.DELETE("/product/:product_code", handler.DeleteProduct(db))

	return router
}

// @title Instashop Swagger API
// @version 1.0
// @description This is a sample server Instashop server.
// @host localhost:3000
// @BasePath /
func main() {
	config.LoadEnv()
	config.ConnectDatabase()

	route := setupRouter(config.DB)

	if err := route.Run(fmt.Sprintf(":%s", config.GetEnv("PORT"))); err != nil {
		log.Fatal("Unable to start server:", err)
	}
}
