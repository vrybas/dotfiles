on run
  set userResponse to display dialog "Current status" default answer "" with icon note buttons {"Cancel", "Continue"} default button "Continue"
  set statusText to text returned of userResponse

  if statusText is not equal to ""
    do shell script "~/.dotfiles/bin/log_current_status \"" & statusText & "\""
  end if
end run
