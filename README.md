docker volume plugin    
go build mydriver.go    
./mydirver    
docker volume create --name test --driver mydriver
docker run -it -v test:/data centos:7 bash    
df -h
`
/dev/loop5                                                                                       10G   33M   10G   1% /data
`
ls /var/lib/docker-volumes/_mydriver/mnt/    
