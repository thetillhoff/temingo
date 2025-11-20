package temingo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitProject(t *testing.T) {
	tests := []struct {
		name        string
		projectType string
		setup       func(tmpDir string) error
		wantErr     bool
		errContains string
		description string
	}{
		{
			name:        "Valid project type - should succeed",
			projectType: "example",
			setup: func(tmpDir string) error {
				// Empty directory - should succeed
				return nil
			},
			wantErr:     false,
			description: "When directory is empty and project type is valid, initialization should succeed",
		},
		{
			name:        "Invalid project type - should fail",
			projectType: "invalid_type",
			setup: func(tmpDir string) error {
				// Empty directory
				return nil
			},
			wantErr:     true,
			errContains: "not a valid project type",
			description: "When project type is invalid, initialization should fail",
		},
		{
			name:        "Directory is not empty - should fail",
			projectType: "example",
			setup: func(tmpDir string) error {
				// Create a file in the directory
				testFile := filepath.Join(tmpDir, "existing_file.txt")
				return os.WriteFile(testFile, []byte("test content"), 0644)
			},
			wantErr:     true,
			errContains: "the directory is not empty",
			description: "When directory contains files, initialization should fail",
		},
		{
			name:        "Directory contains subdirectory - should fail",
			projectType: "example",
			setup: func(tmpDir string) error {
				// Create a subdirectory
				subDir := filepath.Join(tmpDir, "existing_dir")
				return os.MkdirAll(subDir, 0755)
			},
			wantErr:     true,
			errContains: "the directory is not empty",
			description: "When directory contains subdirectories, initialization should fail",
		},
		{
			name:        "Directory only contains output directory - should succeed",
			projectType: "example",
			setup: func(tmpDir string) error {
				// Create only the output directory (should be allowed)
				outputDir := filepath.Join(tmpDir, "output")
				return os.MkdirAll(outputDir, 0755)
			},
			wantErr:     false,
			description: "When directory only contains the output directory, initialization should succeed",
		},
		{
			name:        "InputDir already exists - should fail",
			projectType: "example",
			setup: func(tmpDir string) error {
				// Create inputDir
				inputDir := filepath.Join(tmpDir, "src")
				return os.MkdirAll(inputDir, 0755)
			},
			wantErr:     true,
			errContains: "the folder 'src/' already exists",
			description: "When inputDir already exists, initialization should fail",
		},
		{
			name:        "Temingoignore already exists - should fail",
			projectType: "example",
			setup: func(tmpDir string) error {
				// Create .temingoignore file
				ignoreFile := filepath.Join(tmpDir, ".temingoignore")
				return os.WriteFile(ignoreFile, []byte("# ignore"), 0644)
			},
			wantErr:     true,
			errContains: "the file '.temingoignore' already exists",
			description: "When .temingoignore already exists, initialization should fail",
		},
		{
			name:        "Directory contains file and output directory - should fail",
			projectType: "example",
			setup: func(tmpDir string) error {
				// Create a file
				testFile := filepath.Join(tmpDir, "existing_file.txt")
				if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
					return err
				}
				// Create output directory
				outputDir := filepath.Join(tmpDir, "output")
				return os.MkdirAll(outputDir, 0755)
			},
			wantErr:     true,
			errContains: "the directory is not empty",
			description: "When directory contains files other than output directory, initialization should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new tmpDir for each test case to ensure isolation
			tmpDir := t.TempDir()

			// Setup test environment
			if tt.setup != nil {
				if err := tt.setup(tmpDir); err != nil {
					t.Fatalf("Test setup failed: %v", err)
				}
			}

			engine := DefaultEngine()
			engine.InputDir = "src" + string(filepath.Separator)
			engine.OutputDir = "output" + string(filepath.Separator)
			engine.TemingoignorePath = ".temingoignore"

			// Run InitProject with tmpDir as targetDir
			err := engine.InitProject(tt.projectType, tmpDir)

			if tt.wantErr {
				if err == nil {
					t.Errorf("InitProject() expected error but got nil (%s)", tt.description)
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("InitProject() error = %v, expected error to contain %q (%s)", err, tt.errContains, tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("InitProject() unexpected error: %v (%s)", err, tt.description)
				} else {
					// Verify that files were created
					// Check if inputDir was created
					srcPath := filepath.Join(tmpDir, "src")
					if _, err := os.Stat(srcPath); os.IsNotExist(err) {
						t.Errorf("InitProject() should have created inputDir but it doesn't exist (%s)", tt.description)
					}
					// Check if .temingoignore was created
					ignorePath := filepath.Join(tmpDir, ".temingoignore")
					if _, err := os.Stat(ignorePath); os.IsNotExist(err) {
						t.Errorf("InitProject() should have created .temingoignore but it doesn't exist (%s)", tt.description)
					}
				}
			}
		})
	}
}

func TestInitProject_AllProjectTypes(t *testing.T) {
	// Test that all valid project types can be initialized
	projectTypes := ProjectTypes()
	if len(projectTypes) == 0 {
		t.Fatal("No project types available")
	}

	for _, projectType := range projectTypes {
		t.Run("ProjectType_"+projectType, func(t *testing.T) {
			// Create a new tmpDir for each test case to ensure isolation
			tmpDir := t.TempDir()

			engine := DefaultEngine()
			engine.InputDir = "src/"
			engine.OutputDir = "output/"
			engine.TemingoignorePath = ".temingoignore"

			err := engine.InitProject(projectType, tmpDir)
			if err != nil {
				t.Errorf("InitProject() failed for project type %q: %v", projectType, err)
			}

			// Verify that files were created
			srcPath := filepath.Join(tmpDir, "src")
			if _, err := os.Stat(srcPath); os.IsNotExist(err) {
				t.Errorf("InitProject() should have created inputDir for project type %q", projectType)
			}
		})
	}
}
