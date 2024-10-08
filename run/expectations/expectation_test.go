package expectations

import (
	"github.com/stretchr/testify/assert"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	parametersMocks "github.com/wstool/wst/mocks/generated/run/parameters"
	"testing"
)

func TestCreateMaker(t *testing.T) {
	// Create mock instances for dependencies
	fndMock := appMocks.NewMockFoundation(t)
	parametersMakerMock := parametersMocks.NewMockMaker(t)

	// Call the function under test
	maker := CreateMaker(fndMock, parametersMakerMock)

	// Assert that the returned maker is correctly configured
	assert.NotNil(t, maker, "Maker should not be nil")
	nativeMaker, ok := maker.(*nativeMaker)
	assert.True(t, ok, "Returned maker should be of type *nativeMaker")
	assert.Equal(t, fndMock, nativeMaker.fnd, "Foundation should be set correctly")
	assert.Equal(t, parametersMakerMock, nativeMaker.parametersMaker, "Parameters maker should be set correctly")
}
