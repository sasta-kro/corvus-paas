# This file uses a feature called Multi-Stage Builds.
# The goal is to use a heavy, tool-filled environment to compile the code, (like a build container)
# but then copy only the finished binary into a tiny, clean environment for the final product.
# This keeps the final image small and secure.

# Stage 1 starts the build environment (minimal alphine with Go) to compile the Go application.
# AS builder is like defining a variable to reference this stage later
FROM golang:1.25.7-alpine AS builder

# Installing gcc and musl-dev to provide C compilers required by go-sqlite3 for CGO support.
# RUN = executes a terminal command inside the container
# apk = the Alpine Linux's package manager. add = install packages. Its like (dnf install ...)
RUN apk add --no-cache gcc musl-dev git

# this creates the folder and set it as default location for all subsequent docker instructions.
WORKDIR /app

# Copying go.mod and go.sum files first to utilize Docker layer caching for downloaded modules.
# (so if these files are not changed, docker doesn't redownload the dependencies)
COPY go.mod go.sum ./
RUN go mod download

# Copying the remaining application source code into the build environment (which is /app)
COPY . .

# This executes/runs the Go compiler in the contianer.
# CGO_ENABLED=1 turns on CGO, allowing Go packages to call C code (necessary for SQLite)
# GOOS=linux explicitly tells the compiler to build a binary meant to run on Linux
# `go build -o /corvus-paas .` = build code in cwd and put it into `root/corvus-pass` dir
RUN CGO_ENABLED=1 GOOS=linux go build -o /corvus-paas .

# Stage 2 starts a new minimal Alpine Linux environment to reduce the final image size.

# just alphine img with no Go stuff
FROM alpine:latest

WORKDIR /app


# tzdata = provides timezone data, necessary for accurate timezone information for application logging.
# ca-certificates = installs the public certificates of trusted Certificate Authorities.
# Without this, the Go application cannot make secure HTTPS requests to external APIs (like GitHub)
# because it cannot verify the server's SSL certificate.
RUN apk add --no-cache tzdata ca-certificates

# This reaches back into the stage named builder, grabs the compiled /corvus-paas binary file,
# and places it inside the current container at /app/corvus-paas.
COPY --from=builder /corvus-paas /app/corvus-paas

# (just for documentation) container listens on port 8080 to inform network and port mapping configurations.
# This doeen't automatically publish the port to the host machine,
# that must be done via docker run or docker-compose.yml
EXPOSE 8080

# Defining the default command to execute the compiled app binary when the container starts.
# The square bracket syntax (exec form) is preferred because it runs the executable directly,
# allowing it to receive system signals (like shutdown commands) gracefully.
CMD ["/app/corvus-paas"]