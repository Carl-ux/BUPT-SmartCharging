![Static Badge](https://img.shields.io/badge/Go-v1.19-blue)
# BUPT-SmartCharging
北邮2023 软件工程 校内充电桩计费调度系统后端
> Go + Gin + Gorm + MySQL
> 使用配置管理工具Viper

- api接口文档 https://apifox.com/apidoc/shared-1193e345-a045-4635-9e58-6f3e1a1fc9da/api-75747819

### Build Setup
```bash
#clone Repository
git clone https://github.com/Carl-ux/BUPT-SmartCharging.git
cd BUPT-SmartCharging

#修改数据库DB配置
cd config
可以根据自己的数据库信息修改application.yml文件

# 构建 Docker 镜像
docker build -t my-go-app .

# 运行 Docker 镜像
docker run -p 8000:8000 my-go-app

#这将会构建一个名为my-go-app的Docker镜像，并将其运行在本地的8000端口上
```

### Guide
```bash
├── common
│   ├── database.go                 -数据库信息与连接
│   └── jwt.go                      -jwt鉴权
├── config
│   ├── application.yml
│   └── viper.go
├── controller                      -控制器目录
│   ├── AdminController.go
│   ├── AuthController.go
│   ├── GenericController.go
│   └── UserController.go
├── main.go
├── middleware                      -中间件目录
│   ├── AuthMiddleware.go               -用户信息中间件
│   ├── CORSMiddleware.go               -跨域中间件
│   └── RecoveryMiddleware.go
├── model                           -模型定义目录
│   ├── charger.go
│   ├── order.go
│   ├── record.go
│   ├── user.go
│   └── user_dto.go                 -user_dto.go 数据传输对象
├── routes.go                       -定义路由信息
├── service                         -服务逻辑代码
│   ├── charge.go                   -实现计费功能
│   ├── schedule                    -调度功能目录*
│   │   ├── chargingRequest.go          -充电请求定义
│   │   ├── pileSchedule.go             -充电桩调度器定义
│   │   └── scheduler.go                -系统调度器定义
│   └── timemock.go                 -实现模拟时间
├── utils
│   ├── response
│   │   └── response.go             -封装了响应成功与失败的代码
│   └── utils.go                    -工具类
└── 详细需求.md
```

## License
![GitHub](https://img.shields.io/github/license/Carl-ux/BUPT-SmartCharging)
