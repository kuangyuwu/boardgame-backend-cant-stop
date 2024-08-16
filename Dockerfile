# Use a lightweight debian os
# as the base image
FROM debian:stable-slim

WORKDIR /app

# COPY source destination
COPY boardgame-backend-cant-stop boardgame-backend-cant-stop
CMD [ "./boardgame-backend-cant-stop" ]