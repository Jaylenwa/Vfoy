package task

import (
	"context"
	"encoding/json"
	"path"

	model "github.com/Jaylenwa/Vfoy/models"
	"github.com/Jaylenwa/Vfoy/pkg/filesystem"
	"github.com/Jaylenwa/Vfoy/pkg/filesystem/fsctx"
	"github.com/Jaylenwa/Vfoy/pkg/util"
)

// ImportTask 导入务
type ImportTask struct {
	User      *model.User
	TaskModel *model.Task
	TaskProps ImportProps
	Err       *JobError
}

// ImportProps 导入任务属性
type ImportProps struct {
	PolicyID  uint   `json:"policy_id"`    // 存储策略ID
	Src       string `json:"src"`          // 原始路径
	Recursive bool   `json:"is_recursive"` // 是否递归导入
	Dst       string `json:"dst"`          // 目的目录
}

// Props 获取任务属性
func (job *ImportTask) Props() string {
	res, _ := json.Marshal(job.TaskProps)
	return string(res)
}

// Type 获取任务状态
func (job *ImportTask) Type() int {
	return ImportTaskType
}

// Creator 获取创建者ID
func (job *ImportTask) Creator() uint {
	return job.User.ID
}

// Model 获取任务的数据库模型
func (job *ImportTask) Model() *model.Task {
	return job.TaskModel
}

// SetStatus 设定状态
func (job *ImportTask) SetStatus(status int) {
	job.TaskModel.SetStatus(status)
}

// SetError 设定任务失败信息
func (job *ImportTask) SetError(err *JobError) {
	job.Err = err
	res, _ := json.Marshal(job.Err)
	job.TaskModel.SetError(string(res))
}

// SetErrorMsg 设定任务失败信息
func (job *ImportTask) SetErrorMsg(msg string, err error) {
	jobErr := &JobError{Msg: msg}
	if err != nil {
		jobErr.Error = err.Error()
	}
	job.SetError(jobErr)
}

// GetError 返回任务失败信息
func (job *ImportTask) GetError() *JobError {
	return job.Err
}

// Do 开始执行任务
func (job *ImportTask) Do() {
	ctx := context.Background()

	// 查找存储策略
	policy, err := model.GetPolicyByID(job.TaskProps.PolicyID)
	if err != nil {
		job.SetErrorMsg("Policy not exist.", err)
		return
	}

	// 创建文件系统
	job.User.Policy = policy
	fs, err := filesystem.NewFileSystem(job.User)
	if err != nil {
		job.SetErrorMsg(err.Error(), nil)
		return
	}
	defer fs.Recycle()

	fs.Policy = &policy
	if err := fs.DispatchHandler(); err != nil {
		job.SetErrorMsg("Failed to dispatch policy.", err)
		return
	}

	// 注册钩子
	fs.Use("BeforeAddFile", filesystem.HookValidateFile)
	fs.Use("BeforeAddFile", filesystem.HookValidateCapacity)

	// 列取目录、对象
	job.TaskModel.SetProgress(ListingProgress)
	coxIgnoreConflict := context.WithValue(context.Background(), fsctx.IgnoreDirectoryConflictCtx,
		true)
	objects, err := fs.Handler.List(ctx, job.TaskProps.Src, job.TaskProps.Recursive)
	if err != nil {
		job.SetErrorMsg("Failed to list files.", err)
		return
	}

	job.TaskModel.SetProgress(InsertingProgress)

	// 虚拟目录路径与folder对象ID的对应
	pathCache := make(map[string]*model.Folder, len(objects))

	// 插入目录记录到用户文件系统
	for _, object := range objects {
		if object.IsDir {
			// 创建目录
			virtualPath := path.Join(job.TaskProps.Dst, object.RelativePath)
			folder, err := fs.CreateDirectory(coxIgnoreConflict, virtualPath)
			if err != nil {
				util.Log().Warning("Importing task cannot create user directory %q: %s", virtualPath, err)
			} else if folder.ID > 0 {
				pathCache[virtualPath] = folder
			}
		}
	}

	// 插入文件记录到用户文件系统
	for _, object := range objects {
		if !object.IsDir {
			// 创建文件信息
			virtualPath := path.Dir(path.Join(job.TaskProps.Dst, object.RelativePath))
			fileHeader := fsctx.FileStream{
				Size:        object.Size,
				VirtualPath: virtualPath,
				Name:        object.Name,
				SavePath:    object.Source,
			}

			// 查找父目录
			parentFolder := &model.Folder{}
			if parent, ok := pathCache[virtualPath]; ok {
				parentFolder = parent
			} else {
				folder, err := fs.CreateDirectory(context.Background(), virtualPath)
				if err != nil {
					util.Log().Warning("Importing task cannot create user directory %q: %s",
						virtualPath, err)
					continue
				}
				parentFolder = folder

			}

			// 插入文件记录
			_, err := fs.AddFile(context.Background(), parentFolder, &fileHeader)
			if err != nil {
				util.Log().Warning("Importing task cannot insert user file %q: %s",
					object.RelativePath, err)
				if err == filesystem.ErrInsufficientCapacity {
					job.SetErrorMsg("Insufficient storage capacity.", err)
					return
				}
			}

		}
	}
}

// NewImportTask 新建导入任务
func NewImportTask(user, policy uint, src, dst string, recursive bool) (Job, error) {
	creator, err := model.GetActiveUserByID(user)
	if err != nil {
		return nil, err
	}

	newTask := &ImportTask{
		User: &creator,
		TaskProps: ImportProps{
			PolicyID:  policy,
			Recursive: recursive,
			Src:       src,
			Dst:       dst,
		},
	}

	record, err := Record(newTask)
	if err != nil {
		return nil, err
	}
	newTask.TaskModel = record

	return newTask, nil
}

// NewImportTaskFromModel 从数据库记录中恢复导入任务
func NewImportTaskFromModel(task *model.Task) (Job, error) {
	user, err := model.GetActiveUserByID(task.UserID)
	if err != nil {
		return nil, err
	}
	newTask := &ImportTask{
		User:      &user,
		TaskModel: task,
	}

	err = json.Unmarshal([]byte(task.Props), &newTask.TaskProps)
	if err != nil {
		return nil, err
	}

	return newTask, nil
}
