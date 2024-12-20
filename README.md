<!--
 * @Author: FunctionSir
 * @License: AGPLv3
 * @Date: 2024-12-07 22:09:43
 * @LastEditTime: 2024-12-20 22:09:38
 * @LastEditors: FunctionSir
 * @Description: -
 * @FilePath: /gopher-paste/README.md
-->
# gopher-paste

**!!! WITH ABSOLUTELY NO WARRANTING !!!**  
**!!! THIS PROJECT IS NOW IN ALPHA STAGE !!!**  
**Current version: 0.0.1 (SatenRuiko).**  
A special pastebin, written in Go.  

## How To Use

### Paste something

URL: <http://example.org/>  
Just send a POST request. Keys:  
expiration: Life time of the paste, 0 = inf, x = x hours.  
content-type: Content-type you want to use when someone visit it.  
encoded: Server need to de-base64 your data. "true" or "false".  
data: The actual data.  
A URL to your paste (line 1), and a UUID as your token to manage the paste (line 2) will be returned.  

#### Paste Limits

It's related to "expiration".  
2161 ~ inf: 64KiB max.  
721 ~ 2160: 128KiB max.  
169 ~ 720: 256KiB max.  
73 ~ 168: 512KiB max.  
25 ~ 72: 1MiB max.  
1 ~ 24: 2MiB max.  
The data len is the len of what you put into the "data" field.  

### Get something

URL: <https://example.org/[id]>  
Just send a GET request. Keys:  
ct: Content-type you want. Will overwrite the paster's one.  

### Delete something

URL: <https://example.org/[id]>  
Just send a DELETE request. Keys:  
token: The token gened before used to manage the paste.  

### Modify something

URL: <https://example.org/[id]>  
Just send a PUT request. Keys:  
token: The token gened before used to manage the paste.  
content-type: Content-type you want to use when someone visit it.  
encoded: Server need to de-base64 your data. "true" or "false".  
data: The actual data.  

#### Modify Limits

1. Can't modify the expiration time.  
2. Data size limits are exactly the same as "creation".  
P.S. The "last modify" will set to the time you modified it.  

## How to deploy

### To build

``` bash
git clone https://github.com/FunctionSir/gopher-paste.git
cd gopher-paste
go build -ldflags '-s -w' *.go
mv main gopherpaste
```

### To run

``` bash
./gopherpaste [configfile]
```

If you didn't specify the config file, every thing will use the default value. It's good to debug but bad to production deployments.

### Config file

``` ini
# This "[options]" is necessary.
[options]
# Addr that the server listen to.
Addr = 0.0.0.0:6450
# The home page of the service.
HomePage = /some/path/index.html
# Where the pastes stored.
PastesDir = /some/path/
# In outputs after created a new post.
BaseURL = http://example.org/
```

#### Something about the "base url"

``` go
// ...... //
c.String(http.StatusOK, BaseURL+id+"\n"+token)
// ...... //
if sec.HasKey("BaseURL") {
    BaseURL = sec.Key("BaseURL").String()
    if BaseURL[len(BaseURL)-1] != '/' {
        BaseURL = BaseURL + "/"
    }
}
// ...... //
```

## Some global consts and vars

``` go
const (
    VER               string = "0.0.1"
    CODENAME          string = "SatenRuiko"
    CONTEXT_TYPE_HTML string = "text/html"
    ID_CHARSET        string = "0123456789"
    META_TOKEN_POS    int    = 0
    META_CT_POS       int    = 1
    META_EXP_POS      int    = 2
    META_LM_POS       int    = 3
)

var (
    Addr               string = "127.0.0.1:6450"
    HomePage           string = "index.html"
    PastesDir          string = "pastes"
    IdLen              int    = 8
    DefaultExpiration  int    = 24
    DefaultContentType string = "text/plain"
    BaseURL            string = "http://127.0.0.1:6450/"
    CleanGap           int    = 900
)
```
