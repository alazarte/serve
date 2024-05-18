remote=serverhostname
remote_port=22
docker_img_name=serve
docker_img_tag=0.1
docker_img_out=$(docker_img_name).tar
docker_remote_path=/docker/serve
docker_remote_html=$(docker_remote_path)/html
docker_run_script=docker-run-script.sh
serve_config_script=config.json
files_to_transfer=$(serve_config_script) $(docker_run_script)

all: serve build scp

.PHONY: build
build:
	sudo docker build -f Dockerfile -t $(docker_img_name):$(docker_img_tag) ..

.PHONY: scp
scp: build
	sudo docker save -o $(docker_img_out) $(docker_img_name):$(docker_img_tag)
	scp -P $(remote_port) $(docker_img_out) $(files_to_transfer) $(remote):$(docker_remote_path)
	ssh -p $(remote_port) $(remote) "cd $(docker_remote_path) && bash $(docker_run_script)"

.PHONY: html
html:
	scp -r -P $(remote_port) $(PWD)/html $(remote):$(docker_remote_html)

.PHONY: run
run:
	sudo docker run \
		--rm -ti --name $(docker_img_name) \
		--network web \
		-v $(PWD)/../web:/var/web \
		-v $(PWD)/../config/config.json:/etc/config.json \
		-v cert.pem:/etc/cert.pem \
		-v privkey.pem:/etc/privkey.pem \
		-p 80:80 \
		-p 443:443 \
		$(docker_img_name):$(docker_img_tag)
