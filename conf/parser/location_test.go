package parser

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocation_StartObject(t *testing.T) {
	location := CreateLocation()
	location.StartObject()

	assert.Equal(t, LocationFieldType, location.locType)
	assert.NotNil(t, location.field)
	assert.Nil(t, location.index)
	assert.Equal(t, 1, location.depth)
}

func TestLocation_EndObject(t *testing.T) {
	location := CreateLocation()
	location.StartObject()
	location.SetField("testField")
	location.EndObject()

	assert.Equal(t, 0, location.depth)
	assert.Equal(t, LocationInvalid, location.locType)
	assert.Equal(t, "testField", location.field.name)
}

func TestLocation_SetField(t *testing.T) {
	location := CreateLocation()
	location.StartObject()
	location.SetField("testField")

	assert.Equal(t, "testField", location.field.name)
}

func TestLocation_StartArray(t *testing.T) {
	location := CreateLocation()
	location.StartArray()

	assert.Equal(t, LocationIndexType, location.locType)
	assert.NotNil(t, location.index)
	assert.Nil(t, location.field)
	assert.Equal(t, 1, location.depth)
	assert.Equal(t, -1, location.index.idx)
}

func TestLocation_EndArray(t *testing.T) {
	location := CreateLocation()
	location.StartArray()
	location.SetIndex(0)
	location.EndArray()

	assert.Equal(t, 0, location.depth)
	assert.Equal(t, LocationInvalid, location.locType)
	assert.Equal(t, 0, location.index.idx)
}

func TestLocation_SetIndex(t *testing.T) {
	location := CreateLocation()
	location.StartArray()
	location.SetIndex(5)

	assert.Equal(t, 5, location.index.idx)
}

func TestLocation_String_ObjectField(t *testing.T) {
	location := CreateLocation()
	location.StartObject()
	location.SetField("testField")

	expected := "testField"
	assert.Equal(t, expected, location.String())
}

func TestLocation_String_ArrayIndex(t *testing.T) {
	location := CreateLocation()
	location.StartArray()
	location.SetIndex(3)

	expected := "[3]"
	assert.Equal(t, expected, location.String())
}

func TestLocation_String_Nested(t *testing.T) {
	location := CreateLocation()
	location.StartObject()
	location.SetField("testField")
	location.StartArray()
	location.SetIndex(2)
	location.StartObject()
	location.SetField("innerField")

	expected := "testField[2].innerField"
	assert.Equal(t, expected, location.String())
}

func TestLocation_ComplexNested(t *testing.T) {
	location := CreateLocation()

	// Simulate parsing something like obj.field1.field2[1].field3[2].field4
	location.StartObject()
	location.SetField("field1")
	location.StartObject()
	location.SetField("field2")
	location.StartArray()
	location.SetIndex(1)
	location.StartObject()
	location.SetField("field3")
	location.StartArray()
	location.SetIndex(2)
	location.StartObject()
	location.SetField("field4")

	expected := "field1.field2[1].field3[2].field4"
	assert.Equal(t, expected, location.String())
}
