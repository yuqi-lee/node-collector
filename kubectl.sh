kubectl get po -n hotel-reservation -o wide |  awk '/skv-node4/{print $1, $6}'
# kubectl get po -n hotel-reservation -o wide |  awk '/skv-node3/{print $1, $6}'

kubectl exec -it -n hotel-reservation consul-785fdcc5bc-9rf8g -- wc -l /proc/net/tcp