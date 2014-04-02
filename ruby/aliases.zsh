alias r='rbenv local 1.8.7-p358'

alias sc='script/console'
alias sg='script/generate'
alias sd='script/destroy'

alias migrate='rake db:migrate db:test:clone'

alias v='vagrant'
alias dbundle='BUNDLE_GEMFILE=.Gemfile bundle'

alias cuke='bundle exec spring cucumber --tags @active'
alias cukes='bundle exec spring cucumber --require features/step_definitions/ features/support/'
alias specs='bundle exec spring rspec spec/'
