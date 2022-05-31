# software-engineering

![workflow](https://github.com/evpeople/software-engineering/actions/workflows/go.yml/badge.svg)

## 开发前的准备

`go env -w GOPROXY=https://goproxy.cn,direct`
配置go代理，加速外部库的下载。
先配置Go代理之后，再通过VSC打开这个项目，从而尽量包装成功的安装后Go的开发套件，

[gin中文文档](https://gin-gonic.com/zh-cn/docs/examples/)

## 保底方案
[保底的后台管理系统](https://learnku.com/docs/gin-gonic/1.7/go-gin-document/11352)

## 使用方式
1. docker-compose up
2. go run main.go