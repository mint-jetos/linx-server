## linx-server cleanup

Manually clean up expired files

```
linx-server cleanup [flags]
```

### Options

```
  -c, --config string         Path to the config file (default "$HOME/.config/linx-server/config.toml")
      --files-path string     Path to files directory (default "data/files")
  -h, --help                  help for cleanup
      --meta-path string      Path to metadata directory (default "data/meta")
      --no-logs               Disable logging of deleted files
      --s3-bucket string      S3 bucket to use for files and metadata
      --s3-endpoint string    S3 endpoint
      --s3-force-path-style   Force path-style addressing for S3 (e.g. https://s3.amazonaws.com/linx/example.txt)
      --s3-region string      S3 region
```

### SEE ALSO

* [linx-server](linx-server.md)

