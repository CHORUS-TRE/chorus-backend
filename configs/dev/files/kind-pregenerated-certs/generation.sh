openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key -subj "/CN=kind-ca" -days 10000 -out ca.crt

openssl genrsa -out admin.key 2048
openssl req -new -key admin.key -subj "/CN=admin" -out admin.csr
openssl x509 -req -in admin.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out admin.crt -days 10000
