#!/bin/bash
set -euo pipefail

PLAN=the_rolling_plan

function mkPrefix {
  local module=$1
  local env=$2
  local tag=$3
  
  echo ${env}-kube${module}-${tag}
}

function mkModule {
  if [ "$1" == "master" ] || [ "$1" == "worker" ] || [ "$1" == "ingress" ]
  then
    echo kube-$1
  else
    echo kube-worker-$1
  fi
}

function taint {
  local module=$(mkModule $1)
  terraform taint -module=$module vsphere_virtual_machine.kube.$2
}

function plan {
  local module=$(mkModule $1)
  terraform plan -out $PLAN \
    -target=module.$module.vsphere_virtual_machine.kube[$2] \
    -target=module.$module.dns_a_record_set.kube[$2] \
  || failGuard
}

function apply {
  terraform apply $PLAN || failGuard
}

function wait {
  echo "==========================="
  echo ""
  read -p "Press enter to continue"
}

function waitDelay {
  echo "==> Commencing next step in ${1}s"
  sleep $1
}

function waitTilReady {
  local startTime=$(date +%s)
  until kubectl get node $1 -o json | jq -re '.status.conditions[] | select(.type=="Ready" and .status=="True") | .status'
  do
    local currentTime=$(date +%s)
    echo "Waiting for $1 to become ready... ($(($currentTime - $startTime))s elapsed)"
    sleep 10
  done
}

function failGuard {
  read -p "Action failed, hit enter to ignore, or ctrl-c to quit"
}

function cordon {
  kubectl cordon $1 || failGuard
}

function uncordon {
  kubectl uncordon $1 || failGuard
}

function drain {
  kubectl drain $1 --ignore-daemonsets --delete-local-data --force || failGuard
}

function roll {
  local env=$1
  local tag=$2
  local module=$3
  local size=$4

  local hostnamePrefix=$(mkPrefix $module $env $tag)

  echo "==> Cordon all nodes..."

  for i in $(seq 0 $(($size-1)))
  do
    cordon $hostnamePrefix-$i
  done

  echo "==> Rolling module $module with $size nodes..."

  for i in $(seq 0 $(($size-1)))
  do
    echo "====> Update node $hostnamePrefix-$i ($(($i+1))/$size)? [Y|n]"
    read updateOpt
    case $updateOpt in
    n)
      echo "Skipping..."
      continue
      ;;
    *)
      echo "Updating..."
      ;;
    esac

    echo "====> Updating node $hostnamePrefix-$i ($(($i+1))/$size)"
    drain $hostnamePrefix-$i
    echo "====> Verify pods all drained and VMDKs are detached from VM" 
    wait
    taint $module $i
    plan $module $i
    apply
    waitTilReady $hostnamePrefix-$i
    uncordon $hostnamePrefix-$i
    echo "===> Finished updating node $hostnamePrefix-$i ($(($i+1))/$size)"
  done
}

roll $@
