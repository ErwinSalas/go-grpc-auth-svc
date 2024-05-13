package main

import (
	"fmt"
	"log"
	"net"

	"github.com/ErwinSalas/go-grpc-auth-svc/pkg/auth"
	"github.com/ErwinSalas/go-grpc-auth-svc/pkg/config"
	"github.com/ErwinSalas/go-grpc-auth-svc/pkg/database"
	"github.com/ErwinSalas/go-grpc-auth-svc/pkg/server"
	"github.com/ErwinSalas/go-grpc-auth-svc/pkg/utils"
	authpb "github.com/ErwinSalas/go-grpc-auth-svc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	c, err := config.LoadConfig()

	// c.CertM = gogrpctls.NewCertManager()
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

	//tlsCredentials, err := c.CertM.LoadServerCertificate()
	// if err != nil {
	// 	log.Fatal("cannot load TLS credentials: ", err)
	// }

	noCreds := insecure.NewCredentials()

	grpcServer := grpc.NewServer(grpc.Creds(noCreds))
	authService := auth.NewAuthService(auth.NewUserRepository(h), jwt) // Puedes pasar una conexión de base de datos real aquí.
	authpb.RegisterAuthServiceServer(grpcServer, server.NewAuthServer(authService))

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
