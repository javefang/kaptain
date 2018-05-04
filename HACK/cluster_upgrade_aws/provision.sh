#!/bin/bash

set -euo pipefail
source common.sh

function provision {
  local asg=$(gen-asg-name $1)
  echo "==> Provisoning ASG: $asg"

  local asg_size=$(get-asg-size $asg)
  local target_asg_size=$2

  echo "  ==> Cordoning all outdated nodes"
  get-outdated-instances $asg | xargs -I{} kubectl cordon {}

  echo "  ==> Updating ASG size: $asg_size -> $target_asg_size"
  set-asg-size $asg $target_asg_size

  wait-asg-until-in-service $asg $target_asg_size
}

provision $@
