# Akachain Golang Software Development Kit - v2

[![Go Report Card](https://goreportcard.com/badge/github.com/Akachain/akc-go-sdk-v2)](https://goreportcard.com/report/github.com/Akachain/akc-go-sdk-v2)

golang SDK that supports writing chaincodes on Hyperledger Fabric version 2+ with minimum setup.

We have written a few dozen of complex chaincodes for production environment of enterprise use cases since 2018. 
We found that it would give junior developers tremendous help if they are provided with several utilities to 
quickly interact with the state database as well as mockup tool to quickly test their chaincode without requiring a real Hyperledger Fabric Peer.

### Getting started
Obtaining the Akachain go SDK package

``
go get https://github.com/Akachain/akc-go-sdk-v2
``

We recommend to use the SDK with ``gomod``

### Documentation

Our document is available at [pkg.go.dev](https://pkg.go.dev/github.com/Akachain/akc-go-sdk-v2)

### Examples
Sample Test Code is available at

[sample_test](test/contract/sample_test.go): Basic example that uses SDK to query and execute transaction with a CouchDB state database

### License
This source code are made available under the MIT license, located in the [LICENSE](LICENSE) file. You can do whatever you want with them, we do not bother. But if you have some nice idea that wants to share back with us, please do. 

To add MIT license header to new source code files, please use the following snippet
````
go get -u github.com/google/addlicense
addlicense -c akachain -l mit -y 2021 .
````