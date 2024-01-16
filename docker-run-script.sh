#!/bin/bash

docker ps | grep serve && docker stop serve || echo ""
docker load -i serve.tar
docker run -d --name serve \
	-p 443:443 \
	-p 80:80 \
	-v $(pwd)/html:/var/html \
	-v $(pwd)/public:/var/public \
	-v $(pwd)/tls:/var/tls \
	-v $(pwd)/config.json:/etc/config.json \
	--rm \
	serve:0.1
