.PHONY: build deploy clean status logs stop

build:
	go build -o gogogopher

deploy: build
	sudo cp gogogopher /usr/local/bin
	sudo setcap cap_net_bind_service=+ep /usr/local/bin/gogogopher
	sudo mkdir -p /etc/gogogopher /var/gogogopher
	sudo cp config.toml /etc/gogogopher
	sudo cp -r homedir/* /var/gopher
	sudo useradd -r -s /bin/false gopher 2>/dev/null || true
	sudo chown -R gopher:gopher /etc/gogogopher /var/gogogopher
	sudo cp gogogopher.service /usr/lib/systemd/system
	sudo systemctl daemon-reload
	sudo systemctl enable --now gogogopher
	@echo "Deployment complete."

status:
	sudo systemctl status gogogopher

stop:
	sudo systemctl stop gogogopher