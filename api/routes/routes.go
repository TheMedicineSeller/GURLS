package routes

import (
	"net/http"
	"strconv"
	"time"

	"github.com/TheMedicineSeller/GURLS/config"
	"github.com/TheMedicineSeller/GURLS/database"
	"github.com/TheMedicineSeller/GURLS/utils"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type request struct {
	URL        string        `json:"url"`
	ExpiryTime time.Duration `json:"expiry_time"`
}

type response struct {
	URL          string        `json:"url"`
	ShortenedURL string        `json:"shortened_url"`
	ExpiryTime   time.Duration `json:"expiry_time"`
	// Shortenings left for the quota until reset
	RateRemaining int `json:"rate_limit"`
	// Time left for resetting usage quota
	RateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(c *gin.Context) {
	// Getting the JSON request and filling up jsonbody
	jsonbody := new(request)
	if err := c.ShouldBindJSON(&jsonbody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Implement Rate limiting for Controlling user requests for shortening service
	rdb2 := database.CreateClient(1) // DB for monitoring IP and corresponding usage
	defer rdb2.Close()

	usage_left, err := rdb2.Get(database.Ctx, c.ClientIP()).Result()
	if err == redis.Nil {
		// Smart way to implement Resetting of API_LIMIT. We essentially keep the reset time (time after which API usage amount is refilled) as the IP key's expiry time in rdb2.
		_ = rdb2.Set(database.Ctx, c.ClientIP(), config.API_LIMIT, 30*60*time.Second).Err()
	} else {
		usage_left_int, _ := strconv.Atoi(usage_left)
		if usage_left_int <= 0 {
			time_left, _ := rdb2.TTL(database.Ctx, c.ClientIP()).Result()
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":            "Service usage limit exceeded !",
				"rate_limit_reset": time_left / (time.Nanosecond * time.Minute),
			})
			return
		}
	}
	rdb2.Decr(database.Ctx, c.ClientIP())

	// We will perform a few necessary checks before we do the actual shortening
	// These include checking if the URL is valid, checking for Domain errors and enforcing HTTPS.
	if !govalidator.IsURL(jsonbody.URL) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid URL",
		})
		return
	}
	if !utils.RemoveDomainError(jsonbody.URL) {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Domain wrong",
		})
		return
	}

	jsonbody.URL = utils.EnforceHTTP(jsonbody.URL)

	// Shortening service (random gen lmao)
	id := uuid.New().String()[:6]

	rdb := database.CreateClient(0)
	defer rdb.Close()

	// Most naive rehashing : keep generating random string till it doesnt exist in rdb
	val, _ := rdb.Get(database.Ctx, id).Result()
	for val != "" {
		id = uuid.New().String()[:6]
		val, _ = rdb.Get(database.Ctx, id).Result()
	}

	if jsonbody.ExpiryTime == 0 {
		jsonbody.ExpiryTime = 24 // Expiry time of shortened urls
	}
	err = rdb.Set(database.Ctx, id, jsonbody.URL, jsonbody.ExpiryTime*3600*time.Second).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to connect to server",
		})
	}

	resp := response{
		URL:            jsonbody.URL,
		ShortenedURL:   "",
		ExpiryTime:     jsonbody.ExpiryTime,
		RateRemaining:  config.API_LIMIT,
		RateLimitReset: 30,
	}
	time_left, _ := rdb2.TTL(database.Ctx, c.ClientIP()).Result()

	resp.RateRemaining, _ = strconv.Atoi(usage_left)
	resp.RateLimitReset = time_left / (time.Nanosecond * time.Minute)
	resp.ShortenedURL = config.DOMAIN + "/" + id

	c.JSON(http.StatusOK, resp)
}

func ResolveURL(c *gin.Context) {
	url := c.Param("url")
	rdb := database.CreateClient(0)

	defer rdb.Close()

	actual_url, err := rdb.Get(database.Ctx, url).Result()
	// Get url value in database and check if its non empty and if theres no error
	if err == redis.Nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "short URL not found in db",
		})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Cannot connect to db",
		})
		return
	}

	rInr := database.CreateClient(1)
	defer rInr.Close()
	_ = rInr.Incr(database.Ctx, "counter")

	// If everything goes fine then redirect user to the actual URL
	c.Redirect(301, actual_url)
}
