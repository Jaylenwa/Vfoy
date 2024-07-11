package entity

// ContentItem 代表文件内容中的条目
type ContentItem struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// FileContent 代表文件的内容列表
type FileContent []ContentItem

// FileInfo 代表单个文件的信息
type FileInfo struct {
	Name    string      `json:"name"`
	Content FileContent `json:"content"`
}

type Split struct {
	Data []FileInfo `json:"data"`
}
