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
{{- if or .AllowUnsignedCommits .UnsignedCommitAuthors}}
:warning: Insecure Settings:
{{- if .AllowUnsignedCommits}}
- Allow Unsigned Commits: Yes
{{- end}}
{{- if .UnsignedCommitAuthors}}
- Unsigned Commit Authors:
{{- range .UnsignedCommitAuthors}}
  - {{.}}
{{- end}}
{{- end}}
{{end -}}
