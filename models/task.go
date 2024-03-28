package model

import (
	"github.com/Jaylenwa/Vfoy/pkg/util"
	"github.com/jinzhu/gorm"
)

// Task 任务模型
type Task struct {
	gorm.Model
	Status   int    // 任务状态
	Type     int    // 任务类型
	UserID   uint   // 发起者UID，0表示为系统发起
	Progress int    // 进度
	Error    string `gorm:"type:text"` // 错误信息
	Props    string `gorm:"type:text"` // 任务属性
}

// Create 创建任务记录
func (task *Task) Create() (uint, error) {
	if err := DB.Create(task).Error; err != nil {
		util.Log().Warning("Failed to insert task record: %s", err)
		return 0, err
	}
	return task.ID, nil
}

// SetStatus 设定任务状态
func (task *Task) SetStatus(status int) error {
	return DB.Model(task).Select("status").Updates(map[string]interface{}{"status": status}).Error
}

// SetProgress 设定任务进度
func (task *Task) SetProgress(progress int) error {
	return DB.Model(task).Select("progress").Updates(map[string]interface{}{"progress": progress}).Error
}

// SetError 设定错误信息
func (task *Task) SetError(err string) error {
	return DB.Model(task).Select("error").Updates(map[string]interface{}{"error": err}).Error
}

// GetTasksByStatus 根据状态检索任务
func GetTasksByStatus(status ...int) []Task {
	var tasks []Task
	DB.Where("status in (?)", status).Find(&tasks)
	return tasks
}

// GetTasksByID 根据ID检索任务
func GetTasksByID(id interface{}) (*Task, error) {
	task := &Task{}
	result := DB.Where("id = ?", id).First(task)
	return task, result.Error
}

// ListTasks 列出用户所属的任务
func ListTasks(uid uint, page, pageSize int, order string) ([]Task, int) {
	var (
		tasks []Task
		total int
	)
	dbChain := DB
	dbChain = dbChain.Where("user_id = ?", uid)

	// 计算总数用于分页
	dbChain.Model(&Task{}).Count(&total)

	// 查询记录
	dbChain.Limit(pageSize).Offset((page - 1) * pageSize).Order(order).Find(&tasks)

	return tasks, total
}
