package main

import (
	"log"
	"time"

	_ "gw-currncy-wallet/docs"
	"gw-currncy-wallet/internal/auth"
	"gw-currncy-wallet/internal/changer"
	"gw-currncy-wallet/internal/handlers"
	"gw-currncy-wallet/internal/storages/postgres"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"google.golang.org/grpc/credentials/insecure"

	proto_exchange "github.com/apelsinkoo09/proto-exchange/exchange"
)

func main() {
	db, err := postgres.Connection()
	if err != nil {
		log.Println(err)
	}
	defer db.Close()

	grpcConn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer grpcConn.Close()

	grpcClient := proto_exchange.NewExchangeServiceClient(grpcConn)

	cache := changer.RateCache(5*time.Minute, 10*time.Minute)

	exchangerClient := changer.NewExchangerClient(grpcClient, cache)

	storage := &postgres.StorageConn{DB: db}
	walletService := handlers.NewWalletService(storage, exchangerClient)
	userService := handlers.NewUserService(storage)

	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.POST("/api/v1/login", userService.LoginHandler)
	r.POST("/api/v1/register", userService.RegisterHandler)
	r.GET("/api/v1/getUser", userService.GetUserDataHandler)

	protected := r.Group("/api/v1")
	protected.Use(auth.JWTMiddleware())
	{
		protected.GET("/balance", walletService.GetBalanceHandler)
		protected.POST("/wallet/deposit", walletService.DepositHandler)
		protected.POST("/wallet/withdraw", walletService.WithdrawHandler)
		protected.POST("/wallet/exchange", walletService.ExchangeHandler)
	}

	log.Println("Starting server on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
