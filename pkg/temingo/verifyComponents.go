package temingo

import (
	"errors"
	"strings"
	"text/template"
)

// Verify that each partial has a unique name
func verifyPartials(partialFiles map[string]string) error {
	var (
		err error

		temporaryTemplateEngine     *template.Template
		temporaryTemplateEngineName = "temporaryPartialEngine"
		partialName                 string
		partialLocations            = map[string]string{}
	)

	for partialPath, content := range partialFiles {
		// Checking for duplicate partials
		temporaryTemplateEngine = template.New(temporaryTemplateEngineName) // Create a new temporary template

		// Defining additional template functions
		temporaryTemplateEngine = temporaryTemplateEngine.Funcs(template.FuncMap{
			"concat": tmpl_concat,
		})

		_, err = temporaryTemplateEngine.Parse(content) // Parse the partial into the temporary template engine
		if err != nil {
			return err
		}
		partialName = strings.TrimPrefix(temporaryTemplateEngine.DefinedTemplates(), "; defined templates are: ") // Prefix comes from the offical text.template library
		partialName = strings.ReplaceAll(partialName, "\"", "")                                                   // remove '"'
		partialName = strings.ReplaceAll(partialName, " ", "")                                                    // remove ' '
		for _, subpartialName := range strings.Split(partialName, ",") {                                          // For all partials in this partial file (check if it's name is unique)
			if subpartialName == temporaryTemplateEngineName { // Skip the manually added initial template engine name
				continue
			} else {
				for existingPartialName, existingPartialPath := range partialLocations { // For each partial that already exists
					if subpartialName == existingPartialName { // If new partial would overwrite an existing partial (==same name)
						return errors.New("duplicate partial name '" + subpartialName + "' found in " + partialPath + " and " + existingPartialPath)
					}
				}
				// If the partial is truly new
				partialLocations[subpartialName] = partialPath // Add the partial name to the list. PartialPath is only used to provide a better error message
			}
		}
	}

	return err
}
