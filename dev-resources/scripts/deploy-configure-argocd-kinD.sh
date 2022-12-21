#!/bin/bash

# ------------- Define functions -------------

function access_kind_gitops_cluster () {
  printf "Accessing KinD GitOps Cluster\n"
  kubectl config set-context ${kind_gitops_cluster_name}
}

function deploy_and_configure_argocd () {
  deploy_argocd
}

function deploy_argocd () {
  argocd_preinstalled=$([[ $(kubectl get all -n argocd -o name | grep -c argocd-server) -ge 5 ]] && echo true || echo false)
  
  printf "Deploying ArgoCD.\n"
  download_kustomize
  ${kustomize_bin} build $git_root/dev-resources/argocd-service/argocd/overlays/kinD | kubectl apply --filename -

  printf "Waiting for argocd rollout to complete.\n"
  until kubectl --namespace argocd rollout status deployment argocd-server | grep "successfully rolled out"; do : ; done
}

function download_kustomize () {
  if [[ ! -a ${kustomize_bin} ]]
    then
      printf "Downloading Kustomize\n"
      local kustomize_install_script="https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
      local kustomize_version="v4.5.7"
      curl -s ${kustomize_install_script} | bash -s -- ${kustomize_bin}
  fi
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

function install_kyverno () {
  install_helm
  # local kyverno_helm_release=$(${vcluster_bin} list --output json | jq -r '.[].Name')
  local kyverno_helm_release_list=$(${helm_bin} ls -n kyverno | awk -F ' ' '{print $1}')
  if [[ "${kyverno_helm_release_list}" != *"kyverno"* ]]
    then
      ${helm_bin} repo add kyverno https://kyverno.github.io/kyverno/
      ${helm_bin} install kyverno kyverno/kyverno -n kyverno --create-namespace
    else
      printf "Kyverno already Installed\n"
  fi
}

function install_helm () {
  if [[ ! -a ${helm_bin} ]]
    then
      printf "Downloading Helm\n"
      helm_version="v3.10.3"
      wget https://get.helm.sh/helm-${helm_version}-linux-amd64.tar.gz
	    tar xvf helm-${helm_version}-linux-amd64.tar.gz
	    mv linux-amd64/helm helm_bin
	    rm -rf linux-amd64 helm-${helm_version}-linux-amd64.tar.gz
  fi
}

function create_vcluster () {
  download_vcluster
  # local cluster_name_list=$(${vcluster_bin} list --output json | jq -r '.[].Name')
  local cluster_name_list=$(${vcluster_bin} list | awk -F ' ' '{print $1}' | xargs -n 1 printf)
  if [[ "${cluster_name_list}" != *"$vcluster_name"* ]]
    then
      ${vcluster_bin} create ${vcluster_name} --connect=false -f ./dev-resources/infra/vcluster/values.yaml
  fi
}

function download_vcluster () {
  if [[ ! -a ${vcluster_bin} ]]
    then
      printf "Downloading vCluster\n"
      curl -L -o vcluster "https://github.com/loft-sh/vcluster/releases/latest/download/vcluster-linux-amd64"
	    install -c -m 0755 vcluster ${vcluster_bin}
	    rm -f vcluster
  fi
}

function start_projects_and_applications () {
  printf "Starting projects and applications.\n"
  kubectl apply --filename $git_root/dev-resources/argocd-service/project.yaml
  kubectl apply --filename $git_root/dev-resources/argocd-service/apps.yaml
}

# ------------- End functions -------------

# ------------- Start script -------------

# Define variables
declare -r kind_gitops_cluster_name="kind-kind-gitops"
declare -r git_root=$(git rev-parse --show-toplevel)
declare -r argocd_bin="$git_root/bin/3pp/argocd"
declare -r argocd_username="admin"
declare -r argocd_password="admin123"
declare -r argocd_server_nodeport="127.0.0.1:8080"
declare -r kustomize_bin="$git_root/bin/3pp/kustomize"
declare -r vcluster_bin="$git_root/bin/3pp/vcluster"
declare -r vcluster_name="vc1"
declare -r helm_bin="$git_root/bin/3pp/helm"


# init
access_kind_gitops_cluster
deploy_and_configure_argocd
download_argocd_cli
update_argocd_password_and_login
install_kyverno
create_vcluster
start_projects_and_applications

# ------------- End script -------------