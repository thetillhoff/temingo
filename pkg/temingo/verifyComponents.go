package temingo

import (
	"errors"
	"strings"
	"text/template"
)

// Verify that each component has a unique name
func verifyComponents(componentFiles map[string]string) error {
	var (
		err error

		temporaryTemplateEngine     *template.Template
		temporaryTemplateEngineName = "temporaryComponentEngine"
		componentName               string
		componentLocations          = map[string]string{}
	)

	for componentPath, content := range componentFiles {
		// Checking for duplicate components
		temporaryTemplateEngine = template.New(temporaryTemplateEngineName) // Create a new temporary template
		_, err = temporaryTemplateEngine.Parse(content)                     // Parse the component into the temporary template engine
		if err != nil {
			return err
		}
		componentName = strings.TrimPrefix(temporaryTemplateEngine.DefinedTemplates(), "; defined templates are: ") // Prefix comes from the offical text.template library
		componentName = strings.ReplaceAll(componentName, "\"", "")                                                 // remove '"'
		componentName = strings.ReplaceAll(componentName, " ", "")                                                  // remove ' '
		for _, subcomponentName := range strings.Split(componentName, ",") {                                        // For all components in this component file (check if it's name is unique)
			if subcomponentName == temporaryTemplateEngineName { // Skip the manually added initial template engine name
				continue
			} else {
				for existingComponentName, existingComponentPath := range componentLocations { // For each component that already exists
					if subcomponentName == existingComponentName { // If new component would overwrite an existing component (==same name)
						return errors.New("duplicate component name '" + subcomponentName + "' found in " + componentPath + " and " + existingComponentPath)
					}
				}
				// If the component is truly new
				componentLocations[subcomponentName] = componentPath // Add the component name to the list. ComponentPath is only used to provide a better error message
			}
		}
	}

	return err
}
