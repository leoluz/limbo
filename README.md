# Limbo

This project is a Kubernetes POC to validate how it is possible to isolate a mount point created in a user namespace from other processes running in the host system.

# Running the POC

**Requirements**
- Docker CLI
- Kubernetes running locally

This project uses Linux only capabilities that aren't available in Go sdk if you are on a MacOS or Windows system.
The `deploy.sh` script will build the project inside a docker image and deploy it in your local kubernetes.
Make sure your current kubectl context points to your local Kubernetes server before proceeding.

Reference:
- [Namespaces in Go](https://medium.com/@teddyking/namespaces-in-go-basics-e3f0fc1ff69a)
- [Linux namespace in Go](https://songrgg.github.io/programming/linux-namespace-part01-uts-pid/)
- [Building Container Images Securely on Kubernetes](https://blog.jessfraz.com/post/building-container-images-securely-on-kubernetes/)
- [Improving Kubernetes and container security with user namespaces](https://kinvolk.io/blog/2020/12/improving-kubernetes-and-container-security-with-user-namespaces/)
- [Creating user namespaces inside containers](https://frasertweedale.github.io/blog-redhat/posts/2021-10-15-openshift-userns-in-container.html)
