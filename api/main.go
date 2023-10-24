package main

import (
    "net/http"
    "log"
    "github.com/gin-gonic/gin"
    "github.com/TheMedicineSeller/GURLS/db"
)

// Define endpoints for app and attach corresponding functions to the endpoints with the right method

func main () {
	database, err := db.CreateNewDB(RedisAddr)
	if err != nil {
	    log.Fatalf("Redis connection failed : %s", err.Error())
	}
}

