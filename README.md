# jenkins is a Jenkins Command Line Client written in Go

### Setup

```shell
jenkins config -save host https://my-jenkins.com
jenkins config -save user superdev
jenkins config -save token aabbcc
```

### List a job with wildcard
```shell
jenkings jobs My-Job-*
```

### Start a job
```shell
jenkins start -P branch:awsome-feature deploy-job
```

### Usage

```shell
Usage: jenkins [global flags] <command> [command flags]

global flags:
  -h string
        Host of the jenkins server
  -t string
        Api token to be used for authentication
  -u string
        User that the api token belongs to
  -version
        Print version and exit

config command:
  -save
        Save configuration in the format key value

jobs command:
  -d    Show job details
  -i    Ignore case (default true)

start command:
  -P value
        Job parameter in the format key:value
```


### Example .jenkins file

```text
[jenkins]
user  = username
token = base64EncodedToken
host  = myHost

[alias]
jb     = jobs
myjobs = "!jenkins jobs -d My-Jobs-*"
deploy = "!jenkins start -P branch:$(git branch --show-current) deploy-job"
```
