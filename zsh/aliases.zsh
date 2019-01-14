alias reload!='. ~/.zshrc'
alias l="clear"
alias x="exit"
alias i="imgur2"
alias pwc="pwd | pbcopy; echo 'Current path copied to clipboard! $(pwd)'"

# Docker
alias dk="docker"

# Kubernetes
alias kb="kubectl"
alias kp="kops"
alias mk="minikube"
# Current Status file
alias tstat="tail -f ~/current_status.log"
alias estat="nvim ~/current_status.log"

# Central Kubernetes
alias kbc="kb --kubeconfig ~/Projects/CrossOver/Aurea/kube/config"
## Placeable
alias kbc-pl-dev="kbc --namespace placeable-dev"
alias kbc-pl-qa="kbc --namespace placeable-qa"
alias kbc-pl-ua="kbc --namespace placeable-ua"
alias kbc-pl-prod="kbc --namespace placeable-prod"
## Chute
alias kbc-ch-dev="kbc --namespace chute-dev"
alias kbc-ch-qa="kbc --namespace chute-qa"
alias kbc-ch-staging="kbc --namespace chute-staging"

# Digital Ocean Kubernetes
alias kbdo="kb --kubeconfig ~/do-kubernetes-cluster/v3-20180930192816/admin.conf"
alias kbdo-sys="kbdo --namespace kube-system"
alias kbdo-ovpn="kbdo --namespace openvpn"

# Digital Ocean Helm
alias hdo="KUBECONFIG=~/do-kubernetes-cluster/v3-20180930192816/admin.conf helm"

# Google Kubernetes Engine
alias kbg="kb --kubeconfig ~/gcp-kubernetes-cluster/v1-20180520111558/kube/config"
