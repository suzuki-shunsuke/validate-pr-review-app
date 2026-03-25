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

{{- if or .AllowUnsignedCommits .UnsignedCommitApps .UnsignedCommitMachineUsers}}
:warning: Insecure Settings:
{{- if .AllowUnsignedCommits}}
- Allow Unsigned Commits: Yes
{{- end}}
{{- if .UnsignedCommitApps}}
- Unsigned Commit Apps:
{{- range .UnsignedCommitApps}}
  - {{.}}
{{- end}}
{{- end}}
{{- if .UnsignedCommitMachineUsers}}
- Unsigned Commit Machine Users:
{{- range .UnsignedCommitMachineUsers}}
  - {{.}}
{{- end}}
{{- end}}
{{end -}}
