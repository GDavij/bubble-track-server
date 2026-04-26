package queue

import (
	"github.com/hibiken/asynq"
)

func NewServeMux() *asynq.ServeMux {
	return asynq.NewServeMux()
}
