# Kubernetes deployment

RobotFleetOS can be deployed on Kubernetes: one **fleet** pod (all-in-one area/zone/edge), plus **mes**, **wms**, **traceability**, **qms**, **cmms**, and **plm** as separate deployments. All use the same container image with different commands.

---

## 1. Build and push the image

From the repo root, build the image (includes all binaries):

```bash
docker build -t robotfleetos/robotfleetos:latest .
```

Push to your registry (replace with your registry and tag):

```bash
# Docker Hub
docker tag robotfleetos/robotfleetos:latest YOUR_USERNAME/robotfleetos:latest
docker push YOUR_USERNAME/robotfleetos:latest

# GitHub Container Registry
docker tag robotfleetos/robotfleetos:latest ghcr.io/hansstam86/robotfleetos:latest
docker push ghcr.io/hansstam86/robotfleetos:latest
```

If you use a different image name, override it with Kustomize (see below) or edit the image in each file under `k8s/`.

---

## 2. Deploy with kubectl

### Fix "current-context is not set" or "connection refused to localhost:8080"

Your kubeconfig is pointing at the wrong place (port 8080 is the Fleet app, not Kubernetes). Do this first:

**1. List contexts (ignore errors for now):**
```bash
kubectl config get-contexts
```

**2. If you see `docker-desktop` in the list**, use it and set it as default:
```bash
kubectl config use-context docker-desktop
```

**3. If you do not see `docker-desktop`:** Enable Kubernetes in Docker Desktop (Settings → Kubernetes → "Enable Kubernetes"), wait until it shows "Running", then run step 1 again. Docker Desktop will add a `docker-desktop` context.

**4. If you have a context that uses `http://localhost:8080`:** That context is wrong. Your config file is usually `~/.kube/config` (or whatever `echo $KUBECONFIG` shows). Open it in an editor and:
- Find the `clusters:` entry whose `server:` is `http://localhost:8080` and remove that cluster (and any context that uses it), or
- Set `current-context: docker-desktop` at the top level, and ensure the cluster used by `docker-desktop` has `server: https://127.0.0.1:6443` (Docker Desktop shows this in Settings → Kubernetes).
Then run step 2 again to use `docker-desktop`.

**5. Confirm it works:**
```bash
kubectl config current-context
kubectl cluster-info
```
You should see something like "Kubernetes control plane is running at https://127.0.0.1:6443".

---

### Deploy

**Ensure Kubernetes is running** and the correct context is set (see above).

Create the namespace and all resources:

```bash
kubectl apply -k k8s/
```

If validation fails with "failed to download openapi", ensure the cluster is up, then retry or apply without validation:

```bash
kubectl apply -k k8s/ --validate=false
```

Or apply files one by one:

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/fleet.yaml
kubectl apply -f k8s/mes.yaml
kubectl apply -f k8s/wms.yaml
kubectl apply -f k8s/traceability.yaml
kubectl apply -f k8s/qms.yaml
kubectl apply -f k8s/cmms.yaml
kubectl apply -f k8s/plm.yaml
```

---

## 3. Use your own image (Kustomize)

Create `k8s/kustomization.yaml` (or add to the existing one) an `images` section:

```yaml
images:
  - name: robotfleetos/robotfleetos:latest
    newName: ghcr.io/hansstam86/robotfleetos
    newTag: latest
```

Then:

```bash
kubectl apply -k k8s/
```

---

## 4. Access the services

**Port-forward (quick local access):**

Paste this single line to forward all services to localhost (avoids zsh issues with parentheses or comments):

```bash
bash k8s/port-forward.sh
```

Or make the script executable once, then run it:

```bash
chmod +x k8s/port-forward.sh
./k8s/port-forward.sh
```

Or run each port-forward manually:

```bash
kubectl -n robotfleetos port-forward svc/fleet 8080:8080 &
kubectl -n robotfleetos port-forward svc/mes 8081:8081 &
kubectl -n robotfleetos port-forward svc/wms 8082:8082 &
kubectl -n robotfleetos port-forward svc/traceability 8083:8083 &
kubectl -n robotfleetos port-forward svc/qms 8084:8084 &
kubectl -n robotfleetos port-forward svc/cmms 8085:8085 &
kubectl -n robotfleetos port-forward svc/plm 8086:8086 &
```

- Fleet: http://localhost:8080  
- MES: http://localhost:8081  
- WMS: http://localhost:8082  
- Traceability: http://localhost:8083  
- QMS: http://localhost:8084  
- CMMS: http://localhost:8085  
- PLM: http://localhost:8086  
- ERP: http://localhost:8087  

**Tip:** Run one command per line. In zsh, do not paste lines that start with `#` or contain parentheses like `(Fleet)` — use `bash k8s/port-forward.sh` to avoid that.

**Ingress (optional):**  
Apply the sample Ingress and configure hosts (e.g. `fleet.robotfleetos.local` → cluster or minikube IP):

```bash
kubectl apply -f k8s/ingress.yaml
```

---

## 5. Check status

```bash
kubectl -n robotfleetos get pods
kubectl -n robotfleetos get svc
kubectl -n robotfleetos logs -l app=fleet -f
```

---

## 6. Teardown

```bash
kubectl delete -k k8s/
# or
kubectl delete namespace robotfleetos
```

---

## Layout

| File              | Purpose                                      |
|-------------------|----------------------------------------------|
| `k8s/namespace.yaml` | Namespace `robotfleetos`                    |
| `k8s/fleet.yaml`  | Fleet (all-in-one) Deployment + Service     |
| `k8s/mes.yaml`    | MES Deployment + Service (FLEET_API_URL)    |
| `k8s/wms.yaml`    | WMS Deployment + Service (FLEET_API_URL)     |
| `k8s/traceability.yaml` | Traceability Deployment + Service    |
| `k8s/qms.yaml`    | QMS Deployment + Service                     |
| `k8s/cmms.yaml`   | CMMS Deployment + Service (FLEET_API_URL)    |
| `k8s/plm.yaml`    | PLM Deployment + Service                     |
| `k8s/erp.yaml`    | ERP Deployment + Service (MES_API_URL)       |
| `k8s/kustomization.yaml` | Kustomize resource list                |
| `k8s/ingress.yaml`| Optional Ingress for host-based routing      |

MES, WMS, and CMMS get `FLEET_API_URL=http://fleet:8080` so they can call the Fleet service by name inside the cluster.
