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
	"github.com/Jaylenwa/Vfoy/service/knowledge_base/vo"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type KnowledgeBase struct {
	Path string `uri:"path" json:"path" binding:"required,min=1,max=65535"`
}

func (kb *KnowledgeBase) CreateKnowledgeBase(c *gin.Context) serializer.Response {

	// 获取token
	token, err := kb.getToken()
	if err != nil {
		return serializer.Err(serializer.CodeKnowledgeBaseErr, "Get token error", nil)
	}

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
	split, err := kb.split(token, body, writer)
	if err != nil {
		return serializer.Err(serializer.CodeKnowledgeBaseErr, "Error writing to multipart writer", err)
	}

	u, _ := uuid.NewRandom()
	ustr := u.String()
	docSet := vo.DocSetCreateReq{
		Name: ustr,
		Desc: ustr,
	}

	documents := make([]vo.Document, 0)

	for _, data := range split.Data {
		document := vo.Document{
			Name: data.Name,
		}

		paragraphs := make([]vo.Paragraph, 0)
		for _, v := range data.Content {
			paragraph := vo.Paragraph{
				Title:    v.Title,
				Content:  v.Content,
				IsActive: true,
			}
			paragraph.ProblemList = make([]vo.Problem, 0)
			paragraphs = append(paragraphs, paragraph)
		}

		document.Paragraphs = paragraphs
		documents = append(documents, document)
	}

	docSet.Documents = documents

	// 创建知识库
	docSetRes, err := kb.createDocSet(token, docSet)
	if err != nil {
		return serializer.Err(serializer.CodeKnowledgeBaseErr, "Failed to create knowledge base", err)
	}

	app := vo.ApplicationCreateReq{}.DocSetCreateRes2ApplicationCreateReq(docSetRes)

	app.Desc = ustr

	modelList, err := kb.GetModel(token)
	if err != nil {
		return serializer.Err(serializer.CodeKnowledgeBaseErr, "Failed to obtain model", err)
	}

	if len(modelList.Data) == 0 {
		return serializer.Err(serializer.CodeKnowledgeBaseErr, "model list is empty", nil)
	}
	app.ModelID = modelList.Data[0].ID

	// 创建应用
	err = kb.createApplication(token, app)
	if err != nil {
		return serializer.Err(serializer.CodeKnowledgeBaseErr, "Create application failed", err)
	}

	// 获取应用信息
	appInfo, err := kb.getApps(token, app.Name, app.Desc)
	if err != nil {
		return serializer.Err(serializer.CodeKnowledgeBaseErr, "Create application failed", err)
	}

	// 获取应用AccessToken信息
	assessToken, err := kb.getAppAccessToken(token, appInfo.Data[0].ID)
	if err != nil {
		return serializer.Err(serializer.CodeKnowledgeBaseErr, "Create application failed", err)
	}

	return serializer.Response{Data: map[string]interface{}{
		"open_url": model.GetSettingByName("knowledge_base_url") + "/ui/chat/" + assessToken.Data.AccessToken,
	}}
}

func (kb *KnowledgeBase) getAppAccessToken(token string, applicationId string) (res vo.ApplicationAccessToken, err error) {

	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": token,
	}

	url := model.GetSettingByName("knowledge_base_url") + "/api/application"

	url = fmt.Sprintf(url+"/%s/access_token", applicationId)

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
		return res, errors.New(fmt.Sprintf("Error : %v", res.Message))
	}

	return
}

func (kb *KnowledgeBase) getApps(token string, name string, desc string) (res vo.ApplicationList, err error) {

	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": token,
	}

	url := model.GetSettingByName("knowledge_base_url") + "/api/application"

	url = fmt.Sprintf(url+"?name=%s&desc=%s", name, desc)

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
		return res, errors.New(fmt.Sprintf("Error : %v", res.Message))
	}

	return
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
func (kb *KnowledgeBase) split(token string, body *bytes.Buffer, w *multipart.Writer) (res *vo.Split, err error) {

	header := map[string]string{
		"Content-Type":  w.FormDataContentType(),
		"Authorization": token,
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
func (kb *KnowledgeBase) createDocSet(token string, docSet vo.DocSetCreateReq) (res vo.DocSetCreateRes, err error) {

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
		return vo.DocSetCreateRes{}, errors.New(res.Message)
	}

	return
}

func (kb *KnowledgeBase) getToken() (token string, err error) {
	header := map[string]string{
		"Content-Type": "application/json",
	}

	username := model.GetSettingByName("kb_user")

	pwd := model.GetSettingByName("kb_pwd")

	reqBody, err := json.Marshal(map[string]string{
		"username": username,
		"password": pwd,
	})
	
	if err != nil {
		return
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

	var maxKBResp vo.MaxKBResp

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

func (kb *KnowledgeBase) createApplication(token string, req vo.ApplicationCreateReq) (err error) {

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

	var maxKBResp vo.ApplicationCreateRes

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
func (kb *KnowledgeBase) GetModel(token string) (res vo.ModelList, err error) {

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
		return vo.ModelList{}, errors.New(fmt.Sprintf("Error : %v", res.Message))
	}

	return
}
