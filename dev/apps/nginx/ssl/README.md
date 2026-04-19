# Generate PEM + Key

## Documentation

[Generate a Self-Signed Certificate with Subject Alt Name (stackoverflow)](https://stackoverflow.com/questions/21488845/how-can-i-generate-a-self-signed-certificate-with-subjectaltname-using-openssl)

## Instructions

1. `cd dev/apps/nginx/ssl`.

2. Delete current certs as these will be re-generated

```
rm ./out/certificate.pem
rm ./out/private.key
```

3. Create and run an interactive nginx:latest container (it contains OpenSSL)

```
$ docker pull nginx:latest
$ docker run --rm -i -t -v $PWD/out/openssl.cnf:/etc/ssl/openssl.cnf -v $PWD/out/:/root/out --name openssl nginx bash
```

4. Generate a private key and a certificate from within the container. The openssl.cnf file is pre-populated with all necessary values. Accept all default values.

```
   $ cd /root/out
   $ openssl genrsa -out private.key 3072
   $ openssl req -new -x509 -key private.key -sha256 -out certificate.pem -days 3650
   $ openssl x509 -in certificate.pem -text -noout

```

5. Exit/Stop the container.
6. Set file permissions

```
sudo chmod 644 ./out/certificate.pem
sudo chmod 644 ./out/private.key
```

7. Copy and rename pem and key into develop folder (e.g. develop-docker).

```
mv ./out/certificate.pem ../local.4ks.io.pem
mv ./out/private.key ../local.4ks.io.key
```

# Convert to pfx (for Firefox)

## Documentation

[How to Create a PFX/P12 Certificate](https://www.ssl.com/how-to/create-a-pfx-p12-certificate-file-using-openssl/)

## Instructions

```
openssl pkcs12 -export -out local.4ks.io.pfx -inkey local.4ks.io.key -in local.4ks.io.pem
<password> is 4ks
```

## Import into Firefox

https://knowledge.digicert.com/solution/SO5437.html
