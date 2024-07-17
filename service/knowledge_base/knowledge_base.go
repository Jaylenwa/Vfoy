package knowledge_base

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	model "github.com/Jaylenwa/Vfoy/models"
	"github.com/Jaylenwa/Vfoy/pkg/filesystem"
	"github.com/Jaylenwa/Vfoy/pkg/request"
	"github.com/Jaylenwa/Vfoy/pkg/serializer"
	"github.com/Jaylenwa/Vfoy/service/knowledge_base/entity"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type KnowledgeBase struct {
	Path string `uri:"path" json:"path" binding:"required,min=1,max=65535"`
}

func (kb *KnowledgeBase) CreateKnowledgeBase(c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 获取子项目
	files, err := fs.ListFile(c, kb.Path, nil)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	filesPath := make([]string, 0)
	for _, file := range files {
		ext := filepath.Ext(file.SourceName)
		switch strings.ToLower(ext) {
		case ".txt", ".md", ".markdown", ".pdf", ".docx", ".html":
			filesPath = append(filesPath, file.SourceName)
		}
	}

	// 创建一个bufWriter来存储请求体的内容
	body := &bytes.Buffer{}

	if len(filesPath) == 0 {
		return serializer.Err(serializer.CodeKnowledgeBaseErr, "No supported files found", nil)
	}

	// 文档提取
	writer, err := kb.docExtract(filesPath, body)
	if err != nil {
		return serializer.Err(serializer.CodeKnowledgeBaseErr, "Error creating multipart writer", err)
	}

	// 获取文档分片
	split, err := kb.split(body, writer)
	if err != nil {
		return serializer.Err(serializer.CodeKnowledgeBaseErr, "Error writing to multipart writer", err)
	}

	docSetName := strings.Replace(kb.Path, "/", "_", -1)
	u, _ := uuid.NewRandom()
	docSet := entity.DocSetCreateReq{
		Name: docSetName + u.String(),
		Desc: "文件夹:" + docSetName + "生成的文档库",
	}

	documents := make([]entity.Document, 0)

	for _, data := range split.Data {
		document := entity.Document{
			Name: data.Name,
		}

		paragraphs := make([]entity.Paragraph, 0)
		for _, v := range data.Content {
			paragraph := entity.Paragraph{
				Title:    v.Title,
				Content:  v.Content,
				IsActive: true,
			}
			paragraph.ProblemList = make([]entity.Problem, 0)
			paragraphs = append(paragraphs, paragraph)
		}

		document.Paragraphs = paragraphs
		documents = append(documents, document)
	}

	docSet.Documents = documents

	// 创建知识库
	docSetRes, err := kb.createDocSet(docSet)
	if err != nil {
		return serializer.Err(serializer.CodeKnowledgeBaseErr, "Failed to create knowledge base", err)
	}

	app := entity.ApplicationCreateReq{}.DocSetCreateRes2ApplicationCreateReq(docSetRes)

	app.Desc = "文件夹:" + strings.Replace(kb.Path, "/", "_", -1) + "生成的应用"

	modelList, err := kb.GetModel()
	if err != nil {
		return serializer.Err(serializer.CodeKnowledgeBaseErr, "Failed to obtain model", err)
	}

	if len(modelList.Data) == 0 {
		return serializer.Err(serializer.CodeKnowledgeBaseErr, "model list is empty", nil)
	}
	app.ModelID = modelList.Data[0].ID

	// 创建应用
	err = kb.createApplication(app)
	if err != nil {
		return serializer.Err(serializer.CodeKnowledgeBaseErr, "Create application failed", err)
	}

	return serializer.Response{}
}

func (kb *KnowledgeBase) fileCollection(dirPath string) (filePath []string, err error) {
	filePath = make([]string, 0)

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Error:", err)
			return err
		}
		if !info.IsDir() {
			ext := filepath.Ext(path)
			switch strings.ToLower(ext) {
			case ".txt", ".md", ".markdown", ".pdf", ".docx", ".html":
				filePath = append(filePath, path)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println("Walk error:", err)
	}

	return
}

// Split 文档分片
func (kb *KnowledgeBase) split(body *bytes.Buffer, w *multipart.Writer) (res *entity.Split, err error) {

	header := map[string]string{
		"Content-Type": w.FormDataContentType(),
	}

	url := model.GetSettingByName("knowledge_base_url") + "/api/dataset/document/split"

	resp, err := request.NewHttpClient().Post(context.Background(), url, header, body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if err = json.Unmarshal(bodyBytes, &res); err != nil {
		return
	}

	return
}

func (kb *KnowledgeBase) docExtract(filesPath []string, body io.Writer) (writer *multipart.Writer, err error) {
	writer = multipart.NewWriter(body)
	// 假设我们有两个文件需要上传

	for _, path := range filesPath {
		f, err := os.Open(path)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return writer, err
		}
		defer f.Close()

		// 创建一个multipart form-data的file字段
		part, err := writer.CreateFormFile("file", path)
		if err != nil {
			fmt.Println("Error creating form-data part:", err)
			return writer, err
		}

		// 将文件内容写入到form-data字段
		_, err = io.Copy(part, f)
		if err != nil {
			fmt.Println("Error copying file to buffer:", err)
			return writer, err
		}
	}

	// 关闭writer以完成boundary的写入
	writer.Close()

	return writer, err
}

// 创建知识库
func (kb *KnowledgeBase) createDocSet(docSet entity.DocSetCreateReq) (res entity.DocSetCreateRes, err error) {

	// 获取token
	token, err := kb.getToken()
	if err != nil {
		return
	}

	docSetBytes, err := json.Marshal(docSet)
	if err != nil {
		return
	}

	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": token,
	}

	url := model.GetSettingByName("knowledge_base_url") + "/api/dataset"

	resp, err := request.NewHttpClient().Post(context.Background(), url, header, bytes.NewReader(docSetBytes))
	if err != nil {
		return
	}

	defer resp.Body.Close()

	// 处理响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %s", err)
		return
	}

	if err = json.Unmarshal(body, &res); err != nil {
		return
	}

	if res.Code != 200 {
		return entity.DocSetCreateRes{}, errors.New(res.Message)
	}

	return
}

func (kb *KnowledgeBase) getToken() (token string, err error) {
	reqBody := []byte(`{"username":"admin","password":"wang1318248167@"}`)
	header := map[string]string{
		"Content-Type": "application/json",
	}

	url := model.GetSettingByName("knowledge_base_url") + "/api/user/login"

	// 获取token
	resp, err := request.NewHttpClient().Post(context.Background(), url, header, bytes.NewReader(reqBody))
	if err != nil {
		return
	}

	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var maxKBResp entity.MaxKBResp

	if err = json.Unmarshal(respBytes, &maxKBResp); err != nil {
		return
	}

	token = maxKBResp.Data

	if maxKBResp.Code != 200 {
		err = errors.New(fmt.Sprintf("Error : %v", maxKBResp.Message))
		return
	}

	return
}

func (kb *KnowledgeBase) createApplication(req entity.ApplicationCreateReq) (err error) {

	token, err := kb.getToken()
	if err != nil {
		return
	}

	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": token,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return
	}

	url := model.GetSettingByName("knowledge_base_url") + "/api/application"

	resp, err := request.NewHttpClient().Post(context.Background(), url, header, bytes.NewReader(reqBytes))
	if err != nil {
		return
	}

	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var maxKBResp entity.ApplicationCreateRes

	err = json.Unmarshal(respBytes, &maxKBResp)
	if err != nil {
		return
	}

	if maxKBResp.Code != 200 {
		err = errors.New(fmt.Sprintf("Error : %v", maxKBResp.Message))
	}

	return
}

// GetModel 获取模型
func (kb *KnowledgeBase) GetModel() (res entity.ModelList, err error) {
	token, err := kb.getToken()
	if err != nil {
		return
	}

	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": token,
	}

	url := model.GetSettingByName("knowledge_base_url") + "/api/model"

	resp, err := request.NewHttpClient().Get(context.Background(), url, header)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(respBytes, &res)
	if err != nil {
		return
	}

	if res.Code != 200 {
		return entity.ModelList{}, errors.New(fmt.Sprintf("Error : %v", res.Message))
	}

	return
}
