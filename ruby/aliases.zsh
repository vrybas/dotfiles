alias r='rbenv local 1.8.7-p358'

alias sc='script/console'
alias sg='script/generate'
alias sd='script/destroy'

alias migrate='rake db:migrate db:test:clone'

alias v='vagrant'
alias dbundle='BUNDLE_GEMFILE=.Gemfile bundle'

alias cuke='bundle exec cucumber --tags @active'
alias cukes='bundle exec cucumber --format progress --require features/'
alias specs='bundle exec rspec'
