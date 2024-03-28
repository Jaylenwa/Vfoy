package authn

import (
	"testing"

	"github.com/Jaylenwa/Vfoy/pkg/cache"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	asserts := assert.New(t)
	cache.Set("setting_siteURL", "http://vfoy.org", 0)
	cache.Set("setting_siteName", "Vfoy", 0)
	res, err := NewAuthnInstance()
	asserts.NotNil(res)
	asserts.NoError(err)
}
