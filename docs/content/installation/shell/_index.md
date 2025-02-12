---
title: "Shell Completion"
date: 2022-04-23T23:55:10+01:00
weight: 100
---


Shell command line completions are provided for `bash` and `zsh`. 

To load the command completions in shell, use:

```shell
# bash
eval "$(resticprofile generate --bash-completion)"

# zsh
eval "$(resticprofile generate --zsh-completion)"
```

To install them permanently:

```
$ resticprofile generate --bash-completion > /etc/bash_completion.d/resticprofile
$ chmod +x /etc/bash_completion.d/resticprofile
```
