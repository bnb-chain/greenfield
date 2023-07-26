Helm Chart Deployment for Greenfield Challenger

## Deployment

These are the steps to deploy the greenfield challenger application using Helm Chart V3.

We run these commands first to get the chart and test the installation.

```console
helm repo add bnb-chain https://chart.bnbchain.world/
helm repo update
helm show values bnb-chain/gnfd-challenger-testnet-values > testnet-challenger-values.yaml
helm install greenfield-challenger bnb-chain/gnfd-challenger -f testnet-challenger-values.yaml -n NAMESPACE --debug --dry-run
```

If dry-run runs successfully, we install the chart:

`helm install greenfield-challenger bnb-chain/gnfd-challenger -f testnet-challenger-values.yaml -n NAMESPACE`

## Common Operations

Get the pods lists by running this commands:

```console
kubectl get pods -n NAMESPACE
```
See the history of versions of ``greenfield-challenger`` application with command.

```console
helm history greenfield-challenger -n NAMESPACE
```

## How to uninstall

Remove application with command.

```console
helm uninstall greenfield-challenger -n NAMESPACE
```

## Parameters

The following tables lists the configurable parameters of the chart and their default values.

You **must** change the values according to the your aws environment parametes in ``greenfield-challenger/testnet-values.yaml`` file.

1. In `greenfield-config`, change: `aws_region`, `aws_secret_name`, `aws_bls_secret_name`, `private_key` and `bls_private_key`.

2. In `db_config`, change: `aws_region`, `aws_secret_name`, `password`, `username`, `url`.

3. In `containers`, change values of `AWS_REGION` and `AWS_SECRET_KEY`

