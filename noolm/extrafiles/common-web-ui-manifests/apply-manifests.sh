#!/bin/bash
set -x
NS=${1:-ibm-common-services}
#NS=${1:-ssg}

kubectl apply -f operators.ibm.com_commonwebuis.yaml
kubectl -n $NS apply -f service_account.yaml
kubectl -n $NS apply -f role.yaml
kubectl -n $NS apply -f role_binding.yaml
kubectl -n $NS apply -f manager.yaml
# leader_election_role.yaml  operators.ibm.com_v1alpha1_commonwebui_cr.yaml leader_election_role_binding.yaml operators.ibm.com_switcheritems.yaml