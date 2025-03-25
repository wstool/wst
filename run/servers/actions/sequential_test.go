package actions

import (
	"github.com/stretchr/testify/assert"
	"github.com/wstool/wst/conf/types"
	"testing"
)

func Test_nativeMaker_makeSequentialActions(t *testing.T) {
	configActions := map[string]types.ServerSequentialAction{
		"action1": {
			Actions: []types.Action{
				&types.RequestAction{Service: "service1"},
				&types.RequestAction{Service: "service2"},
			},
		},
		"action2": {
			Actions: []types.Action{
				&types.RequestAction{Service: "service3"},
			},
		},
	}

	maker := &nativeMaker{}
	result := maker.makeSequentialActions(configActions)

	assert := assert.New(t)
	assert.Len(result, len(configActions))

	// Check action1
	action1, exists := result["action1"]
	assert.True(exists)
	assert.Equal(configActions["action1"].Actions, action1.Actions())

	// Check action2
	action2, exists := result["action2"]
	assert.True(exists)
	assert.Equal(configActions["action2"].Actions, action2.Actions())
}

func Test_nativeSequentialAction_Actions(t *testing.T) {
	expectedActions := []types.Action{
		&types.RequestAction{Service: "service1"},
		&types.RequestAction{Service: "service2"},
	}

	action := &nativeSequentialAction{
		actions: expectedActions,
	}

	assert.Equal(t, expectedActions, action.Actions())
}
