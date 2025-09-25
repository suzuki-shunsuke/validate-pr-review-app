## Settings
{{if .TrustedApps}}
Trusted Apps:
{{range .TrustedApps}}
- {{. -}}
{{end}}
{{else}}
Trusted Apps: Nothing
{{- end}}
{{if .UntrustedMachineUsers -}}
Untrusted Machine Users:
{{- range .UntrustedMachineUsers}}
- {{. -}}
{{end}}
{{else}}
Untrusted Machine Users: Nothing
{{end}}

{{- if .TrustedMachineUsers}}
Trusted Machine Users:
{{range .TrustedMachineUsers}}
- {{. -}}
{{end}}
{{else}}
Trusted Machine Users: Nothing
{{end -}}
