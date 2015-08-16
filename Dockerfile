# Docker image for the itamae plugin
#
#     docker build --rm=true -t plugins/drone-itamae .

FROM ubuntu:14.04

RUN apt-get update 

RUN apt-get --force-yes -y install curl                        
RUN apt-get --force-yes -y install apt-transport-https                       
RUN apt-get --force-yes -y install ca-certificates                              
RUN apt-get --force-yes -y install lxc     
RUN apt-get --force-yes -y install aufs-tools
RUN apt-get --force-yes -y install iptables 
RUN curl -sSL https://get.docker.com/ubuntu/ -o /tmp/install.sh  
RUN sed -i -e "s/apt-get install/apt-get install --force-yes -y/g" /tmp/install.sh
RUN sh /tmp/install.sh

RUN apt-get -y install software-properties-common
RUN add-apt-repository -y ppa:brightbox/ruby-ng
RUN apt-get update
RUN apt-get -y --force-yes install ruby2.2
RUN apt-get -y --force-yes install ruby-dev

RUN apt-get -y install  openssh-client
RUN ssh-keygen -t rsa -N "" -f /root/.ssh/id_rsa
RUN rm -rf /var/lib/apt/lists/*

RUN gem install itamae

ADD drone-itamae /go/bin/
ADD wrapdocker /bin/

VOLUME /var/lib/docker
ENTRYPOINT ["/go/bin/drone-itamae"]
