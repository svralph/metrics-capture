package goply

import (
	"testing"
)
import "strings"

func TestNew(t *testing.T) {
	plyData := `ply
format ascii 1.0           
comment made by Greg Turk  
comment this file is a cube
element vertex 3      
property float x           
property float y           
property float z           
element face 1    
property list uchar int vertex_index
end_header
0 0 0
0 1 0
1 0 0
3 0 1 2`

	reader := strings.NewReader(plyData)
	ply := New(reader)

	if len(ply.data["vertex"]) != 3 {
		t.Errorf("Read incorrect number of vertices, expect %d, got %d", 3, len(ply.data["face"]))
	}
	if len(ply.data["face"]) != 1 {
		t.Errorf("Read incorrect number of faces, expect %d, got %d", 1, len(ply.data["face"]))
	}
}

func TestPly_Elements(t *testing.T) {
	plyData := `ply
format ascii 1.0           
comment made by Greg Turk  
comment this file is a cube
element vertex 3      
property float x           
property float y           
property float z           
element face 1    
property list uchar int vertex_index
end_header
0 0 0
0 1 0
1 0 0
3 0 1 2`

	reader := strings.NewReader(plyData)
	ply := New(reader)

	vertices := ply.Elements("vertex")
	if len(vertices) != 3 {
		t.Errorf("Expected to retrieve %d elements, got %d", 3, len(vertices))
	}
}

func TestPlyElement_Property(t *testing.T) {
	plyData := `ply
format ascii 1.0           
comment made by Greg Turk  
comment this file is a cube
element vertex 3      
property float x           
property float y           
property float z           
element face 1    
property list uchar int vertex_index
end_header
0 0 0
0 1 0
1 0 0
3 0 1 2`

	reader := strings.NewReader(plyData)
	ply := New(reader)

	vertices := ply.Elements("vertex")
	p := vertices[2].Property("x").(float32)

	if p != 1.0 {
		t.Errorf("Expected to retrieve %f, got %f", 1.0, p)
	}
}
