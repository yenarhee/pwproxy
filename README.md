# pwproxy
A proxy that intercepts login requests and forward a duplicate to a specified host, using goproxy. Written for Bachelor thesis, together with pwstat.

Usage
=====

In order to use the proxy, you need to:
- configure the proxy address (e.g. in the browser settings)
- add the proxy server's certificate in the trusted list 

To run the proxy:
- change to the project directory
- set GOPATH environment variable: export GOPATH=`pwd`
- `go get github.com/elazarl/goproxy`
- run `go build` in this folder
- execute the binary file by `./pwproxy`

Options:
`-v`: verbose mode
`--addr`: specify address. default is on port 8080.


To create self-signed certificate:
`openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365`
To remove password from your key file:
`openssl rsa -in key.pem -out key.unencrypted.pem -passin pass:TYPE_YOUR_PASS`
