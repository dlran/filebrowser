package worker

import (
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"log"
)

type Worker struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

func NewWorker(redisURL string) (*Worker, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", err)
	}

	redisOpt := asynq.RedisClientOpt{
		Addr:      options.Addr,
		Password:  options.Password,
		DB:        options.DB,
		Username:  options.Username,
		TLSConfig: options.TLSConfig,
	}
	mux := asynq.NewServeMux()

	return &Worker{
		server: asynq.NewServer(redisOpt, asynq.Config{
			Concurrency: 3,
		}),
		mux: mux,
	}, nil
}

func (w *Worker) Start() error {
	w.registerHandlers()

	log.Println("Asynq worker started")
	return w.server.Run(w.mux)
}

func (w *Worker) Shutdown() {
	if w.server != nil {
		w.server.Shutdown()
	} else {
		log.Println("Asynq server not assign")
	}
}

func (w *Worker) registerHandlers() {
	w.mux.HandleFunc("internal:init_queue", initHandler)
	w.mux.HandleFunc("file:encrypt", encryptFileHandler)
	w.mux.HandleFunc("file:exec_scripts", executeScriptTaskHandler)
}
