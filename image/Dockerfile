FROM scratch

ADD ca-bundle.crt /etc/ssl/certs/ca-certificates.crt
ADD aagregator_linux /aagregator

ENTRYPOINT ["/aagregator"]