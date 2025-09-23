## Settings

{{if .TrustedApps}}
Trusted Apps:
{{range .TrustedApps}}
- {{.Login}}
{{end}}
{{else}}
Trusted Apps: Nothing.
{{end}}

Untrusted Machine Users:

{{if .UntrustedMachineUsers}}
Untrusted Machine Users:
{{range .UntrustedMachineUsers}}
- {{.Login}}
{{end}}
{{else}}
Untrusted Machine Users: Nothing
{{end}}

Trusted Machine Users:

{{if .TrustedMachineUsers}}
Trusted Machine Users:
{{range .TrustedMachineUsers}}
- {{.Login}}
{{end}}
{{else}}
Trusted Machine Users: Nothing
{{end}}
