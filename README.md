# oidc-cli

`oidc-cli` assists developers in automating authorization flow for local development and testing purpose.

## Installation

Download the program from [Release](https://github.com/ycio/oidc-cli/releases) page, rename it to `oidc-cli` and make it executable, e.g:

```
mv oidc-cli_darwin_amd64 oidc-cli
chmod +x oidc-cli
```

Run the command to print the help messages:

```
./oidc-cli implicit-flow -h
```

By default `oidc-cli` is blocked by Mac OSX, you must click `Allow Anyway` from `Security & Privacy`.

## Usage

Currently it supports fetch OIDC token via implicit flow.

### Fetch OIDC token via implicit flow

> Note that Chrome must be installed.

Given the open id configuration endpoint `http://localhost:8091/auth/realms/awesome-realms/.well-known/openid-configuration`, redirect uri `http://localhost:8080/auth` and client id `awesome-application`, run the following command to open Chrome and sign in for token printed to console:

```bash
./oidc-cli implicit-flow  -e http://localhost:8091/auth/realms/awesome-realms/.well-known/openid-configuration -r http://localhost:8080/auth -c awesome-application
```
