package main

import (
	"net/http"
	"time"

	"example.com/redis_rate_limiter"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type album struct {
	ID string `json:"id"`
	Title string `json:"title"`
	Artist string `json:"artist"`
	Price float64 `json:"price"`
}

var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan", Artist: "Sarah Vaughan", Price: 39.99},
}

func main() {

	router := gin.Default()

	// First setup a custom request handler wrapper for rate limiting 

	// Redis connection for storing rate limit data
	rdb := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "", // no password set
        DB:       0,  // use default DB
    })

	// A Counter Based (Fixed window) strategy
	var ratedStrategy redis_rate_limiter.Strategy = redis_rate_limiter.NewCounterStrategy(rdb, time.Now)
	
	// Which field from header to rate limit based on
	var ratedExtractor redis_rate_limiter.Extractor = redis_rate_limiter.NewHTTPHeadersExtractor("userId")
	
	// Rate limit configuratio
	var ratedConfig *redis_rate_limiter.RateLimiterConfig = &redis_rate_limiter.RateLimiterConfig{
		Extractor: ratedExtractor,
		Strategy: ratedStrategy,
		// Says 1 request per hour
		Expiration: time.Hour,
		MaxRequests: 1,

	}

	// The custom rate limit handler
	var ratedHandler http.Handler = redis_rate_limiter.NewHTTPRateLimiterHandler(router, ratedConfig)
	
	// These routes are all rate limited
	router.GET("/albums", getAlbums)
	router.POST("/albums", postAlbums)
	router.GET("/albums/:id", getAlbumByID)

	// Start a custom http server with the created http handler
	s := &http.Server{
		Addr:           ":8080",
		Handler:        ratedHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()
}

func getAlbums(c *gin.Context){
	c.IndentedJSON(http.StatusOK, albums)
}

func postAlbums(c *gin.Context){
	var newAlbum album
	
	if err := c.BindJSON(&newAlbum); err != nil {
		return
	}
	
	albums = append(albums, newAlbum)
	c.IndentedJSON(http.StatusCreated, newAlbum)
}

func getAlbumByID(c *gin.Context){
	id := c.Param("id")
	for _, a := range albums {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
}
