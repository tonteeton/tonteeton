package eprojects

import (
	"encoding/json"
	"os"
)

type Project struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Projects []Project

func LoadProjects(filename string) (Projects, error) {
	var projects Projects

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &projects)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

// Len returns the number of projects in database.
func (p Projects) Len() int {
	return len(p)
}

// GetByIndex returns the project at the specified index.
func (p Projects) GetByIndex(index int) *Project {
	if index < 0 || index >= len(p) {
		return nil
	}
	return &p[index]
}
