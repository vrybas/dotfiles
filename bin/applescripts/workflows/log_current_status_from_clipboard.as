on run
  tell application "System Events"
    key code 123 using {command down, shift down}
    key code 124 using {command down, shift down}
    keystroke "c" using {command down}
    key code 124
    delay 1

    set statusText to (the clipboard as text)

    if statusText is not equal to ""
      do shell script "~/.dotfiles/bin/log_current_status \"" & statusText & "\""
    end if
  end tell
end run
