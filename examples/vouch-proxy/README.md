To start the service, execute:

```bash
docker run -p 9090:9090 -v ${PWD}/config:/config -v ${PWD}/data:/data voucher/vouch-proxy
```