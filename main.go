package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/joho/godotenv"
	http_api "github.com/tolopsy/url-shortener/api"
	m_repo "github.com/tolopsy/url-shortener/repository/mongodb"
	r_repo "github.com/tolopsy/url-shortener/repository/redis"
	"github.com/tolopsy/url-shortener/shortener"
)

func main() {
	godotenv.Load() // load .env file if exists

	repo := chooseRepo()
	service := shortener.NewRedirectService(repo)
	handler := http_api.NewHandler(service)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get("/{code}", handler.Get)
	router.Post("/", handler.Post)

	errChan := make(chan error, 2)
	defer close(errChan)

	go func(){
		port := httpPort()
		fmt.Println("Listening on port ", port)
		errChan <- http.ListenAndServe(httpPort(), router)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	fmt.Printf("Terminated %s\n", <- errChan)
}

func httpPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	return fmt.Sprintf(":%s", port)
}

func chooseRepo() shortener.RedirectRepository {
	var repo shortener.RedirectRepository
	switch os.Getenv("URL_DB") {
	case "redis":
		redisURL := os.Getenv("REDIS_URL")
		redisRepo, err := r_repo.NewRedisRepository(redisURL)
		if err != nil {
			log.Fatal(err)
		}
		repo = redisRepo

	case "mongo":
		mongoURL := os.Getenv("MONGO_URL")
		mongoDB := os.Getenv("MONGO_DB")
		mongoCollection := os.Getenv("MONGO_REDIRECT_COLLECTION")
		mongoTimeout, _ := strconv.Atoi(os.Getenv("MONGO_TIMEOUT"))
		mongoRepo, err := m_repo.NewMongoRepository(mongoURL, mongoDB, mongoCollection, mongoTimeout)
		if err != nil {
			log.Fatal(err)
		}
		repo = mongoRepo
	}
	return repo
}
