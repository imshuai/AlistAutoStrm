# AlistAutoStrm  
Auto generate .strm file for EMBY or Jellyfin server use Alist API  
使用 Alist API自动生成 .strm 文件用于 EMBY 或 Jellyfin 服务器  
## Usage 使用方法  
```
D:\AlistAutoStrm\bin>ass_windows_amd64.exe --help                 
NAME:
   AlistAutoStrm - Auto generate .strm file for EMBY or Jellyfin server use Alist API

USAGE:
   AlistAutoStrm [global options] command [command options] [arguments...]

VERSION:
   1.1.1

DESCRIPTION:
   Auto generate .strm file for EMBY or Jellyfin server use Alist API

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config FILE, -c FILE  Load configuration from FILE (default: "config.json") [%ALIST_AUTO_STRM_CONFIG%]
   --help, -h              show help
   --version, -v           print the version

D:\AlistAutoStrm\bin>ass_windows_amd64.exe --config config.json
```
## Config file example 配置文件示例  
### JSON format JSON格式
```json
{
      "database": "strm.db",
      "loglevel": "info",
      "timeout": 30,
      "create-sub-directory": true,
      "exts":[".mp4",".mkv",".avi",".rmvb"],
      "endpoints": [
            {
                  "base-url": "https://alist.cn",
                  "username": "test",
                  "password": "test",
                  "inscure-tls-verify": false,
                  "dirs": [
                        {
                              "local-directory": "data/movie",
                              "remote-directories": [
                                    "/path/to/movie",
                                    "/path/to/movie2"
                              ],
                              "create-sub-directory":true,
                              "not-recursive":false,
                              "force-refresh":false,
                              "disabled":true
                        }
                  ],
                  "max-connections": 10
            }
      ]
}
```
### YAML format YAML格式  
```yaml
database: "strm.db"
loglevel: "info"
timeout: 30
create-sub-directory: true
exts:
  - ".mp4"
  - ".mkv"
  - ".avi"
  - ".rmvb"
endpoints:
  - base-url: "https://alist.cn"
    username: "test"
    password: "test"
    inscure-tls-verify: false
    dirs:
      - local-directory: "data/movie"
        remote-directories:
          - "/path/to/movie"
          - "/path/to/movie2"
        create-sub-directory: true
```
### Tips 提示  
#### `create-sub-directory`
global `create-sub-directory` option is high proriority than per dir `create-sub-directory` option  
全局 `create-sub-directory` 优先级高于各自目录的`create-sub-directory`设置  
for example 举例说明:  
* global `create-sub-directory` is `true`, per dir `create-sub-directory` no matter what is, final result is `true`;  
  当全局 `create-sub-directory` 设置为 `true` 时, 各自目录的 `create-sub-directory` 设置无关紧要, 最终结果为 `true`;  
* global `create-sub-directory` is false, per dir `create-sub-directory` is true, final result is true;  
  当全局 `create-sub-directory` 设置为 `false` 时, 各自目录的 `create-sub-directory` 设置为 `true` 时, 最终结果为 `true`;  
#### `force-refresh`  
`force-refresh` option is control whether to refresh the remote directory per request or not, default is `false`. Attention: set it to `true` may cause some issue.  
`force-refresh` 选项控制是否每次请求时强制刷新远端目录，默认为 `false`，注意: 设置为 `true` 时可能会导致一些问题。  
#### `not-recursive`
`not-recursive` option is control whether to generate .strm file recursively or not, default is `false`.  
`not-recursive` 选项控制是否递归生成 .strm 文件，默认为 `false`。
## Author  
[@imshuai](https://github.com/imshuai)  
## License  
MIT License  
## Thanks  
[alist](https://github.com/alist-org/alist)