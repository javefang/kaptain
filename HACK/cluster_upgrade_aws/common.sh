function check-context {
  echo "========== ATTENTION ============"
  echo "Target cluster: $(kubectl config current-context)"
  echo ""
  local input="n"
  read -p "Continue? (y/n) " input
  
  if [ $input != "y" ]; then
    echo "Cancelled"
    exit 1
  fi
}

function wait {
  echo "==========================="
  echo ""
  read -p "Press enter to continue"
}

function trim {
  cat - | sed -e 's/^[ \t]*//'
}

function gen-asg-name {
  echo -n "kube-node-$1-asg-$AWS_VPC_NAME"
}

function get-asg-size {
  aws autoscaling describe-auto-scaling-groups --auto-scaling-group-names=$1 | jq -r '.AutoScalingGroups[0].DesiredCapacity'
}

function get-outdated-instances {
  local asg=$(aws autoscaling describe-auto-scaling-groups --auto-scaling-group-names=$1 | jq -r '.AutoScalingGroups[0]')

  local asg_lc=$(echo -n $asg | jq -r '.LaunchConfigurationName')  
  local outdated_instance_ids=$(echo -n $asg | jq -r --arg LC "$asg_lc" '.Instances[] | select(.LaunchConfigurationName!=$LC) | .InstanceId' | tr '\n' ' ')
  if [ $(echo $outdated_instance_ids | wc -w) -gt 0 ]
  then
    aws ec2 describe-instances --instance-ids $outdated_instance_ids --query 'Reservations[*].Instances[*].[PrivateDnsName]' | jq -r 'flatten[]'
  fi
}

function set-asg-size {
  aws autoscaling update-auto-scaling-group --auto-scaling-group-name=$1 --desired-capacity=$2
}

function get-asg-size-in-service {
  aws autoscaling describe-auto-scaling-groups --auto-scaling-group-names=$1 | jq -r '.AutoScalingGroups[0].Instances[] | select(.LifecycleState=="InService") | .InstanceId' | wc -l | trim
}

function wait-asg-until-in-service {
  local in_service=0
  until [ $in_service -eq $2 ]; do
    in_service=$(get-asg-size-in-service $1)
    echo "  ==> Waiting for instances to bootstrap... ($in_service/$2)"
    sleep 5
  done
}

