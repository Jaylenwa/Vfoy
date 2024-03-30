# Vfoy
云盘系统


# 构建静态资源

```
# 进入前端子模块
cd assets
# 安装依赖
yarn install
# 开始构建
yarn build
# 构建完成后删除映射文件
cd build
find . -name "*.map" -type f -delete
```

完成后，所构建的静态资源文件位于 assets/build 目录下。

将build目录改名为statics 目录，放置在 主程序同级目录下并重启服务。