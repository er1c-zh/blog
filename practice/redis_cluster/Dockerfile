# A 3-master 3-slave redis cluster

FROM ubuntu:18.04

COPY ./create_cluster.exp ./create_cluster.exp

# 解决apt安装时需要输入的情况
ENV DEBIAN_FRONTEND=noninteractive

RUN apt update \
    && apt -y install wget build-essential expect \
    && wget https://download.redis.io/releases/redis-6.0.12.tar.gz \
    && tar -zvxf ./redis-6.0.12.tar.gz \
    && rm ./redis-6.0.12.tar.gz \
    && mv redis-6.0.12 redis \
    && cd redis \
    && make \
    && cd ..

EXPOSE 7000 7001 7002 7003 7004 7005

CMD    ./redis/src/redis-server --protected-mode no --daemonize yes --port 7000 --cluster-enabled yes --cluster-config-file nodes0.conf --cluster-node-timeout 5000 --appendonly yes \
    && ./redis/src/redis-server --protected-mode no --daemonize yes --port 7001 --cluster-enabled yes --cluster-config-file nodes1.conf --cluster-node-timeout 5000 --appendonly yes \
    && ./redis/src/redis-server --protected-mode no --daemonize yes --port 7002 --cluster-enabled yes --cluster-config-file nodes2.conf --cluster-node-timeout 5000 --appendonly yes \
    && ./redis/src/redis-server --protected-mode no --daemonize yes --port 7003 --cluster-enabled yes --cluster-config-file nodes3.conf --cluster-node-timeout 5000 --appendonly yes \
    && ./redis/src/redis-server --protected-mode no --daemonize yes --port 7004 --cluster-enabled yes --cluster-config-file nodes4.conf --cluster-node-timeout 5000 --appendonly yes \
    && ./redis/src/redis-server --protected-mode no --daemonize yes --port 7005 --cluster-enabled yes --cluster-config-file nodes5.conf --cluster-node-timeout 5000 --appendonly yes \
    # && chmod +x ./create_cluster.exp \
    # && ./create_cluster.exp \
    # && ./redis/src/redis-cli --cluster create 127.0.0.1:7000 127.0.0.1:7001 \
        # 127.0.0.1:7002 127.0.0.1:7003 127.0.0.1:7004 127.0.0.1:7005 \
        # --cluster-replicas 1 \
    && echo "done" \
    && touch log.log \
    && tail -f ./log.log

