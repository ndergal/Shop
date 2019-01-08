FROM golang

# Copy the local package files to the container's workspace.
COPY udp.go .
COPY key.txt .

RUN go get -u github.com/gorilla/mux
 
# Run the outyet command by default when the container starts.
ENTRYPOINT go run udp.go

# Document that the service listens on port 8080.
EXPOSE 8080
