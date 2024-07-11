package controllers

import (
	// "context"

	"github.com/Jaylenwa/Vfoy/service/knowledge_base"
	"github.com/gin-gonic/gin"
)

// CreateKnowledgeBase 创建知识库
func CreateKnowledgeBase(c *gin.Context) {
	var service knowledge_base.KnowledgeBase
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.CreateKnowledgeBase(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}
