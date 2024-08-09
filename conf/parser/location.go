package parser

import (
	"fmt"
	"strings"
)

type LocationItem interface {
	String(first bool) string
	Parent() LocationItem
}

type LocationFieldItem struct {
	parent LocationItem
	name   string
}

func (l LocationFieldItem) Parent() LocationItem {
	return l.parent
}

func (l LocationFieldItem) String(first bool) string {
	if first {
		return l.name
	}
	return "." + l.name
}

type LocationIndexItem struct {
	parent LocationItem
	idx    int
}

func (l LocationIndexItem) Parent() LocationItem {
	return l.parent
}

func (l LocationIndexItem) String(first bool) string {
	return fmt.Sprintf("[%d]", l.idx)
}

type LocationType int

const (
	LocationInvalid LocationType = iota
	LocationFieldType
	LocationIndexType
)

type Location struct {
	parent  LocationItem
	locType LocationType
	field   *LocationFieldItem
	index   *LocationIndexItem
	depth   int
}

func (l *Location) start() {
	if l.locType == LocationFieldType {
		l.parent = l.field
	} else if l.locType == LocationIndexType {
		l.parent = l.index
	}
	l.depth++
}

func (l *Location) end() {
	if l.parent == nil {
		l.depth = 0
		l.locType = LocationInvalid
	} else {
		switch item := l.parent.(type) {
		case *LocationIndexItem:
			l.locType = LocationIndexType
			l.index = item
		case *LocationFieldItem:
			l.locType = LocationFieldType
			l.field = item
		}
		l.parent = l.parent.Parent()
		l.depth--
	}
}

func (l *Location) StartObject() {
	l.start()
	l.locType = LocationFieldType
	l.field = &LocationFieldItem{parent: l.parent}
}

func (l *Location) EndObject() {
	l.end()
}

func (l *Location) SetField(name string) {
	if l.locType == LocationFieldType {
		l.field.name = name
	}
}

func (l *Location) StartArray() {
	l.start()
	l.locType = LocationIndexType
	l.index = &LocationIndexItem{parent: l.parent, idx: -1}
}

func (l *Location) EndArray() {
	l.end()
}

func (l *Location) SetIndex(idx int) {
	if l.locType == LocationIndexType {
		l.index.idx = idx
	}
}

func (l *Location) String() string {
	idents := make([]string, 0, l.depth)
	var item LocationItem
	if l.locType == LocationFieldType && l.field.name != "" {
		item = l.field
	} else if l.locType == LocationIndexType && l.index.idx >= 0 {
		item = l.index
	} else {
		item = l.parent
	}
	for item != nil {
		idents = append(idents, item.String(item.Parent() == nil))
		item = item.Parent()
	}

	// Reverse the order of idents
	for i, j := 0, len(idents)-1; i < j; i, j = i+1, j-1 {
		idents[i], idents[j] = idents[j], idents[i]
	}

	// Concatenate the reversed idents
	return strings.Join(idents, "")
}

func CreateLocation() *Location {
	return &Location{}
}
