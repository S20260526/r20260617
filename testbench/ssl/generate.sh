set -e

rm -f *.csr *.key *.crt

openssl genrsa -out ca.key 4096
openssl req -noenc -x509 -key ca.key -days 36500 -out ca.crt -batch

openssl genrsa -out entry.key 4096
openssl req -new -key entry.key -out entry.csr -config entry.cnf
openssl x509 -req -in entry.csr -CA ca.crt -CAkey ca.key -out entry.crt -days 3650 -extfile entry.cnf -extensions v3_req

openssl genrsa -out hello.key 4096
openssl req -new -key hello.key -out hello.csr -config hello.cnf
openssl x509 -req -in hello.csr -CA ca.crt -CAkey ca.key -out hello.crt -days 3650 -extfile hello.cnf -extensions v3_req

openssl genrsa -out world.key 4096
openssl req -new -key world.key -out world.csr -config world.cnf
openssl x509 -req -in world.csr -CA ca.crt -CAkey ca.key -out world.crt -days 3650 -extfile world.cnf -extensions v3_req
