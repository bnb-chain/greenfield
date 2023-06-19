Helm Chart Deployment for Greenfield Validator

## Dependence
1. VMServiceScrape

## Deployment
1. `helm repo add bnb-chain https://chart.bnbchain.world/`
2. `helm repo update`
3. `helm install greenfield-validator bnb-chain/gnfd-validator`

## Common Operations

### Check Pod Status

```
$ kubectl get pod
```

You should see a `2/2` pod running for the validator. This means there are 2 containers running for the validator.

To see more details about a pod, you can describe it:

```
$ kubectl describe pod <POD_NAME> 
```

### Check the Pod Logs 

```
$ kubectl logs <POD_NAME> -c <CONTAINER_NAME>
