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
	docker build -t $(docker_img_name):$(docker_img_tag) .

.PHONY: scp
scp: build
	docker save -o $(docker_img_out) $(docker_img_name):$(docker_img_tag)
	scp -P $(remote_port) $(docker_img_out) $(files_to_transfer) $(remote):$(docker_remote_path)
	ssh -p $(remote_port) $(remote) "cd $(docker_remote_path) && bash $(docker_run_script)"

.PHONY: html
html:
	scp -r -P $(remote_port) $(PWD)/html $(remote):$(docker_remote_html)

.PHONY: run
run: serve
	docker run \
		--rm -ti --name $(docker_img_name) \
		-p 8080:80 \
		-v ./html:/var/html \
		-v ./config.json:/etc/config.json \
		-v ./public:/var/public \
		$(docker_img_name):$(docker_img_tag)
