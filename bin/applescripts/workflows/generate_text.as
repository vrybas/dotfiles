on run {input, parameters}
  set userResponse to display dialog "Function?" default answer "" with icon note buttons {"Cancel", "Continue"} default button "Continue"
  set funcName to text returned of userResponse

  set funcResult to null

  if funcName equal to "utc"
    set funcResult to do shell script "ruby -e \"puts '!' << Time.now.utc.to_s\""
  end if

  if funcName equal to "utcd"
    set funcResult to do shell script "ruby -e \"puts '!' << Time.now.utc.strftime('%Y-%m-%d')\""
  end if

  if funcName equal to "utcn"
    set funcResult to do shell script "ruby -e \"puts Time.now.utc.strftime('%Y%m%d%H%M%S')\""
  end if

  if funcName equal to "hash"
    set funcResult to do shell script "ruby -e \"require 'securerandom'; puts '**' << SecureRandom.hex[0..6]\""
  end if

  if funcResult is not null
    tell application "System Events"
      set the clipboard to funcResult
    end tell
  end if
end run
