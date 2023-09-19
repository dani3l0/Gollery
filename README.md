
# Gollery

Extremely simple self-hosted gallery app written in Go.

![Preview](preview.png)


## 🏃 How to run

⬇️ Just download and run latest executable binary file from [Releases](https://github.com/dani3l0/Gollery/releases)!
It works out of the box.

📂 Gollery creates two folders on first run:

- `cache` for generated thumbnails
- `images` for your image gallery (here you can place your files)

🔗 You can place symlinked directories at the root of `images` if you don't want to move all your files there.


## 🔧 Configuration

Configuration can be done via **environment variables**.

| Name                | Description                                             | Default value  | Example values                               |
|---------------------|---------------------------------------------------------|----------------|----------------------------------------------|
| `GOLLERY_LISTEN`    | Sets address and port app will listen to                | `:8080`        | `127.0.0.1:8080` `0.0.0.0:3000` `:5000`      |
| `GOLLERY_SECRET`    | A secret phrase you use to log in                       | `mysecret2137` | `username:password` `mywallpapershere` `aruhgjf9r7s693h2` |
| `GOLLERY_THUMBSIZE` | Defines the thumbnail size in pixels: using large values (>480) will slow down the Web UI.<br>**Does not resize existing thumbnails.** | `400`          | `192` `320` `480` `640` |

Then run like this:
```
GOLLERY_SECRET="i-love-my-cat" ./gollery
```


## 🔨 Build
```
# Download dependencies
make deps

# Build binary for your current platform
make build

# Build binaries for all platforms
make build-all
```

- Built binaries will be available under `dist` directory.

- If you don't have `make`, you can just copy-paste commands from `Makefile`.


## 🤔 Final words?

Well, that's my first project in `Go`. It is not recommended to use it for production and extremely large image libraries (but still you can)

**However, I did my best so app is pretty usable :)**
