[![Travis CI][Travis-Image]][Travis-Url]

# Skytap GO SDK

skytap-sdk-go is a Go client library for accessing the Skytap API.

You can view Skytap API docs here: [https://help.skytap.com/api.html](https://help.skytap.com/api.html)

### Setup for local development

- Ensure you have set up your Go environment correctly - see [here](https://golang.org/doc/code.html) for more details.

- Install Go (skytap-sdk-go is currently built against 1.11)

- Checkout repository

```
        mkdir -p "$GOPATH/src/github.com/skytap/"
        git clone https://github.com/skytap/skytap-sdk-go.git "$GOPATH/src/github.com/skytap/skytap-sdk-go"
        cd "$GOPATH/src/github.com/skytap/skytap-sdk-go"
```

### Build SDK

    make build

### Test SDK

    make test

[Travis-Image]: https://travis-ci.org/skytap/skytap-sdk-go.svg?branch=v2
[Travis-Url]: https://travis-ci.org/skytap/skytap-sdk-go