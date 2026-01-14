FROM node:20 as build-web
WORKDIR /build

COPY ./web/package.json ./web/package.lock ./
RUN yarn install --frozen-lockfile

COPY ./web .
RUN yarn build

FROM golang:1.25.4 as build

# Set the Current Working Directory inside the container
WORKDIR /build

# Copy the modules files
COPY go.mod .
COPY go.sum .

# Download the modules
RUN go mod download

# Copy rest fo the code
COPY . .

# Copt the web build into the expected folder
COPY --from=build-web /build/dist ./web/dist

RUN CGO_ENABLED=0 go build -buildvcs=false -o ./bin/go-liift ./main.go

FROM alpine:3.14

COPY --from=build /build/bin/go-liift /usr/bin/go-liift

# This container exposes port 3000 to the outside world
EXPOSE 3000

# Run the executable
CMD ["/usr/bin/go-vite"]
