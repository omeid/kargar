package kargar

import (
	"text/template"
)

// BuildHelpTemplate holds the template used for generating
// the help message for a build
// TODO: Move this to /kar?
var BuildHelpTemplate = `KARGAR:
   {{.Name}} - {{.Usage}}

USAGE:
   {{.Name}} [global options]  {task [flags]...}

VERSION:
   {{.Version}}{{if or .Author .Email}}

AUTHOR:{{if .Author}}
  {{.Author}}{{if .Email}} - <{{.Email}}>{{end}}{{else}}
  {{.Email}}{{end}}{{end}}

TASKS:
   {{range .Tasks }}{{printf "%-15s %s" .Name .Usage}}
   {{end}}{{if .Flags}}
GLOBAL OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{end}}
`

// TaskTemplate holds the template used to generate
// the help message for a specific task.
var TaskHelpTemplate = `TASK:
   {{.Name}} - {{.Usage}}{{if .Description}}

DESCRIPTION:
   {{.Description}}{{end}}{{if .Deps}}

DEPENDENCIES:
   {{ range .Deps }}{{ . }}
   {{ end }}{{ end }}{{ if .Flags }}

OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{ end }}
`

//Help templates hold both Build and Task help templates.
var HelpTemplate *template.Template

func init() {
	HelpTemplate = template.Must(template.New("build").Parse(BuildHelpTemplate))
	HelpTemplate = template.Must(HelpTemplate.New("task").Parse(TaskHelpTemplate))
}
