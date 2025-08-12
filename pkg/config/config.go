package config

type Config struct {
	AppID                 int64               `yaml:"app_id"`
	InstallationID        int64               `yaml:"installation_id"`
	TrustedApps           map[string]struct{} `yaml:"trusted_apps"`
	UntrustedMachineUsers map[string]struct{} `yaml:"untrusted_machine_users"`
	AWS                   *AWS                `yaml:"aws"`
	CheckName             string              `yaml:"check_name"`
}

type AWS struct {
	SecretID string `yaml:"secret_id"`
}
