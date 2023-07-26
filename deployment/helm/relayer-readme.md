Helm Chart Deployment for Greenfield Relayer

## Deployment

These are the steps to deploy the greenfield relayer application using Helm Chart V3.

We run these commands first to get the chart and test the installation.

```console
helm repo add bnb-chain https://chart.bnbchain.world/
helm repo update
helm show values bnb-chain/gnfd-relayer-testnet-values > testnet-relayer-values.yaml
helm install greenfield-relayer bnb-chain/gnfd-relayer -f testnet-relayer-values.yaml -n NAMESPACE --debug --dry-run
```

If dry-run runs successfully, we install the chart:

`helm install greenfield-relayer bnb-chain/gnfd-relayer -f testnet-relayer-values.yaml -n NAMESPACE`

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

1. In `greenfield-config`, change: `private_key` and `bls_private_key`.

2. In `bsc-config`, change: `private_key`

3. In `db_config`, change: `password`, `username`, `url`.
