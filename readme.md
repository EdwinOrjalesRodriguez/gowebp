# gowebp

gowebp is a tiny cli tool used to create webp images from jpegs and png files
the tool will create webp images alongside the jpeg and png files it finds in a target
directory with the same base file name

- requires libwebp to be installed

Usage:
```
  -append string
        append string to the end of file name
  -d string
        the directory to crawl
  -dry-run
        whether to handle this as a dry run and only print target files
  -min-size string
        smallest file size that will have a webp image created (default "10KB")
  -prepend string
        prepend string to the beginning of file name
  -q uint
        the quality for the webp images
  -r
  		replace existing webp files
  -w int
        the number of worker routines to spawn. Defaults to number of CPUs
  -chown string
  		assign unix owner to created webp files
  -chgrp string
  		assign unix group to created webp files
  -chmod string/int
  		assign unix permissions to created webp files (string or int will both work)
```


# dependencies 
webp libraries
```
yum install libwebp -y
yum install libwebp-tools -y
```

# compiling Unix executable in GoLand (Windows)
```
 env GOOS=linux go build -o gowebp
```