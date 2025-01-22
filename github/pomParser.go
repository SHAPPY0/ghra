package github

import (
	"log"
	"encoding/xml"
	"github.com/vifraa/gopom"
)

func Parse(content string) (*gopom.Project, error) {
	if content != "" {
		var parsedPom gopom.Project
		if err := xml.Unmarshal([]byte(content), &parsedPom); err != nil {
			return nil, err
		}
		return &parsedPom, nil
	} else {
		return nil, nil
	}
}

func ModifyDeps(content string, newDeps map[string]interface{}) (string, error) {
	if (content == "") {
		return "", nil
	}
	properties := newDeps["properties"].(map[string]interface{})
	dependencies := newDeps["dependencies"].([]interface{})

	parsedContent, err := Parse(content)
	if err != nil {
		return "", err
	}
	//modify properties
	for k, v := range properties {
		props := *parsedContent.Properties
		props.Entries[k] = v.(string)
	}
	//modify Dependencies
	for _, dep := range dependencies {
		newDep := dep.(map[string]interface{})
		for _, dep := range *parsedContent.Dependencies {
			if *dep.GroupID == newDep["groupId"] && *dep.ArtifactID == newDep["artifactId"] {
				*dep.Version = newDep["version"].(string)
			}
		}
	}
	//Prepare new deps content
	output, err := xml.MarshalIndent(parsedContent, "", "  ")
	if err != nil {
		log.Printf("Error marshaling POM to XML: %v\n", err)
		return "", nil
	}
	return string(output), nil
}