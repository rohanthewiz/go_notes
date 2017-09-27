FROM golang

ENV WORKDIR /go/src/github.com/rohanthewiz/go_notes
#ENV APPDIR /app
ENV GLIDE_VERSION 0.12.3
ENV GLIDE_DOWNLOAD_URL https://github.com/Masterminds/glide/releases/download/v${GLIDE_VERSION}/glide-v${GLIDE_VERSION}-linux-amd64.tar.gz

RUN curl -fsSL "$GLIDE_DOWNLOAD_URL" -o glide.tar.gz \
    && tar -xzf glide.tar.gz \
    && mv linux-amd64/glide /usr/bin/ \
    && rm -r linux-amd64 \
    && rm glide.tar.gz

WORKDIR $WORKDIR

COPY glide.* $WORKDIR/
RUN glide install --force

# Copy the local package files to the container's workspace.
ADD . $WORKDIR

RUN go install github.com/rohanthewiz/go_notes

VOLUME ["/app"]
#COPY /go/bin/go_notes $APPDIR
#COPY config/options.yml.sample config/options.yml

#WORKDIR $APPDIR

# Document that the service listens on the specified.
EXPOSE 8092

CMD ["/go/bin/go_notes", "-svr", "-db", "/app/go_notes.sqlite"]
