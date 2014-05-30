alias be='bundle exec'

alias migrate='rake db:migrate db:test:clone'

alias v='vagrant'
alias dbundle='BUNDLE_GEMFILE=.Gemfile bundle'

alias cuke='bundle exec spring cucumber --tags @active'
alias cukes='bundle exec spring cucumber --format progress --require features/'
alias specs='bundle exec spring rspec'

alias server='bundle exec spring rails server'
alias console='bundle exec spring rails console'
alias dbconsole='bundle exec spring rails dbconsole'
