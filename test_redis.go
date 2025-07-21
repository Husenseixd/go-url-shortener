package main

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {
	opt, _ := redis.ParseURL("rediss://default:AXRBAAIjcDE2MWI3ZDEwZTFlMGI0NWVjODhjMTAwYWI5NzE1YmVhMHAxMA@sound-duck-29761.upstash.io:6379")
	client := redis.NewClient(opt)

	client.Set(ctx, "foo", "bar", 0)
	val := client.Get(ctx, "foo").Val()
	print(val)
}
