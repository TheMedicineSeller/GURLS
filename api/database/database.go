package database

import (
	"github.com/TheMedicineSeller/GURLS/config"
	"github.com/go-redis/redis/v8"
	"context"
)

var Ctx = context.Background()

func CreateClient (dbnum int) *redis.Client {
    client := redis.NewClient(&redis.Options{
        Addr: config.DB_ADDR,
        Password: config.DB_PASSWORD,
        DB: dbnum,
    })
    return client
}
