#!/bin/bash

set -euo pipefail
source common.sh

function migrate {
  local asg=$(gen-asg-name $1)
  echo "==> Migrating workloads on ASG: $asg"

  echo "  ==> Draining outdated node"
  for i in $(get-outdated-instances $asg)
  do
    read -p "  ==> Press any key to continue"
    kubectl drain --force --ignore-daemonsets --delete-local-data $i
  done
}

migrate $@
