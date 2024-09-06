package output

import (
	"github.com/stretchr/testify/assert"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	"testing"
)

func TestCreateMaker(t *testing.T) {
	fnd := appMocks.NewMockFoundation(t)
	maker := CreateMaker(fnd)
	assert.NotNil(t, maker)
	nm, ok := maker.(*nativeMaker)
	assert.True(t, ok)
	assert.Equal(t, fnd, nm.fnd)
}

func Test_nativeMaker_MakeCollector(t *testing.T) {
	fnd := appMocks.NewMockFoundation(t)
	maker := CreateMaker(fnd)
	assert.NotNil(t, maker)
	c := maker.MakeCollector("tid")
	assert.NotNil(t, c)
	bc, ok := c.(*BufferedCollector)
	assert.True(t, ok)
	assert.Equal(t, fnd, bc.fnd)
	assert.Equal(t, "tid", bc.tid)
}
