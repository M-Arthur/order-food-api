package bootstrap

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // or your driver

	"github.com/M-Arthur/order-food-api/internal/config"
	"github.com/M-Arthur/order-food-api/internal/domain"
	"github.com/M-Arthur/order-food-api/internal/httpapi/handlers"
	"github.com/M-Arthur/order-food-api/internal/service"
	"github.com/M-Arthur/order-food-api/internal/storage"
)

type Infra struct {
	DB *sql.DB
}

type Repos struct {
	Product domain.ProductRepository
	Order   domain.OrderRepository
}

type Services struct {
	Product service.ProductService
	Order   service.OrderService
}

type Handlers struct {
	Product *handlers.ProductHandler
	Order   *handlers.OrderHandler
}

type Dependencies struct {
	Infra    Infra
	Repos    Repos
	Services Services
	Handlers Handlers
}

func BuildDependencies(c config.Config) (*Dependencies, error) {
	infraPtr, err := buildInfra(c)
	if err != nil {
		return nil, err
	}

	infra := *infraPtr
	repos := buildRepos(infra)
	services := buildServices(repos)
	handlers := buildHandlers(services)

	return &Dependencies{
		Infra:    infra,
		Repos:    repos,
		Services: services,
		Handlers: handlers,
	}, nil
}

func buildInfra(c config.Config) (*Infra, error) {
	// DB connection
	dsn := c.DB.DSN
	if dsn == "" {
		return nil, fmt.Errorf("DB_DSN env var is required")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(c.DB.MaxOpenConns)
	db.SetMaxIdleConns(c.DB.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(c.DB.MaxLifetime))

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return &Infra{DB: db}, nil
}

func buildRepos(inf Infra) Repos {
	// Product repo
	seedProducts := []domain.Product{
		{ID: domain.ProductID("10"), Name: "Chicken Waffle", Price: domain.NewMoneyFromFloat(12.5), Category: "Waffle"},
		{ID: domain.ProductID("11"), Name: "Fries", Price: domain.NewMoneyFromFloat(5.5), Category: "Sides"},
	}
	pr := storage.NewInMemoryProductRepository(seedProducts)

	// Order repo
	or := storage.NewPgOrderRepository(inf.DB)

	return Repos{
		Product: pr,
		Order:   or,
	}
}

func buildServices(r Repos) Services {
	ps := service.NewProductService(r.Product)
	os := service.NewOrderService(r.Order, r.Product)

	return Services{
		Product: ps,
		Order:   os,
	}
}

func buildHandlers(svc Services) Handlers {
	ph := handlers.NewProductHandler(svc.Product)
	oh := handlers.NewOrderHandler(svc.Order, svc.Product)

	return Handlers{
		Product: ph,
		Order:   oh,
	}
}
