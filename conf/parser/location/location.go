// Copyright 2024 Jakub Zelenka and The WST Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package location

import (
	"fmt"
	"strings"
)

type Item interface {
	String(first bool) string
	Parent() Item
}

type FieldItem struct {
	parent Item
	name   string
}

func (l FieldItem) Parent() Item {
	return l.parent
}

func (l FieldItem) String(first bool) string {
	if first {
		return l.name
	}
	return "." + l.name
}

type IndexItem struct {
	parent Item
	idx    int
}

func (l IndexItem) Parent() Item {
	return l.parent
}

func (l IndexItem) String(first bool) string {
	return fmt.Sprintf("[%d]", l.idx)
}

type Type int

const (
	InvalidType Type = iota
	FieldType
	IndexType
)

type Location struct {
	parent  Item
	locType Type
	field   *FieldItem
	index   *IndexItem
	depth   int
}

func (l *Location) start() {
	if l.locType == FieldType {
		l.parent = l.field
	} else if l.locType == IndexType {
		l.parent = l.index
	}
	l.depth++
}

func (l *Location) end() {
	if l.parent == nil {
		l.depth = 0
		l.locType = InvalidType
	} else {
		switch item := l.parent.(type) {
		case *IndexItem:
			l.locType = IndexType
			l.index = item
		case *FieldItem:
			l.locType = FieldType
			l.field = item
		}
		l.parent = l.parent.Parent()
		l.depth--
	}
}

func (l *Location) StartObject() {
	l.start()
	l.locType = FieldType
	l.field = &FieldItem{parent: l.parent}
}

func (l *Location) EndObject() {
	l.end()
}

func (l *Location) SetField(name string) {
	if l.locType == FieldType {
		l.field.name = name
	}
}

func (l *Location) StartArray() {
	l.start()
	l.locType = IndexType
	l.index = &IndexItem{parent: l.parent, idx: -1}
}

func (l *Location) EndArray() {
	l.end()
}

func (l *Location) SetIndex(idx int) {
	if l.locType == IndexType {
		l.index.idx = idx
	}
}

func (l *Location) String() string {
	idents := make([]string, 0, l.depth)
	var item Item
	if l.locType == FieldType && l.field.name != "" {
		item = l.field
	} else if l.locType == IndexType && l.index.idx >= 0 {
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

func (l *Location) Reset() {
	l.parent = nil
	l.locType = InvalidType
}

func CreateLocation() *Location {
	return &Location{}
}
