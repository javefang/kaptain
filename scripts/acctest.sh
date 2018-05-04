#!/bin/bash

set -euxo pipefail

CLUSTER_NAME=dev.acctest.aws
TESTDIR=/tmp/kaptain_test

rm -rf /tmp/kaptain_test*

kaptain create -n $CLUSTER_NAME

mkdir -p $TESTDIR/etcd/etc/pki/tls/certs
mkdir -p $TESTDIR/etcd/etc/pki/tls/private
mkdir -p $TESTDIR/master/etc/{sysconfig,docker}
mkdir -p $TESTDIR/master/etc/kubernetes/manifests
mkdir -p $TESTDIR/master/var/lib/{kubelet,kube-proxy,kubernetes}
mkdir -p $TESTDIR/worker/etc/{sysconfig,docker}
mkdir -p $TESTDIR/worker/var/lib/{kubelet,kube-proxy}

sailor provision --name $CLUSTER_NAME --prefix="$TESTDIR/etcd" --role=etcd
sailor provision --name $CLUSTER_NAME --prefix="$TESTDIR/master" --role=master
sailor provision --name $CLUSTER_NAME --prefix="$TESTDIR/worker" --role=worker

kaptain delete -n $CLUSTER_NAME

# compare generated files
cat << 'EOF' > /tmp/kaptain_test_expected
/tmp/kaptain_test/etcd/etc/pki/tls/certs/etcd-ca.pem
/tmp/kaptain_test/etcd/etc/pki/tls/certs/etcd-server.pem
/tmp/kaptain_test/etcd/etc/pki/tls/private/etcd-server-key.pem
/tmp/kaptain_test/master/etc/docker/daemon.json
/tmp/kaptain_test/master/etc/kubernetes/manifests/kube-apiserver.yaml
/tmp/kaptain_test/master/etc/kubernetes/manifests/kube-controller-manager.yaml
/tmp/kaptain_test/master/etc/kubernetes/manifests/kube-scheduler.yaml
/tmp/kaptain_test/master/etc/sysconfig/docker
/tmp/kaptain_test/master/etc/sysconfig/kube-proxy-kaptain
/tmp/kaptain_test/master/etc/sysconfig/kubelet-kaptain
/tmp/kaptain_test/master/etc/sysconfig/kubelet-kaptain-extra
/tmp/kaptain_test/master/var/lib/kube-proxy/kubeconfig
/tmp/kaptain_test/master/var/lib/kubelet/kubeconfig
/tmp/kaptain_test/master/var/lib/kubernetes/ca-key.pem
/tmp/kaptain_test/master/var/lib/kubernetes/ca.pem
/tmp/kaptain_test/master/var/lib/kubernetes/etcd-ca.pem
/tmp/kaptain_test/master/var/lib/kubernetes/etcd-client-key.pem
/tmp/kaptain_test/master/var/lib/kubernetes/etcd-client.pem
/tmp/kaptain_test/master/var/lib/kubernetes/kube-controller-manager.kubeconfig
/tmp/kaptain_test/master/var/lib/kubernetes/kube-scheduler.kubeconfig
/tmp/kaptain_test/master/var/lib/kubernetes/kubernetes-key.pem
/tmp/kaptain_test/master/var/lib/kubernetes/kubernetes.pem
/tmp/kaptain_test/master/var/lib/kubernetes/token.csv
/tmp/kaptain_test/worker/etc/docker/daemon.json
/tmp/kaptain_test/worker/etc/sysconfig/docker
/tmp/kaptain_test/worker/etc/sysconfig/kube-proxy-kaptain
/tmp/kaptain_test/worker/etc/sysconfig/kubelet-kaptain
/tmp/kaptain_test/worker/etc/sysconfig/kubelet-kaptain-extra
/tmp/kaptain_test/worker/var/lib/kube-proxy/kubeconfig
/tmp/kaptain_test/worker/var/lib/kubelet/bootstrap.kubeconfig
EOF

find $TESTDIR -type f > /tmp/kaptain_test_actual

diff /tmp/kaptain_test_expected /tmp/kaptain_test_actual
