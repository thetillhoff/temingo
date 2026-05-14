package temingo

import (
	"testing"

	"github.com/thetillhoff/fileIO"
)

func TestValidateFolderNames(t *testing.T) {
	tests := []struct {
		name    string
		files   []string
		wantErr bool
	}{
		{
			name:    "no files",
			files:   []string{},
			wantErr: false,
		},
		{
			name:    "files at root only",
			files:   []string{"index.html", "style.css"},
			wantErr: false,
		},
		{
			name:    "valid nested paths",
			files:   []string{"blog/post.html", "assets/images/logo.png"},
			wantErr: false,
		},
		{
			name:    "folder with space",
			files:   []string{"my folder/index.html"},
			wantErr: true,
		},
		{
			name:    "folder with hash",
			files:   []string{"section#1/page.html"},
			wantErr: true,
		},
		{
			name:    "folder with question mark",
			files:   []string{"search?q/index.html"},
			wantErr: true,
		},
		{
			name:    "folder with ampersand",
			files:   []string{"foo&bar/index.html"},
			wantErr: true,
		},
		{
			name:    "folder with angle brackets",
			files:   []string{"<tag>/index.html"},
			wantErr: true,
		},
		{
			name:    "deeply nested valid path",
			files:   []string{"a/b/c/d/index.html"},
			wantErr: false,
		},
		{
			name:    "deeply nested invalid path",
			files:   []string{"a/b/bad folder/d/index.html"},
			wantErr: true,
		},
		{
			name:    "valid filename with space is not checked",
			files:   []string{"blog/my file.html"},
			wantErr: false,
		},
		{
			name:    "hyphens and underscores are valid",
			files:   []string{"my-section/my_page.html"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fl := fileIO.FileList{Files: tt.files}
			err := validateFolderNames(fl)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFolderNames() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
