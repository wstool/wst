package actions

import (
	"github.com/stretchr/testify/assert"
	"github.com/wstool/wst/run/expectations"
	"github.com/wstool/wst/run/parameters"
	"testing"
)

func TestExpectOutputAction(t *testing.T) {
	p := testParams(t, 1)
	params := parameters.Parameters{"param2": p[0]}
	outputExp := &expectations.OutputExpectation{OutputType: expectations.OutputTypeAny}

	action := expectOutputAction{
		parameters:        params,
		outputExpectation: outputExp,
	}

	t.Run("Parameters", func(t *testing.T) {
		gotParams := action.Parameters()
		assert.Equal(t, params, gotParams, "expectOutputAction.Parameters() should return the correct parameters")
	})

	t.Run("OutputExpectation", func(t *testing.T) {
		gotOutputExp := action.OutputExpectation()
		assert.Equal(t, outputExp, gotOutputExp, "expectOutputAction.OutputExpectation() should return the correct output expectation")
	})

	t.Run("ResponseExpectation", func(t *testing.T) {
		gotResponseExp := action.ResponseExpectation()
		assert.Nil(t, gotResponseExp, "expectOutputAction.ResponseExpectation() should return nil")
	})
}

func TestExpectResponseAction(t *testing.T) {
	p := testParams(t, 1)
	params := parameters.Parameters{"param2": p[0]}
	responseExp := &expectations.ResponseExpectation{BodyContent: "data"}

	action := expectResponseAction{
		parameters:          params,
		responseExpectation: responseExp,
	}

	t.Run("Parameters", func(t *testing.T) {
		gotParams := action.Parameters()
		assert.Equal(t, params, gotParams, "expectResponseAction.Parameters() should return the correct parameters")
	})

	t.Run("OutputExpectation", func(t *testing.T) {
		gotOutputExp := action.OutputExpectation()
		assert.Nil(t, gotOutputExp, "expectResponseAction.OutputExpectation() should return nil")
	})

	t.Run("ResponseExpectation", func(t *testing.T) {
		gotResponseExp := action.ResponseExpectation()
		assert.Equal(t, responseExp, gotResponseExp, "expectResponseAction.ResponseExpectation() should return the correct response expectation")
	})
}
