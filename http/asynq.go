package fbhttp

import (
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"log"
	"sort"
	"sync"
	"time"
)

type AsynqHandler struct {
	inspector *asynq.Inspector
	client    *asynq.Client
}

var (
	manager *AsynqHandler
	once    sync.Once
)

func NewAsynqInspector(redisURL string) error {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("invalid redis URL: %w", err)
	}
	redisOpt := asynq.RedisClientOpt{
		Addr:      options.Addr,
		Password:  options.Password,
		DB:        options.DB,
		Username:  options.Username,
		TLSConfig: options.TLSConfig,
	}
	once.Do(func() {
		manager = &AsynqHandler{
			inspector: asynq.NewInspector(redisOpt),
			client:    asynq.NewClient(redisOpt),
		}
		q, _ := manager.ListQueues()
		if len(q) == 0 {
			task := asynq.NewTask("internal:init_queue", nil)
			info, err := manager.client.Enqueue(task)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf(
				"Enqueued initialization task: %s. This ensures the default queue exists in Redis so the Asynq inspector can detect it. Without this, inspector.Queues() would return empty and querying tasks would fail.",
				info.ID,
			)
		}
	})
	return nil
}

func getAsynqHandler() *AsynqHandler {
	if manager == nil {
		log.Fatalf("InspectorManager not initialized")
	}
	return manager
}

func (h *AsynqHandler) ListQueues() ([]string, error) {
	queues, err := h.inspector.Queues()
	if err != nil {
		log.Fatalf("Failed to get queues")
		return []string{}, err
	}

	return queues, nil
}

type TaskResponse struct {
	Tasks  []*TaskView `json:"tasks"`
	Counts TaskCounts  `json:"counts"`
}

type TaskCounts struct {
	Active    int `json:"active"`
	Pending   int `json:"pending"`
	Scheduled int `json:"scheduled"`
	Retry     int `json:"retry"`
	Archived  int `json:"archived"`
	Completed int `json:"completed"`
}

func (h *AsynqHandler) GetDefaultQueueTasks() (TaskResponse, error) {
	queue := "default"
	state := "all"

	var tasks []*asynq.TaskInfo
	var err error

	if state != "all" {
		tasks, err = h.queryTasksByState(queue, state)
	} else {
		tasks, err = h.queryAllTasks(queue)
	}

	if err != nil {
		log.Fatalf("failed to get tasks %s\n", err)
		return TaskResponse{}, err
	}

	// 转换为自定义视图
	taskViews := ConvertToTaskViews(tasks)

	info, err := h.inspector.GetQueueInfo(queue)
	if err != nil {
		log.Fatalf("Failed to get queue stats")
		return TaskResponse{}, err
	}
	counts := TaskCounts{
		Active:    info.Active,
		Pending:   info.Pending,
		Scheduled: info.Scheduled,
		Retry:     info.Retry,
		Archived:  info.Archived,
		Completed: info.Completed,
	}
	taskResponse := TaskResponse{
		Tasks:  taskViews,
		Counts: counts,
	}

	return taskResponse, nil
}

// 查询特定状态的任务
func (h *AsynqHandler) queryTasksByState(queue, state string) ([]*asynq.TaskInfo, error) {
	switch state {
	case "pending":
		return h.inspector.ListPendingTasks(queue)
	case "active":
		return h.inspector.ListActiveTasks(queue)
	case "scheduled":
		return h.inspector.ListScheduledTasks(queue)
	case "retry":
		return h.inspector.ListRetryTasks(queue)
	case "archived":
		return h.inspector.ListArchivedTasks(queue)
	case "completed":
		return h.inspector.ListCompletedTasks(queue, asynq.PageSize(50), asynq.Page(0))
	default:
		return nil, fmt.Errorf("invalid state parameter: %s", state)
	}
}

// 查询所有状态的任务
func (h *AsynqHandler) queryAllTasks(queue string) ([]*asynq.TaskInfo, error) {
	var allTasks []*asynq.TaskInfo

	// 定义所有可能的状态
	states := []string{"pending", "active", "scheduled", "retry", "completed", "archived"}

	for _, state := range states {
		tasks, err := h.queryTasksByState(queue, state)
		if err != nil {
			return nil, err
		}
		sortTask(tasks)
		allTasks = append(allTasks, tasks...)
	}

	return allTasks, nil
}
func sortTask(tasks []*asynq.TaskInfo) {
	sort.Slice(tasks, func(i, j int) bool {
		// 获取任务 i 的时间
		timeI := getTaskTime(tasks[i])
		// 获取任务 j 的时间
		timeJ := getTaskTime(tasks[j])

		// 如果任务 i 没有时间，则放在后面
		if timeI.IsZero() {
			return false
		}
		// 如果任务 j 没有时间，则 i 在前面
		if timeJ.IsZero() {
			return true
		}
		// 正常情况下，按时间排序（升序）
		return timeI.After(timeJ)
	})
}

func getTaskTime(task *asynq.TaskInfo) time.Time {
	if !task.CompletedAt.IsZero() {
		return task.CompletedAt
	}
	if !task.LastFailedAt.IsZero() {
		return task.LastFailedAt
	}
	return time.Time{} // 返回零值
}

// {
//   "ID": "9b8f7063-ee7a-4785-8c31-78c447814f4b",
//   "Queue": "default",
//   "Type": "test_task",
//   "Payload": null,
//   "State": 6,
//   "MaxRetry": 3,
//   "Retried": 0,
//   "LastErr": "",
//   "LastFailedAt": "0001-01-01T00:00:00Z",
//   "Timeout": 1800000000000,
//   "Deadline": "0001-01-01T00:00:00Z",
//   "Group": "",
//   "NextProcessAt": "0001-01-01T00:00:00Z",
//   "IsOrphaned": false,
//   "Retention": 3600000000000,
//   "CompletedAt": "2025-03-30T14:28:18+08:00",
//   "Result": null
// }

type TaskView struct {
	ID      string `json:"id"`      // 任务ID
	Queue   string `json:"queue"`   // 队列名称
	Type    string `json:"type"`    // 任务类型
	Payload string `json:"payload"` // 任务负载(JSON字符串)
	State   string `json:"state"`   // 任务状态
	// Priority     int       `json:"priority"`      // 优先级
	Retried       int       `json:"retried"`         // 已重试次数
	MaxRetry      int       `json:"max_retry"`       // 最大重试次数
	LastFailedAt  string    `json:"last_failed_at"`  // 最后失败时间
	LastErr       string    `json:"last_error"`      // 最后错误信息
	NextProcessAt time.Time `json:"next_process_at"` // 下次处理时间
	CompletedAt   string    `json:"completed_at"`    // 完成时间
	// CreatedAt    time.Time `json:"created_at"`    // 创建时间
	Timeout  int       `json:"timeout"` // 超时时间(秒)
	Result   string    `json:"result"`  // 任务结果
	Deadline time.Time `json:"deadline"`
}

func ConvertToTaskViews(tasks []*asynq.TaskInfo) []*TaskView {
	views := make([]*TaskView, 0, len(tasks))

	for _, task := range tasks {
		views = append(views, &TaskView{
			ID:            task.ID,
			Queue:         task.Queue,
			Type:          task.Type,
			Payload:       string(task.Payload),
			State:         task.State.String(),
			Retried:       task.Retried,
			MaxRetry:      task.MaxRetry,
			LastFailedAt:  FormatTaskTime(task.LastFailedAt),
			LastErr:       task.LastErr,
			NextProcessAt: task.NextProcessAt,
			CompletedAt:   FormatTaskTime(task.CompletedAt),
			Timeout:       int(task.Timeout / time.Second),
			Result:        string(task.Result),
			Deadline:      task.Deadline,
		})
	}

	return views
}

func FormatTaskTime(t time.Time) string {
	layout := "2006-01-02 15:04:05"

	return t.Format(layout)
}

func (h *AsynqHandler) DeleteTask(taskID string) error {
	queue := "default"
	if taskID == "" {
		log.Println("Invalid request params")
		return fmt.Errorf("Invalid taskID params: %s", taskID)
	}
	err := h.inspector.DeleteTask(queue, taskID)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
