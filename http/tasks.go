package fbhttp

import (
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"log"
	"net/http"
	"time"
)

type BaseParams struct {
	Src  []string `json:"src"`
	Root string   `json:"root"`
}

type ExtendParams struct {
	BaseParams
	Password string `json:"password"`
	Mode     bool   `json:"mode"`
}

var taskCallHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	asynqHandler := getAsynqHandler()

	taskName := r.URL.Query().Get("taskName")
	payload := &ExtendParams{}
	err := json.NewDecoder(r.Body).Decode(payload)
	if err != nil {
		return http.StatusBadRequest, err
	}
	payload.Root = d.server.Root

	payloadJson, _ := json.Marshal(payload)
	task := asynq.NewTask(taskName, payloadJson)
	info, err := asynqHandler.client.Enqueue(task,
		asynq.Retention(4*time.Hour),
		asynq.MaxRetry(0),
	)
	if err != nil {
		return errToStatus(err), err
	}
	res := fmt.Sprintf("Start task id: %s", info.ID)
	log.Println(res)

	return 0, nil
	//renderJSON(w, r, res)
})

var taskListHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	asynqHandler := getAsynqHandler()
	taskRes, err := asynqHandler.GetDefaultQueueTasks()
	if err != nil {
		return errToStatus(err), err
	}
	return renderJSON(w, r, taskRes)
})

var taskDeleteHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	queue := "default"
	taskID := r.URL.Query().Get("id")
	state := r.URL.Query().Get("state")
	if taskID == "" {
		err := fmt.Errorf("Invalid request parameters")
		return errToStatus(err), err
	}

	asynqHandler := getAsynqHandler()
    var err error
    if state == "active" {
	    err = asynqHandler.inspector.CancelProcessing(taskID)
    } else {
	    err = asynqHandler.inspector.DeleteTask(queue, taskID)
    }
	if err != nil {
		return errToStatus(err), err
	}
	return 0, nil
})
