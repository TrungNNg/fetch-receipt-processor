FROM debian:stable-slim

# COPY source destination
COPY /bin/fetch /bin/fetch

CMD ["/bin/fetch"]