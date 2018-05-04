#!/bin/bash

set -euo pipefail
source common.sh

function status {
  local asg=$(gen-asg-name $1)
  echo "==> ASG: $asg"
  
  local asg_size=$(get-asg-size $asg)
  local asg_in_service_size=$(get-asg-size-in-service $asg)
  local asg_outdated_size=$(get-outdated-instances $asg | wc -l | trim)

  echo -e "  ==> Desired:\t\t$asg_size"
  echo -e "  ==> In Service:\t$asg_in_service_size"
  echo -e "  ==> Outdated:\t\t$asg_outdated_size"
}

status $@
