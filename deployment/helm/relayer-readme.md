Helm Chart Deployment for Greenfield Relayer

## Deployment

These are the steps to deploy the greenfield relayer application using Helm Chart V3.

```console
helm repo add bnb-chain https://chart.bnbchain.world/
helm repo update
helm show values bnb-chain/gnfd-relayer-testnet-values > testnet-values.yaml
helm install greenfield-relayer bnb-chain/gnfd-relayer -f testnet.values.yaml -n NAMESPACE --dry-run
```

## Common Operations

Get the pods lists by running this commands:

```console
kubectl get pods -n NAMESPACE
```
See the history of versions of ``greenfield-relayer`` application with command.

```console
helm history greenfield-relayer -n NAMESPACE
```

## How to uninstall

Remove application with command.

```console
helm uninstall greenfield-relayer -n NAMESPACE
```

## Parameters

The following tables lists the configurable parameters of the chart and their default values.

You **must** change the values according to the your aws environment parametes in ``greenfield-relayer/testnet-values.yaml`` file.

1. In `greenfield-config`, change: `aws_region`, `aws_secret_name`, `aws_bls_secret_name`, `private_key` and `bls_private_key`.

2. In `bsc-config`, change: `aws_region`, `aws_secret_name` and `private_key`

3. In `db_config`, change: `aws_region`, `aws_secret_name`, `password`, `username`, `url`.

4. In `containers`, change values of `AWS_REGION` and `AWS_SECRET_KEY`
