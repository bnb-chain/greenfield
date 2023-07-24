Helm Chart Deployment for Greenfield Relayer

## Dependence
1. VMServiceScrape

## Deployment
1. `helm repo add bnb-chain https://chart.bnbchain.world/`
2. `helm repo update`
3. `helm install greenfield-relayer bnb-chain/gnfd-relayer`

## Common Operations

### Check Pod Status

```
$ kubectl get pod
```

You should see a `1/1` pod running for the relayer. This means there is 1 container running for the relayer.

To see more details about a pod, you can describe it:

```
$ kubectl describe pod <POD_NAME> 
```

### Check the Pod Logs 

```
$ kubectl logs <POD_NAME>
