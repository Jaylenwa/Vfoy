package cluster

import (
	"testing"

	model "github.com/Jaylenwa/Vfoy/models"
	"github.com/stretchr/testify/assert"
)

func TestNewNodeFromDBModel(t *testing.T) {
	a := assert.New(t)
	a.IsType(&SlaveNode{}, NewNodeFromDBModel(&model.Node{
		Type: model.SlaveNodeType,
	}))
	a.IsType(&MasterNode{}, NewNodeFromDBModel(&model.Node{
		Type: model.MasterNodeType,
	}))
}
