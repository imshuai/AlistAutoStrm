# AlistAutoStrm  
Auto generate .strm file for EMBY or Jellyfin server use Alist API  
使用 Alist API自动生成 .strm 文件用于 EMBY 或 Jellyfin 服务器  

## 最近更新  
 - 2025-01-03 v1.2.5增加同时下载额外文件功能，例如：.jpg、.nfo、.srt、.ass等, 在配置文件中使用`alt-exts`配置一个字符串数组，例如：`[".jpg",".nfo",".srt",".ass"]`
## Usage 使用方法  
```
D:\AlistAutoStrm\bin>ass_windows_amd64.exe --help                 
NAME:
   AlistAutoStrm - Auto generate .strm file for EMBY or Jellyfin server use Alist API

USAGE:
   AlistAutoStrm [global options] command [command options]

VERSION:
   1.2.5

DESCRIPTION:
   Auto generate .strm file for EMBY or Jellyfin server use Alist API

COMMANDS:
   fresh-all        (TODO) generate all strm files from alist server, whatever the file has been generated or not
   update           update strm file with choosed mode
   update-database  clean database and get all local strm files stored in database
   check            check if strm file is valid
   version          show version
   help, h          Shows a list of commands or help for one command

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
      "alt-exts":[".jpg",".nfo",".srt",".ass"],
      "endpoints": [
            {
                  "base-url": "https://alist.cn",
                  "token": "test-token",
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
alt-exts:
  - ".jpg"
  - ".nfo"
  - ".srt"
  - ".ass"
endpoints:
  - base-url: "https://alist.cn"
    token: "test-token"
    username: "test"
    password: "test"
    inscure-tls-verify: false
    dirs:
      - local-directory: "data/movie"
        remote-directories:
          - "/path/to/movie"
          - "/path/to/movie2"
        create-sub-directory: true
        not-recursive: false
        force-refresh: false
        disabled: false
    max-connections: 10
```
### Tips 提示  
* 初次使用时，请先使用 `update-database` 命令，将所有本地目录中的 .strm 文件记录到数据库中，以便后续更新时使用。后续只需使用 `update` 命令更新。
* `update`命令支持两种模式：`local`或`remote`，默认为`local`，意为当远程文件路径与本地strm内容不一致时，保持本地strm文件不变；`remote`意为当远程文件路径与本地strm内容不一致时，更新本地strm文件内容，并更新数据库。
* `update`命令还接受一个`--no-incremental-update`参数，意为不进行增量更新，程序会进入每一个远程文件夹获取文件列表，并根据规则生成strm文件及下载额外的文件，如图片、字幕等，默认为`false`。
* 配置文件中，全局 `create-sub-directory` 与各自目录的`create-sub-directory`取逻辑或关系，举例说明:  
  * 当全局 `create-sub-directory` 设置为 `false` 时, 各自目录的 `create-sub-directory` 设置为 `true` 时, 最终结果为 `true`;
  * 当全局 `create-sub-directory` 设置为 `true` 时, 各自目录的 `create-sub-directory` 设置为 `false` 时, 最终结果为 `true`;
* `force-refresh` 配置项控制是否每次请求时强制刷新远端目录，默认为 `false`，注意: 设置为 `true` 时可能会导致一些问题。  
* `not-recursive` 配置项控制是否不要递归生成 .strm 文件到子目录中，默认为 `false`。
## Author  
[@imshuai](https://github.com/imshuai)  
## License  
MIT License  
## Thanks  
[alist](https://github.com/alist-org/alist)