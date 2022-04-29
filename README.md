<h1 align="center">
  <a href="https://github.com/YojimboSecurity/skytap-sdk-go">
    <!-- Please provide path to your logo here -->
    <img src="docs/images/icon_108.png" alt="Logo" width="108" height="108">
  </a>
</h1>

<div align="center">
  skytap-sdk-go
  <br />
  <a href="#about"><strong>Explore the docs »</strong></a>
  <br />
  <br />
  <a href="https://github.com/YojimboSecurity/skytap-sdk-go/issues/new?assignees=&labels=bug&template=01_BUG_REPORT.md&title=bug%3A+">Report a Bug</a>
  ·
  <a href="https://github.com/YojimboSecurity/skytap-sdk-go/issues/new?assignees=&labels=enhancement&template=02_FEATURE_REQUEST.md&title=feat%3A+">Request a Feature</a>
  .
  <a href="https://github.com/YojimboSecurity/skytap-sdk-go/issues/new?assignees=&labels=question&template=04_SUPPORT_QUESTION.md&title=support%3A+">Ask a Question</a>
</div>

<div align="center">
<br />

![Go version](https://img.shields.io/badge/Go-v1.18-blue)

</div>

<details open="open">
<summary>Table of Contents</summary>

- [About](#about)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
- [Usage](#usage)
- [Contributing](#contributing)
- [Security](#security)
- [License](#license)
- [Acknowledgements](#acknowledgements)

</details>

---

## About

This package provides a Go client for the [Skytap API](https://help.skytap.com/API_v2_Documentation.htm).

## Getting Started

Get started by installing the package:

```bash
go get -t github.com/YojimboSecurity/skytap-sdk-go
```

### Prerequisites

All you need is Go.

## Usage

First, create a client:

```go
token := os.Getenv("SKYTAP_TOKEN")
user := os.Getenv("SKYTAP_USER")
if token == "" || user == "" {
    panic("SKYTAP_TOKEN and SKYTAP_USER must be set")
}
client := api.NewSkytapClient(user, token)
```

Here, I am using the environment variables to set the client. To set 
them, you can use the following command:

```bash
export SKYTAP_TOKEN=<your token>
export SKYTAP_USER=<your user>
```

Next, you can use the client to make API calls:

```go
resp := interface{}(nil)
api.GetSkytapResource(*client, "https://cloud.skytap.com/v2/configurations?scope=company&count=40", &resp)
```

Here, I am using the `GetSkytapResource` method to make a call to the 
`/v2/configurations` endpoint. The `resp` variable is used to store the data. This returns all the environment configurations in the company.

To get an environment configuration, you can use the following call:

```go
configurationId := "12345" // Replace with your configuration id
URL := fmt.Sprintf("https://cloud.skytap.com/v2/configurations/%s.json", configurationId)
resp := interface{}(nil)
api.GetSkytapResource(*client, URL, &resp)
```

To a virtual machine, you can use the following call:

```go
vmId := "12345" // Replace with your virtual machine id
vm, err = api.GetVirtualMachine(*client, vmId)
if err != nil {
    log.Error(err)
}
```

To change the state of a virtual machine, you can use the following call:

```go
vm, err := vm.Start(*client)
if err != nil {
    log.Error(err)
}
```

### Test

The tests use canned API responses downloaded from the production service and
slightly sanitized. Using this data, they validate that the API calls  are being
made correctly.

```bash
export GOPATH=`pwd`
go get -t github.com/skytap/skytap-sdk-go/api
cd api
go test -v
```

## Contributing

First off, thanks for taking the time to contribute! Contributions are what make the open-source community such an amazing place to learn, inspire, and create. Any contributions you make will benefit everybody else and are **greatly appreciated**.

Please read [our contribution guidelines](docs/CONTRIBUTING.md), and thank you for being involved!

## Security

skytap-sdk-go follows good practices of security, but 100% security cannot be assured.
skytap-sdk-go is provided **"as is"** without any **warranty**. Use at your own risk.

_For more information and to report security issues, please refer to our [security documentation](docs/SECURITY.md)._

## License

This project is licensed under the **Apache Software License 2.0**.

See [LICENSE](LICENSE) for more information.

## Acknowledgements

I would like to acknowledge that this is a fork of the original [skytap-sdk-go](https://github.com/skytap/skytap-sdk-go) project.

I would also like to thank the following projects for their contributions:

- [Awesome Readme Templates](https://awesomeopensource.com/project/elangosundar/awesome-README-templates)
- [Awesome README](https://github.com/matiassingers/awesome-readme)
- [How to write a Good readme](https://bulldogjob.com/news/449-how-to-write-a-good-readme-for-your-github-project)
- [shields.io](https://shields.io/)
