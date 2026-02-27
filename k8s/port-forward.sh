#!/usr/bin/env bash
# Port-forward all RobotFleetOS services to localhost. Run in background or in separate terminals.
set -e
NS="${ROBOTFLEETOS_NS:-robotfleetos}"

echo "Port-forwarding namespace $NS to localhost (8080-8087)..."
kubectl -n "$NS" port-forward svc/fleet 8080:8080 &
kubectl -n "$NS" port-forward svc/mes 8081:8081 &
kubectl -n "$NS" port-forward svc/wms 8082:8082 &
kubectl -n "$NS" port-forward svc/traceability 8083:8083 &
kubectl -n "$NS" port-forward svc/qms 8084:8084 &
kubectl -n "$NS" port-forward svc/cmms 8085:8085 &
kubectl -n "$NS" port-forward svc/plm 8086:8086 &
kubectl -n "$NS" port-forward svc/erp 8087:8087 &

echo "Fleet http://localhost:8080  MES 8081  WMS 8082  Traceability 8083  QMS 8084  CMMS 8085  PLM 8086  ERP 8087"
echo "Press Ctrl+C to stop all port-forwards."
wait
