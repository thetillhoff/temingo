package temingo

import "testing"

func TestValidateEngine(t *testing.T) {
	tests := []struct {
		name    string
		engine  Engine
		wantErr bool
	}{
		{
			name:    "valid default config",
			engine:  DefaultEngine(),
			wantErr: false,
		},
		{
			name: "beautify and minify both enabled",
			engine: func() Engine {
				e := DefaultEngine()
				e.Beautify = true
				e.Minify = true
				return e
			}(),
			wantErr: true,
		},
		{
			name: "template and metatemplate extension equal",
			engine: func() Engine {
				e := DefaultEngine()
				e.TemplateExtension = ".tmpl"
				e.MetaTemplateExtension = ".tmpl"
				return e
			}(),
			wantErr: true,
		},
		{
			name: "template and partial extension equal",
			engine: func() Engine {
				e := DefaultEngine()
				e.TemplateExtension = ".tmpl"
				e.PartialExtension = ".tmpl"
				return e
			}(),
			wantErr: true,
		},
		{
			name: "metatemplate and partial extension equal",
			engine: func() Engine {
				e := DefaultEngine()
				e.MetaTemplateExtension = ".tmpl"
				e.PartialExtension = ".tmpl"
				return e
			}(),
			wantErr: true,
		},
		{
			name: "all extensions distinct",
			engine: func() Engine {
				e := DefaultEngine()
				e.TemplateExtension = ".template"
				e.MetaTemplateExtension = ".metatemplate"
				e.PartialExtension = ".partial"
				return e
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.engine.validateEngine()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateEngine() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
