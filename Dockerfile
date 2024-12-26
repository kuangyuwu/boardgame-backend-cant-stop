# Use a lightweight debian os
# as the base image
FROM  debian:stable-slim

WORKDIR /app

# COPY source destination
COPY cantstop cantstop
CMD [ "./cantstop" ]