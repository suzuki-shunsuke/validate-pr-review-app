## Settings

{{if .TrustedApps}}
Trusted Apps:
{{range .TrustedApps}}
- {{.Login}}
{{end}}
{{else}}
Trusted Apps: Nothing.
{{end}}

{{if .UntrustedMachineUsers}}
Untrusted Machine Users:
{{range .UntrustedMachineUsers}}
- {{.Login}}
{{end}}
{{else}}
Untrusted Machine Users: Nothing
{{end}}

{{if .TrustedMachineUsers}}
Trusted Machine Users:
{{range .TrustedMachineUsers}}
- {{.Login}}
{{end}}
{{else}}
Trusted Machine Users: Nothing
{{end}}
