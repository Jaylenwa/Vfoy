package serializer

import (
	"testing"

	model "github.com/Jaylenwa/Vfoy/models"
	"github.com/stretchr/testify/assert"
)

func TestBuildObjectList(t *testing.T) {
	a := assert.New(t)
	res := BuildObjectList(1, []Object{{}, {}}, &model.Policy{})
	a.NotEmpty(res.Parent)
	a.NotNil(res.Policy)
	a.Len(res.Objects, 2)
}
