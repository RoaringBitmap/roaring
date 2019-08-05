from amazonlinux:latest
RUN yum install -y tar gzip git gcc
ENV GOVERSION 1.12.7
ENV ARCH linux-amd64
ENV GOROOT $PWD/go-$GOVERSION.$ARCH
RUN  mkdir $GOROOT
ENV GOPATH $PWD/build
ENV PATH $PWD/bin:$GOPATH/bin:$GOROOT/bin:$PATH
RUN curl -X GET "https://storage.googleapis.com/golang/go$GOVERSION.$ARCH.tar.gz" | tar -zx -C $GOROOT --strip-components=1
RUN git clone https://github.com/jacksonrnewhouse/roaring.git
WORKDIR roaring/
RUN git checkout inline
RUN go test -run "^$"
CMD go test -run TestComparison
