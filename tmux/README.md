## My Tmux configuration ##
```
___________
\__    ___/____  __ _____  ___
  |    | /     \|  |  \  \/  /
  |    ||  Y Y  \  |  />    <
  |____||__|_|  /____//__/\_ \
              \/            \/
```

## Warning ##

If you're not on OSX, you should comment out the line below, otherwise tmux won't start:

`set -g default-command "reattach-to-user-namespace -l zsh"`


If you're on OSX, install this script via Homebrew:

`> brew install reattach-to-user-namespace`

## Keymap (some standard mappings changed) ##

* `Ctrl-space` -  PREFIX

### Tabs (aka "windows")

PREFIX + ...
* `c` - create tab
* `,` - rename tab
* `h/l` - jump to letf/right tab
* `1..9` - select tab by number
* `&` - close tab

### Splits (aka "panes")

PREFIX + ...
* `v` - split window vertically
* `s` - split window horizontally
* `H/J/K/L` - resize split
* `o` - toggle maximize split
* `x` - close split

no prefix
* `Ctrl-h/j/k/l` - jump between splits left/down/up/right

### Copy-mode

PREFIX + ...
* `Esc` - enter copy-mode

inside copy-mode
* `h/j/k/l` - move left/down/up/right
* `v` - begin selection
* `y` - copy selection

PREFIX + ...
* `p` - paste copied selection

### Other
* `q` - switch between sessions
* `r` - reload configuration

