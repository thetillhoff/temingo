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

			// Create engine with absolute paths to avoid needing to change working directory
			engine := DefaultEngine()
			engine.InputDir = filepath.Join(tmpDir, "src") + string(filepath.Separator)
			engine.OutputDir = filepath.Join(tmpDir, "output") + string(filepath.Separator)
			engine.TemingoignorePath = filepath.Join(tmpDir, ".temingoignore")

			// Temporarily change to tmpDir for InitProject (it checks current working directory)
			// We need to do this because InitProject uses os.Getwd() to check if directory is empty
			originalCwd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get current working directory: %v", err)
			}
			defer func() {
				// Restore original directory in defer to ensure it happens even on panic
				if err := os.Chdir(originalCwd); err != nil {
					t.Logf("Warning: Failed to restore original working directory: %v", err)
				}
			}()

			// Use a mutex-like approach: change directory only for this test
			// Since Go tests run in parallel by default, we need to ensure this is safe
			// However, os.Chdir is process-wide, so we need to be careful
			// The best approach is to make InitProject accept a base directory parameter
			// For now, we'll change directory but this test should not run in parallel
			// We'll add t.Parallel() = false implicitly by not calling it
			err = os.Chdir(tmpDir)
			if err != nil {
				t.Fatalf("Failed to change to temp directory: %v", err)
			}

			// Update engine paths to be relative to tmpDir (since we changed to it)
			engine.InputDir = "src" + string(filepath.Separator)
			engine.OutputDir = "output" + string(filepath.Separator)
			engine.TemingoignorePath = ".temingoignore"

			// Run InitProject
			err = engine.InitProject(tt.projectType)

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
					if _, err := os.Stat("src"); os.IsNotExist(err) {
						t.Errorf("InitProject() should have created inputDir but it doesn't exist (%s)", tt.description)
					}
					// Check if .temingoignore was created
					if _, err := os.Stat(".temingoignore"); os.IsNotExist(err) {
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

			// Temporarily change to tmpDir for InitProject (it checks current working directory)
			originalCwd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get current working directory: %v", err)
			}
			defer func() {
				// Restore original directory in defer to ensure it happens even on panic
				if err := os.Chdir(originalCwd); err != nil {
					t.Logf("Warning: Failed to restore original working directory: %v", err)
				}
			}()

			err = os.Chdir(tmpDir)
			if err != nil {
				t.Fatalf("Failed to change to temp directory: %v", err)
			}

			engine := DefaultEngine()
			engine.InputDir = "src/"
			engine.OutputDir = "output/"
			engine.TemingoignorePath = ".temingoignore"

			err = engine.InitProject(projectType)
			if err != nil {
				t.Errorf("InitProject() failed for project type %q: %v", projectType, err)
			}

			// Verify that files were created
			if _, err := os.Stat("src"); os.IsNotExist(err) {
				t.Errorf("InitProject() should have created inputDir for project type %q", projectType)
			}
		})
	}
}
