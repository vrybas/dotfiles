tell application "Tomato" to activate
tell application "System Events"
  tell process "Tomato"
    key down space
    key up space
    key down command
    key down tab
    key up tab
    key up command
  end tell
end tell
