#秒杀系统：接入层
分接入层->逻辑层(暂时不实现)->后台

##环境配置
golang配置
go env -w GOPROXY=https://goproxy.cn

##项目配置
bee new SecKill
go mod init SecKill
go mod download
go mod vendor //第三方包复制到项目
bee run
sudo chmod -R 777 logs/

##中间件启动
docker start redis
cd etcd_file
docker-compose up -d
docker-compose start

##数据模拟
tools添加数据：
运行etcd main.go,插入商品模拟数据到etcd

12:00



