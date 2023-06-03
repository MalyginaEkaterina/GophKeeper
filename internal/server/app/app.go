package app

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	pb "github.com/MalyginaEkaterina/GophKeeper/internal/common/proto"
	"github.com/MalyginaEkaterina/GophKeeper/internal/server"
	"github.com/MalyginaEkaterina/GophKeeper/internal/server/handlers"
	"github.com/MalyginaEkaterina/GophKeeper/internal/server/service"
	"github.com/MalyginaEkaterina/GophKeeper/internal/server/storage"
	"github.com/caarlos0/env/v6"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"math/big"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Start() {
	var cfg server.Config
	flag.StringVar(&cfg.Address, "a", "localhost:3200", "address to listen on")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "database connection string")
	flag.StringVar(&cfg.FileUserStoragePath, "fu", "", "file data storage path")
	flag.StringVar(&cfg.FileDataStoragePath, "fd", "", "file user storage path")
	flag.StringVar(&cfg.SecretFilePath, "s", "", "path to file with secret")
	flag.Parse()
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal("Error while parsing env: ", err)
	}
	db, userStore, dataStore, err := initStore(cfg)
	if err != nil {
		log.Fatal("Init store error: ", err)
	}
	defer func() {
		if db != nil {
			db.Close()
		}
	}()
	defer userStore.Close()
	secretKey, err := getSecret(cfg.SecretFilePath)
	if err != nil {
		log.Fatal("Error while reading secret key", err)
	}
	authService := &service.AuthServiceImpl{Store: userStore, SecretKey: secretKey}
	userServer := handlers.NewUserServer(authService)

	listen, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Fatal(err)
	}

	cert, err := generateTLSCertificate()
	if err != nil {
		log.Fatal("Error while generating TLS certificate", err)
	}
	conf := &tls.Config{
		Certificates: []tls.Certificate{*cert},
	}
	tlsCredentials := credentials.NewTLS(conf)
	serverOpts := []grpc.ServerOption{grpc.UnaryInterceptor(userServer.AuthInterceptor), grpc.Creds(tlsCredentials)}
	s := grpc.NewServer(serverOpts...)

	sigint := make(chan os.Signal, 1)
	connsClosed := make(chan struct{})
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		<-sigint
		s.GracefulStop()
		close(connsClosed)
	}()

	pb.RegisterUserServer(s, userServer)
	pb.RegisterKeeperServer(s, handlers.NewKeeperServer(dataStore))
	log.Printf("Started server on %s\n", cfg.Address)
	if err := s.Serve(listen); err != nil {
		log.Fatal(err)
	}
	<-connsClosed
	log.Printf("Stopped server on %s\n", cfg.Address)
}

func initStore(cfg server.Config) (*sql.DB, storage.UserStorage, storage.DataStorage, error) {
	var usersStore storage.UserStorage
	var dataStore storage.DataStorage
	var db *sql.DB
	var err error
	if cfg.DatabaseURI != "" {
		db, err = sql.Open("postgres", cfg.DatabaseURI)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("database connection error: %w", err)
		}
		err = storage.DoMigrations(db)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("running migrations error: %w", err)
		}
		log.Printf("Using database storage %s\n", cfg.DatabaseURI)
		usersStore, err = storage.NewDBUserStorage(db)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("create user store error: %w", err)
		}
		dataStore, err = storage.NewDBDataStorage(db)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("create data store error: %w", err)
		}
	} else {
		log.Printf("Using cached storage")
		usersStore = storage.NewCachedFileUserStorage()
		dataStore = storage.NewCachedFileDataStorage()
	}
	return db, usersStore, dataStore, nil
}

func generateTLSCertificate() (*tls.Certificate, error) {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"Yandex.Praktikum"},
			Country:      []string{"RU"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:     []string{"localhost"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf(`error while key generating: %w`, err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf(`error while creating certificate: %w`, err)
	}

	return &tls.Certificate{
		Certificate: [][]byte{certBytes},
		PrivateKey:  privateKey,
		Leaf:        cert,
	}, nil
}

func getSecret(path string) ([]byte, error) {
	if path == "" {
		return nil, errors.New("secret path must be configured")
	}
	return os.ReadFile(path)
}
