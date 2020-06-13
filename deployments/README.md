# Deploying Authcore in Kubernetes

## Install Helm

Follow the [instructions](https://helm.sh/docs/using_helm/#installing-helm) to install Helm.
Please note: the default Helm installation is insuecre. Follow this additional guide to [secure
Helm](https://docs.helm.sh/using_helm/#securing-your-helm-installation).

For a minimal installation:

```bash
$ kubectl create -f kubernetes/rbac-config.yaml
serviceaccount "tiller" created
clusterrolebinding "tiller" created
$ helm init --service-account tiller --history-max 200
```

## Create a namespace for the deployment

```bash
$ kubectl create namespace authcore
```

## Install Dependencies

### MySQL
Install MySQL:

```bash
$ helm install --namespace authcore --set imageTag=8.0 --set mysqlDatabase=authcore --name authcore-mysql stable/mysql
```

After installation, you need to create a secret for database URL providing to Authcore.
```bash
$ kubectl create secret generic <secret-name> --from-literal=database_url=<database_url> --namespace authcore
```
where `<database_url>` is `mysql://root:<mysql-root-password>@<mysql_service>/authcore?parseTime=true`

And put the `<secret-name>` in `mysql.database_url`.

### Redis
Redis can be installed via dependency. 
If you want to install redis by your own configuration, you can do it by following.

Install Redis (optional):

```bash
helm install --namespace authcore --name authcore-redis stable/redis -f values.yaml
```

You must also set `tags.install_redis` and configuration in `redis` in `values.yaml` to `false` if you choose to customize install redis.

## Install Authcore

Set up secret for accessing Authcore docker registry:

```bash
$ kubectl create secret docker-registry authcore-registry --namespace authcore --docker-server=registry.gitlab.com --docker-username=<deploy-token-username> --docker-password=<deploy-token-password> --docker-email=<email>
```

Generate random string for secret_key_base:

```bash
openssl rand -hex 32
```

The server requires a ECDSA key to sign JWT token. Generate the key pair using the following commands:

```sh
$ openssl ecparam -genkey -name prime256v1 -noout -out private.pem
$ openssl ec -in private.pem -pubout -out public.pem
```

Change the values in `values.yaml` to override default values for OAuth configuration, root keys etc.

```bash
$ helm install --namespace authcore --name authcore-app charts/authcore -f values.yaml
```

TLS configuration:

First add the TLS secret:

``` bash
$ kubectl create secret tls ${CERT_NAME} --key ${KEY_FILE} --cert ${CERT_FILE}
```

then refer the TLS secret (and the hostname) in values.yaml

for the nginx ingress, a ingress controller must be installed.

Refer to `https://kubernetes.github.io/ingress-nginx/deploy/` for the ingress controller deployment guide.