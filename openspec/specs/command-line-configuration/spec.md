## Purpose

Configure the Penda CLI target directory and non-interactive interface.

## Requirements

### Requirement: Target directory selection
The Penda CLI SHALL accept a target directory through `--dir` and its `-d` shorthand. When neither option is supplied, it SHALL use the process current working directory as the target directory.

#### Scenario: Explicit directory is supplied
- **WHEN** a user invokes Penda with `--dir /workspace/repository` or `-d /workspace/repository`
- **THEN** Penda SHALL use `/workspace/repository` as its target directory

#### Scenario: Directory is omitted
- **WHEN** a user invokes Penda without `--dir` or `-d`
- **THEN** Penda SHALL use the process current working directory as its target directory

### Requirement: Target directory validation
Before beginning an operation, Penda SHALL verify that its target directory exists and is a directory. It SHALL terminate with a non-zero exit status and a human-readable error when the target is missing or is not a directory.

#### Scenario: Target directory exists
- **WHEN** Penda is invoked with a path that exists and is a directory
- **THEN** Penda SHALL continue with that path as the target directory

#### Scenario: Target directory is invalid
- **WHEN** Penda is invoked with a path that does not exist or is not a directory
- **THEN** Penda SHALL not begin the operation and SHALL exit with a non-zero status

### Requirement: Non-interactive command interface
Penda SHALL parse command-line options with `alecthomas/kong` and SHALL complete without requiring interactive terminal input.

#### Scenario: CI invocation
- **WHEN** Penda is invoked in an environment without an interactive terminal
- **THEN** Penda SHALL parse its supplied options and run or fail using standard process exit status
