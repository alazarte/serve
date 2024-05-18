remote_user=web
remote_host=remotehost
remote_dir=/home/$(remote_user)/serve/
remote=$(remote_user)@$(remote_host)

.PHONY: all
all: serve rsync

.PHONY: rsync
rsync:
	rsync --delete -rav -e "ssh" --exclude=".git" . \
		$(remote):$(remote_dir)
