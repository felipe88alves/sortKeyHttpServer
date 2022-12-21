#!/bin/bash

# ------------- Define functions -------------

function access_kind_mgmt_cluster () {
  printf "Accessing KinD Mgmt Cluster\n"
  kubectl config set-context ${kind_mgmt_cluster_name}
}

function deploy_and_configure_argocd () {
  deploy_argocd
}

function deploy_argocd () {
  argocd_preinstalled=$([[ $(kubectl get all -n argocd -o name | grep -c argocd-server) -ge 5 ]] && echo true || echo false)
  
  printf "Deploying ArgoCD.\n"
  $git_root/bin/3pp/kustomize build $git_root/dev-resources/argocd-service/argocd/overlays/kinD | kubectl apply --filename -

  printf "Waiting for argocd rollout to complete.\n"
  until kubectl --namespace argocd rollout status deployment argocd-server | grep "successfully rolled out"; do : ; done
}

function download_argocd_cli () {
  if [[ ! -a ${argocd_bin} ]]
    then
      printf "Downloading the Argo CD CLI\n"
      curl -sSL -o ${argocd_bin} https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64
      chmod +x ${argocd_bin}
  fi
}

function update_argocd_password_and_login () {
  if ${argocd_preinstalled}
    then
      # TODO: verify if password has alread been modified
      login_to_argocd_cli ${argocd_password}
    else
      local argocd_default_password=$(kubectl --namespace argocd get secret argocd-initial-admin-secret --output jsonpath="{.data.password}" | base64 --decode)
      login_to_argocd_cli ${argocd_default_password}
      ${argocd_bin} account update-password --current-password ${argocd_default_password} --new-password ${argocd_password}
      login_to_argocd_cli ${argocd_password}
      # TODO: Delete secret with password
  fi
}

function login_to_argocd_cli () {
    ${argocd_bin} login --insecure --username ${argocd_username} --password $1 --grpc-web ${argocd_server_nodeport} --plaintext
}


function start_projects_and_applications () {
  printf "Starting projects and applications.\n"
  kubectl apply --filename $git_root/dev-resources/argocd-service/project.yaml
  kubectl apply --filename $git_root/dev-resources/argocd-service/apps.yaml
}

# ------------- End functions -------------

# ------------- Start script -------------

# Define variables
declare -r kind_mgmt_cluster_name="kind-kind-mgmt"
declare -r git_root=$(git rev-parse --show-toplevel)
declare -r argocd_bin="$git_root/bin/3pp/argocd"
declare -r argocd_username="admin"
declare -r argocd_password="admin123"
declare -r argocd_server_nodeport="127.0.0.1:8080"

# init
access_kind_mgmt_cluster
deploy_and_configure_argocd
download_argocd_cli
update_argocd_password_and_login
start_projects_and_applications

# ------------- End script -------------