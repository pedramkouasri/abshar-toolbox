package types

type CommandType string;

const(
	DockerCommandType CommandType = "docker"
	ShellCommandType CommandType = "shell"
)

type ComposerCommand struct {
	Type CommandType `yaml:type`
	Container string `yaml:container`
	Cmd     string `yaml:cmd`
}