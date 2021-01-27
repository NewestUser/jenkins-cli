package main

import (
	b64 "encoding/base64"
	"flag"
	"fmt"
	"gopkg.in/ini.v1"

	"log"
	"os"
)

const JenkinsConfig = ".jenkins"

var JenkinsConfigPath string

const JenkinsSection = "jenkins"
const AliasSection = "alias"

func init() {
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("cannot acquire access to home directory")
	}

	JenkinsConfigPath = dir + "/" + JenkinsConfig
}

var configSchema = map[string]string{
	"host":  JenkinsSection,
	"user":  JenkinsSection,
	"token": JenkinsSection,
}

func configCmd() command {
	fs := flag.NewFlagSet(fmt.Sprintf("%s config", CliName), flag.ExitOnError)

	opts := &configOpts{
		saveAction: false,
	}

	fs.BoolVar(&opts.saveAction, "save", false, "Save configuration in the format key value")

	return command{fs: fs, fn: func(globalOpts *jenkinsOpts, args []string) error {
		fs.Parse(args)

		if !opts.saveAction || len(fs.Args()) != 2 {
			return fmt.Errorf("invalid number of arguments: %s, a valid example would be: %s config -save foo bar", fs.Args(), CliName)
		}

		prop, err := newProperty(fs.Args()[0], fs.Args()[1])
		if err != nil {
			return err
		}

		return saveProperty(prop)
	}}
}

func saveProperty(prop *configProp) error {
	iniConfig, err := loadJenkinsConfig()
	if err != nil {
		return err
	}

	propVal, err := processBeforeWrite(prop)
	if err != nil {
		return fmt.Errorf("failed processing property %s with value %s, err: %s", prop.name, prop.value, err)
	}

	iniConfig.Section(prop.section).Key(prop.name).SetValue(propVal)

	if err = iniConfig.SaveTo(JenkinsConfigPath); err != nil {
		return fmt.Errorf("failed saving property %s=%s to %s", prop.name, prop.value, JenkinsConfig)
	}

	return nil
}

func loadJenkinsConfig() (*ini.File, error) {
	if err := TouchFile(JenkinsConfigPath); err != nil {
		return nil, fmt.Errorf("failed creating %s, err: %s", JenkinsConfig, err)
	}

	iniConfig, err := ini.Load(JenkinsConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed loading %s, err: %s", JenkinsConfig, err)
	}

	return iniConfig, nil
}

func newProperty(name, value string) (*configProp, error) {
	if section, exists := configSchema[name]; exists {
		return &configProp{section: section, name: name, value: value}, nil
	}

	return nil, fmt.Errorf("unkown property %s", name)
}

func loadHostConfig(file *ini.File) string {
	return file.Section(JenkinsSection).Key("host").Value()
}

func loadTokenConfig(file *ini.File) (string, error) {
	encodedValue := file.Section(JenkinsSection).Key("token").Value()
	token, err := base64Decode(encodedValue)
	if err != nil {
		return "", fmt.Errorf("could not base64 decode token property")
	}
	return token, nil
}

func loadUserConfig(file *ini.File) string {
	return file.Section(JenkinsSection).Key("user").Value()
}

func loadAliases(file *ini.File) map[string]string {
	aliases := make(map[string]string)
	for _, key := range file.Section(AliasSection).Keys() {
		aliases[key.Name()] = key.Value()
	}
	return aliases
}

type configOpts struct {
	saveAction bool
}

type configProp struct {
	section string
	name    string
	value   string
}

func processBeforeWrite(prop *configProp) (string, error) {
	if prop.name == "token" {
		return base64Encode(prop.value)
	}
	return prop.value, nil
}

func base64Decode(value string) (string, error) {
	bytes, err := b64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func base64Encode(value string) (string, error) {
	return b64.StdEncoding.EncodeToString([]byte(value)), nil
}
