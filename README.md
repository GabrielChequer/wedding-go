## Update pods after code change

1. Edit code
2. Rebuild docker image
    * `docker build -t wedding-go:latest .`
3. restart the pod so kubernetes picks up the new image
    * `kubectl rollout restart deployment/wedding-go`
4. watch the pod restart
    * `kubectl get pods -w`
5. test

## Update deployment/service file

In case I want to add more replicas or change some settings, I have to reapply these files:
```bash
kubectl apply -f k8s/deployment.yaml
```
Or
```bash
kubectl apply -f k8s/service.yaml
```

It automatically sees what's up and updates as needed.