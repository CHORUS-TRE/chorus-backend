Generated using

```bash
openssl ecparam -name prime256v1 -genkey -noout -out chorus_privkey.pem
openssl ec -in chorus_privkey.pem -pubout -out chorus_pubkey.pem
```