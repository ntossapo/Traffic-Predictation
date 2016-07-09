package model

import (
	"geo"
)

type Model struct {
	Host 	geo.Point	`bson:host`
	Parent 	[]geo.Point	`bson:parent`
}

func (m *Model) NewInstance(host geo.Point, parent []geo.Point){
	m.Host = host
	m.Parent = parent
}

func (m *Model) CopyInstance() Model{
	result := Model{}
	result.Parent = m.Parent
	result.Host = m.Host
	return result
}

func (m *Model) Append(parent geo.Point){
	if m.Parent == nil{
		m.Parent = make([]geo.Point, 1)
		m.Parent[0] = parent
	}else {
		oldParent := m.Parent
		oldParent = append(oldParent, parent)
		m.Parent = oldParent
	}
}

func (m Model) ContainParent(p geo.Point) bool{
	parents := m.Parent
	for i:=0;i<len(parents);i++{
		if (parents[i].Lat == p.Lat) && (parents[i].Lng == p.Lng){
			return true
		}
	}
	return false
}