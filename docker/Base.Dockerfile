FROM ubuntu

# Install basic stuff
RUN apt-get update
RUN apt-get install curl wget -y

# Install Go
RUN apt-get update
RUN apt-get install software-properties-common -y
RUN add-apt-repository ppa:longsleep/golang-backports
RUN apt-get install golang-go -y
RUN mkdir -p ~/go/src/github.com/luckybroman5

# Install Charles Proxy

RUN wget -q -O - https://www.charlesproxy.com/packages/apt/PublicKey | apt-key add -
RUN sh -c 'echo deb https://www.charlesproxy.com/packages/apt/ charles-proxy main > /etc/apt/sources.list.d/charles.list'
RUN apt-get update
RUN apt-get install charles-proxy -y

# Install Basic K6
RUN apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 379CE192D401AB61
RUN echo "deb https://dl.bintray.com/loadimpact/deb stable main" | tee -a /etc/apt/sources.list
RUN apt-get update
RUN apt-get install k6 -y
