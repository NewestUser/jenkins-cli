package main

import (
	"flag"
	"fmt"
	"gopkg.in/ini.v1"
	"os"
	"os/exec"
	"strings"
)

const shellToUse = "bash"

const evalNotation = "!"
const evalPrefix = `"` + evalNotation

func loadAliasCmds(conf *ini.File) map[string]command {
	aliasCmds := make(map[string]command)
	for name, value := range loadAliases(conf) {
		aliasName := strings.TrimSpace(name)
		aliasValue := strings.TrimSpace(value)

		if strings.HasPrefix(aliasValue, evalNotation) {
			cmdToEval := strings.Replace(aliasValue, evalNotation, "", 1)
			aliasCmds[aliasName] = shellCmd(aliasName, cmdToEval)
		} else if originalCmd, cmdExists := commands[aliasValue]; cmdExists {
			aliasCmds[aliasName] = originalCmd
		}
	}

	return aliasCmds
}

func shellCmd(cmdName, cmdToEval string) command {
	return command{fs: flag.NewFlagSet(cmdName, flag.ExitOnError), fn: func(globalOpts *jenkinsOpts, args []string) error {
		shellPath, err := exec.LookPath(shellToUse)
		if err != nil {
			return fmt.Errorf("could not find sh installation, please install shell, err: %s", err)
		}

		evalString := strings.TrimSpace(fmt.Sprintf("%s %s", cmdToEval, strings.Join(args, " ")))
		shellCmd := exec.Command(shellPath, "-c", evalString)
		shellCmd.Stdout = os.Stdout
		shellCmd.Stderr = os.Stderr

		if err := shellCmd.Run(); err != nil {
			return fmt.Errorf("failed evaluating alias %s, error: %s", cmdName, err)
		}
		return nil
	}}
}
