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

func ModifyDeps(content string, deps map[string]interface{}) (string, error) {
	if (content == "") {
		return "", nil
	}
	properties := deps["properties"].(map[string]interface{})
	dependencies := deps["dependencies"].([]interface{})
	newProperties := deps["newProperties"].([]interface{})
	newDependencies := deps["newDependencies"].([]interface{})

	parsedPom, err := Parse(content)
	if err != nil {
		return "", err
	}
	//modify properties
	for k, v := range properties {
		props := *parsedPom.Properties
		props.Entries[k] = v.(string)
	}
	//add new properties
	for _, newProps := range newProperties {
		props := *parsedPom.Properties
		name := newProps.(map[string]interface{})["name"]
		version := newProps.(map[string]interface{})["version"]
		props.Entries[name.(string)] = version.(string)
	}
	//modify Dependencies
	for _, dep := range dependencies {
		newDep := dep.(map[string]interface{})
		for i := range *parsedPom.Dependencies {
			if *(*parsedPom.Dependencies)[i].GroupID == newDep["groupId"] && *(*parsedPom.Dependencies)[i].ArtifactID == newDep["artifactId"] {
				newVersion := newDep["version"].(string)
				(*parsedPom.Dependencies)[i].Version = &newVersion
			} 
		}
	}
	//add new dependencies
	for _, dep := range newDependencies {
		newDep := dep.(map[string]interface{})
		groupId, _ := newDep["groupId"].(string)
		artifactId, _ := newDep["artifactId"].(string)
		version, _ := newDep["version"].(string)
		*parsedPom.Dependencies = append(*parsedPom.Dependencies, gopom.Dependency{
			GroupID:	&groupId,
			ArtifactID:	&artifactId,
			Version:	&version,
		})
	}

	//Prepare new deps content
	output, err := xml.MarshalIndent(parsedPom, "", "  ")
	if err != nil {
		log.Printf("Error marshaling POM to XML: %v\n", err)
		return "", nil
	}
	return string(output), nil
}