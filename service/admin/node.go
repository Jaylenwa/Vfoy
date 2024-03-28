package admin

import (
	"strings"

	model "github.com/Jaylenwa/Vfoy/models"
	"github.com/Jaylenwa/Vfoy/pkg/cluster"
	"github.com/Jaylenwa/Vfoy/pkg/serializer"
)

// AddNodeService 节点添加服务
type AddNodeService struct {
	Node model.Node `json:"node" binding:"required"`
}

// Add 添加节点
func (service *AddNodeService) Add() serializer.Response {
	if service.Node.ID > 0 {
		if err := model.DB.Save(&service.Node).Error; err != nil {
			return serializer.DBErr("Failed to save node record", err)
		}
	} else {
		if err := model.DB.Create(&service.Node).Error; err != nil {
			return serializer.DBErr("Failed to create node record", err)
		}
	}

	if service.Node.Status == model.NodeActive {
		cluster.Default.Add(&service.Node)
	}

	return serializer.Response{Data: service.Node.ID}
}

// Nodes 列出从机节点
func (service *AdminListService) Nodes() serializer.Response {
	var res []model.Node
	total := 0

	tx := model.DB.Model(&model.Node{})
	if service.OrderBy != "" {
		tx = tx.Order(service.OrderBy)
	}

	for k, v := range service.Conditions {
		tx = tx.Where(k+" = ?", v)
	}

	if len(service.Searches) > 0 {
		search := ""
		for k, v := range service.Searches {
			search += k + " like '%" + v + "%' OR "
		}
		search = strings.TrimSuffix(search, " OR ")
		tx = tx.Where(search)
	}

	// 计算总数用于分页
	tx.Count(&total)

	// 查询记录
	tx.Limit(service.PageSize).Offset((service.Page - 1) * service.PageSize).Find(&res)

	isActive := make(map[uint]bool)
	for i := 0; i < len(res); i++ {
		if node := cluster.Default.GetNodeByID(res[i].ID); node != nil {
			isActive[res[i].ID] = node.IsActive()
		}
	}

	return serializer.Response{Data: map[string]interface{}{
		"total":  total,
		"items":  res,
		"active": isActive,
	}}
}

// ToggleNodeService 开关节点服务
type ToggleNodeService struct {
	ID      uint             `uri:"id"`
	Desired model.NodeStatus `uri:"desired"`
}

// Toggle 开关节点
func (service *ToggleNodeService) Toggle() serializer.Response {
	node, err := model.GetNodeByID(service.ID)
	if err != nil {
		return serializer.DBErr("Node not found", err)
	}

	// 是否为系统节点
	if node.ID <= 1 {
		return serializer.Err(serializer.CodeInvalidActionOnSystemNode, "", err)
	}

	if err = node.SetStatus(service.Desired); err != nil {
		return serializer.DBErr("Failed to change node status", err)
	}

	if service.Desired == model.NodeActive {
		cluster.Default.Add(&node)
	} else {
		cluster.Default.Delete(node.ID)
	}

	return serializer.Response{}
}

// NodeService 节点ID服务
type NodeService struct {
	ID uint `uri:"id" json:"id" binding:"required"`
}

// Delete 删除节点
func (service *NodeService) Delete() serializer.Response {
	// 查找用户组
	node, err := model.GetNodeByID(service.ID)
	if err != nil {
		return serializer.DBErr("Node record not found", err)
	}

	// 是否为系统节点
	if node.ID <= 1 {
		return serializer.Err(serializer.CodeInvalidActionOnSystemNode, "", err)
	}

	cluster.Default.Delete(node.ID)
	if err := model.DB.Delete(&node).Error; err != nil {
		return serializer.DBErr("Failed to delete node record", err)
	}

	return serializer.Response{}
}

// Get 获取节点详情
func (service *NodeService) Get() serializer.Response {
	node, err := model.GetNodeByID(service.ID)
	if err != nil {
		return serializer.DBErr("Node not exist", err)
	}

	return serializer.Response{Data: node}
}
