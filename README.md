### TCP Meter 

#### Overview
TCP Meter is a Go utility designed to proxy TCP connections while monitoring the incoming and outgoing bandwidth. It listens on a local address and forwards traffic to a specified remote address, logging the traffic rate in real-time.

#### Features
- **TCP Proxy**: Forwards TCP connections from a local address to a remote address.
- **Bandwidth Monitoring**: Logs the incoming and outgoing traffic rates in real-time.
- **JSON Logging**: Outputs logs in JSON format for easy integration with monitoring tools.

#### Installation
Ensure you have Go installed on your system. Clone the repository and build the utility:
```bash
go install github.com/JackKCWong/tcp-meter@latest
```

#### Usage
Run the utility with the following command:
```bash
tcpmtr -l <local-address> -r <remote-address>
```
- `-l <local-address>`: The local address to listen on. Defaults to `:8080`.
- `-r <remote-address>`: The remote address to forward traffic to. This is a required parameter.


#### Logging
The utility logs traffic information in JSON format, including the connection ID, incoming and outgoing traffic rates, and the total number of bytes transferred. Logs are printed to the standard output.

```json
{"time":"2025-03-29T21:35:34.44401+08:00","level":"INFO","msg":"traffic","id":1,"in":"0 B/s","out":"0 B/s","bytesIn":5524,"bytesOut":619}
```