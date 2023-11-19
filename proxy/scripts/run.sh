#! /bin/bash

docker rm -f proxy

docker build -t proxy:latest .

if [[ $? -ne 0 ]] ; then
    echo "Can't build, checksum failed" | logger --server 127.0.0.1 --port 5514 --priority user.error
    exit 1
fi

docker run --name proxy -dit -p 8080:8080 --network proxy_net  proxy:latest
