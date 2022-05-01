# go_api_rate_limiter
A simple API rate limiter in GO

### Redis Rate Limiter

The redis_rate_limiter has a Fixed Window Based implementation of a rate limiter using redis as storage. A Http handler has also been created which can be plugged in as a middleware to any http based application.

web_service_gin demonstrates how to use this middleware.
