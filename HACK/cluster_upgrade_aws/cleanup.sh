#!/bin/bash

set -euo pipefail
source common.sh

function cleanup {
  local asg=$(gen-asg-name $1)
  echo "==> Cleaning up ASG: $asg"

  local asg_size=$(get-asg-size $asg)
  local target_asg_size=$2

  echo "  ==> Updating ASG size: $asg_size -> $target_asg_size"
  set-asg-size $asg $target_asg_size

  wait-asg-until-in-service $asg $target_asg_size

  echo "  ==> All instances in $asg are up-to-date now"
  echo "  ==> Removed nodes should disappear from kubectl in 40s (--node-monitor-grace-period)"
}

cleanup $@
