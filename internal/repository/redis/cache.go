package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"task-service/internal/domain/models"
	"task-service/internal/metrics"
)

const redisCacheName = "redis"

type Cache struct {
	client *goredis.Client
}

func NewCache(client *goredis.Client) *Cache {
	return &Cache{client: client}
}

func (c *Cache) GetTaskList(ctx context.Context, key string) (models.TaskList, bool, error) {
	raw, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, goredis.Nil) {
		metrics.RecordCacheLookup(redisCacheName, "get_task_list", false, nil)
		return models.TaskList{}, false, nil
	}
	if err != nil {
		metrics.RecordCacheLookup(redisCacheName, "get_task_list", false, err)
		return models.TaskList{}, false, err
	}
	var dto taskListDTO
	if err := json.Unmarshal(raw, &dto); err != nil {
		metrics.RecordCacheLookup(redisCacheName, "get_task_list", false, err)
		return models.TaskList{}, false, err
	}
	metrics.RecordCacheLookup(redisCacheName, "get_task_list", true, nil)
	return dto.toDomain(), true, nil
}

func (c *Cache) SetTaskList(ctx context.Context, key string, value models.TaskList, ttl time.Duration) error {
	raw, err := json.Marshal(newTaskListDTO(value))
	if err != nil {
		metrics.RecordCacheOperation(redisCacheName, "set_task_list", err)
		return err
	}
	err = c.client.Set(ctx, key, raw, ttl).Err()
	metrics.RecordCacheOperation(redisCacheName, "set_task_list", err)
	return err
}

func (c *Cache) DeleteTeamTasks(ctx context.Context, teamID int64) error {
	pattern := fmt.Sprintf("team_tasks:%d:*", teamID)
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			metrics.RecordCacheOperation(redisCacheName, "delete_team_tasks", err)
			return err
		}
	}
	err := iter.Err()
	metrics.RecordCacheOperation(redisCacheName, "delete_team_tasks", err)
	return err
}
