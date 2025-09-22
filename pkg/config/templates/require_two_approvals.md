# :x: Two approvals are required

This pull request requires two approvals because:

{{if .SelfApprovals}}
Someone who contributed commits to the PR approves the latest commit of the PR:

{{range .SelfApprovals}}
- {{.Login}}
{{end}}
{{end}}

{{if .UntrustedMachineUsers}}
Untrusted Machine Users committed:
{{range .UntrustedMachineUsers}}
- {{.Login}}
{{end}}
{{end}}

{{if .UntrustedApps}}
Untrusted GitHub Apps committed:
{{range .UntrustedApps}}
- {{.Login}}
{{end}}
{{end}}

{{if not .Author.Trusted}}
{{if .Author.IsApp}}
The pull request author is an untrusted GitHub App `{{.Author.Login}}`.
{{else}}
The pull request author is an untrusted Machine User `{{.Author.Login}}`.
{{end}}
{{end}}

{{if .UnknownCommits}}
Some commits aren't linked to any GitHub Users.
{{range .UnknownCommits}}
- {{.Commit}}
{{end}}
{{end}}

{{if .IgnoredApprovals}}
## :warning: Some approvals are ignored

Approvals from GitHub Apps and Untrusted Machine Users are ignored.

Approvals from the following approvers are ignored:
{{range .IgnoredApprovals}}
- {{.Login}} {{if .IsApp}}GitHub App{{else}}Untrusted Machine User{{end}}
{{end}}
{{end}}

{{template "settings" .}}

{{template "footer"}}
