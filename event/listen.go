package event

import (
	"fmt"
	"github.com/TicketsBot/common/eventforwarding"
	"github.com/TicketsBot/worker"
	"github.com/go-redis/redis"
	"github.com/rxdn/gdl/cache"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"github.com/rxdn/gdl/rest/ratelimit"
)

func Listen(redis *redis.Client, cache *cache.PgCache) {
	ch := make(chan eventforwarding.Event)
	go eventforwarding.Listen(redis, ch)

	for event := range ch {
		var keyPrefix string

		if event.IsWhitelabel {
			keyPrefix = fmt.Sprintf("ratelimiter:%d", event.BotId)
		} else {
			keyPrefix = "ratelimiter:public"
		}

		ctx := &worker.Context{
			Token:        event.BotToken,
			BotId:        event.BotId,
			IsWhitelabel: event.IsWhitelabel,
			ShardId:      event.ShardId,
			Cache:        cache,
			RateLimiter:  ratelimit.NewRateLimiter(ratelimit.NewRedisStore(redis, keyPrefix), 1),
		}

		execute(ctx, events.EventType(event.EventType), event.Data, event.Extra)
	}
}
