kubectl delete -f manifests && \
docker build -t limbo:local . && \
kubectl apply -f manifests
