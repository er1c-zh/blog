# A 3-master 3-slave redis cluster

FROM centos:7

RUN yum -y update \
    && yum -y install epel-release \
    && yum install -y http://rpms.famillecollet.com/enterprise/remi-release-7.rpm \
    && yum -y --enablerepo=remi install redis \
    && mkdir cfgs \
    && cd cfgs \
    && redis-server --daemonize yes --port 7000 --cluster-enabled yes --cluster-config-file nodes0.conf --cluster-node-timeout 5000 --appendonly yes \
    && redis-server --daemonize yes --port 7001 --cluster-enabled yes --cluster-config-file nodes1.conf --cluster-node-timeout 5000 --appendonly yes \
    && redis-server --daemonize yes --port 7002 --cluster-enabled yes --cluster-config-file nodes2.conf --cluster-node-timeout 5000 --appendonly yes \
    && redis-server --daemonize yes --port 7003 --cluster-enabled yes --cluster-config-file nodes3.conf --cluster-node-timeout 5000 --appendonly yes \
    && redis-server --daemonize yes --port 7004 --cluster-enabled yes --cluster-config-file nodes4.conf --cluster-node-timeout 5000 --appendonly yes \
    && redis-server --daemonize yes --port 7005 --cluster-enabled yes --cluster-config-file nodes5.conf --cluster-node-timeout 5000 --appendonly yes \
    && redis-cli --cluster create 127.0.0.1:7000 127.0.0.1:7001 \
        127.0.0.1:7002 127.0.0.1:7003 127.0.0.1:7004 127.0.0.1:7005 \
        --cluster-replicas 1 \
    && top

EXPOSE 7000 7001 7002 7003 7004 7005
