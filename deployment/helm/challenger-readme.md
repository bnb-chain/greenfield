Helm Chart Deployment for Greenfield Challenger

## Dependence
1. VMServiceScrape

## Deployment
1. `helm repo add bnb-chain https://chart.bnbchain.world/`
2. `helm repo update`
3. `helm install greenfield-challenger bnb-chain/gnfd-challenger`

## Common Operations

### Check Pod Status

```
$ kubectl get pod
```

You should see a `1/1` pod running for the challenger. This means there is 1 container running for the challenger.

To see more details about a pod, you can describe it:

```
$ kubectl describe pod <POD_NAME> 
```

### Check the Pod Logs 

```
$ kubectl logs <POD_NAME>
