alias r='rbenv local 1.8.7-p358'

alias sc='script/console'
alias sg='script/generate'
alias sd='script/destroy'

alias migrate='rake db:migrate db:test:clone'

alias v='vagrant'

alias cuke='bundle exec spring cucumber --tags @active  --format progress'
alias cukes='bundle exec spring cucumber --format progress --require features'
alias specs='bundle exec spring rspec spec/'
