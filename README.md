docker volume plugin    
为了解决容器和主机都能访问同一个目录，并且限制目录大小。volume挂载后可以在母机器/var/lib/docker-volumes/_mydriver/mnt/ 里面看到    
go build mydriver.go    
./mydirver    
docker volume create --name test --driver mydriver
docker run -it -v test:/data centos:7 bash    
df -h    
/dev/loop5                                                                                       10G   33M   10G   1% /data    
    
ls /var/lib/docker-volumes/_mydriver/mnt/    
