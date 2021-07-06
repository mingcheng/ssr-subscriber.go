# SSR Subscriber

A command-line tool for subscribe ssr config and check availability.

基于 golang 的 SSR 订阅解析工具，并提供简单的配置有效性校验，并保存到 Redis 或者本地。

## 安装和编译


需要有 golang 的开发环境，然后简单的执行

```sh
# sudo pacman -S go
make
```

即可有对应的二进制可执行文件。如果你觉得还是太麻烦，有对应的 docker 镜像可供使用或者本地编译，参考 `docker-compose.yaml` 这个文件即可。

## 配置

ssr-subscriber 支持 文件、网络等订阅形式，可以参考代码库中的 config-example.yml 文件进行配置。示例配置文件如下：

```yaml
url:
  - <your-ssr-subscriber-url>

file:
  - <you-local-ssr-subscriber-file-base64-encoded>

check:
  timeout: 3s
  not: 免费|普通|回国|过期|剩余

proxy: localhost:1088
redis: localhost:6379
bind: 0.0.0.0:8080
interval: 30 #minutes
```

其中，`check` 字段中 `timeout` 为检查对应 ssr 节点的主机和端口超时时间，`not` 字段为过滤和忽略 `Remarks` 中包含对应的字符串（支持正则表达式）。`proxy` 字段支持 socks5 的代理地址，~~强烈建议使用代理请求数据~~。

同时，支持 `redis` 将配置存储以及缓存，非常建议这样子做以加快下次启动以及或许的速度。`bind` 配置为绑定 http 的地址以及端口，作为配置文件以及状态的输出。


## 整合（已废弃）

请使用 interval 参数，具体的在上面的配置字段中说明，单位是分钟。

<del>
建议配合 crontab 使用，定期抓取和更新订阅链接的节点信息，例如：

```
30 5 * * * rm -f $HOME/configs/*.json && ssr-subscriber -c $HOME/ssr-subscriber.yaml
```
</del>

## @TODO

- 代码优化和精简
- 支持其他格式的输出
