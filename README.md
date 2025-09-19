limit-order-book

Run server:
go run main.go

K8S:

Caveats:
Run `eval $(minikube -p minikube docker-env)` to allow minikube to access localy built docker images
Build the image after running eval

- Make modifications to code
- Build docker image and tag it `docker build . -t limit-order-book:<TAG>`
- Update deployment.yaml and apply it with `kubectl apply -f deployment.yaml`
- Expose the service:
    `kubectl expose deployment/kubernetes-limit-order-book --type="LoadBalancer" --port 3000`
    NOTE: If you're running minikube with Docker Desktop as the container driver, a tunnel is needed:
    `minikube service kubernetes-limit-order-book --url`

- K8s dashboard `minikube dashboard`


Scale:
    kubectl scale deployments/kubernetes-limit-order-book --replicas=4

Rollout updated image:
    kubectl set image deployments/kubernetes-limit-order-book limit-order-book=limit-order-book:<TAG>

Undo Rollout:
    kubectl rollout undo deployments/kubernetes-limit-order-book


Example:
    Starting the tunnel returns url `http://127.0.0.1:<PORT>`
    We can then access it like this: `curl http://127.0.0.1:<PORT>/<ENDPOINT>`
