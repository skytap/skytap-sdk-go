# Skytap GO SDK
## API

API package provides go based REST client.

## API Tests
The tests use canned API responses downloaded from the production service and
slightly sanitized. Using this data, they validate that the API calls  are being
made correctly.

    export GOPATH=`pwd`
    go get -t github.com/skytap/skytap-sdk-go/api
    cd api
    go test -v

## License
Apache 2.0; see [LICENSE](LICENSE) for details
