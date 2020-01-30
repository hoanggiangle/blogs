ARG gover=1
FROM golang:${gover}

RUN apt-get update 2>&1 | tee /tmp/apt.err \
	&& ! grep -q '^[WE]' /tmp/apt.err \
	&& R=$? \
	&& rm -f /tmp/apt.err \
	&& exit $R

RUN apt-get install -y sudo git

# utils
RUN apt-get install -y bash-completion ngrep net-tools bwm-ng zip vim lsof knot-dnsutils telnet

RUN dep_ver=`curl --silent "https://api.github.com/repos/golang/dep/releases/latest" | grep -Po '"tag_name": "\K.*?(?=")'` \
	&& curl -L https://github.com/golang/dep/releases/download/$dep_ver/dep-linux-amd64 > /usr/local/bin/dep \
	&& chmod +x /usr/local/bin/dep

ARG USER_ID
RUN useradd -o -u ${USER_ID:-1000} -m -s /bin/bash dev \
	&& echo "dev ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/dev \
	&& echo 'Defaults:dev env_keep="ftp_proxy http_proxy https_proxy no_proxy"' \
		>> /etc/sudoers.d/dev

RUN chown -R dev /go

USER dev

RUN go get github.com/derekparker/delve/cmd/dlv && rm -rf $HOME/.cache/go-build/*

VOLUME /home/dev/.cache/go-build

COPY entrypoint.sdk.sh /usr/local/bin/docker-entrypoint.sh
ENTRYPOINT ["docker-entrypoint.sh"]
