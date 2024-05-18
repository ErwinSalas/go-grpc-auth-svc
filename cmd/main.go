package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/ErwinSalas/go-grpc-auth-svc/pkg/auth"
	"github.com/ErwinSalas/go-grpc-auth-svc/pkg/config"
	"github.com/ErwinSalas/go-grpc-auth-svc/pkg/database"
	"github.com/ErwinSalas/go-grpc-auth-svc/pkg/server"
	"github.com/ErwinSalas/go-grpc-auth-svc/pkg/utils"
	authpb "github.com/ErwinSalas/go-grpc-auth-svc/proto"
	"github.com/ErwinSalas/go-grpc-tls/pkg/gogrpctls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/reflection"

	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Printf("Received request: %v", req)
	log.Printf("Method: %s", info.FullMethod)

	resp, err := handler(ctx, req)
	return resp, err
}

func main() {
	c, err := config.LoadConfig()

	c.CertM = gogrpctls.NewCertManager()
	if err != nil {
		log.Fatalln("Failed at config", err)
	}

	h := database.Init(c.DBUrl)

	jwt := utils.JwtWrapper{
		SecretKey:       c.JWTSecretKey,
		Issuer:          "go-grpc-auth-svc",
		ExpirationHours: 24 * 365,
	}

	listen, err := net.Listen("tcp", c.Port)

	if err != nil {
		log.Fatalln("Failed to listing:", err)
	}

	fmt.Println("Auth Svc on", c.Port)

	tlsCredentials, err := c.CertM.LoadServerCertificate()
	if err != nil {
		log.Fatal("cannot load TLS credentials: ", err)
	}

	grpcServer := grpc.NewServer(grpc.Creds(tlsCredentials), grpc.UnaryInterceptor(loggingInterceptor))
	authService := auth.NewAuthService(auth.NewUserRepository(h), jwt) // Puedes pasar una conexión de base de datos real aquí.
	authpb.RegisterAuthServiceServer(grpcServer, server.NewAuthServer(authService))

	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthcheck)

	// Start health check routine
	go func() {
		for {
			var count int64
			if err := h.DB.Table("users").Count(&count).Error; err != nil {
				log.Println("Database query error:", err)
				healthcheck.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
				return
			} else {
				healthcheck.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
				time.Sleep(5 * time.Second)
			}
		}
	}()
	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
